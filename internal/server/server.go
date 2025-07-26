package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"tragedylooper/internal/game/engine"
	"tragedylooper/internal/game/proto/model"
	"tragedylooper/internal/llm"
	"tragedylooper/internal/logger"
)

// Server 管理多个游戏房间和 WebSocket 连接。
type Server struct {
	upgrader websocket.Upgrader
	rooms    map[string]*Room // 游戏 ID 到 Room 的映射
	mu       sync.RWMutex     // rooms 映射的互斥锁
	// 用于发出服务器关闭信号的通道
	shutdownChan chan struct{}
	// 可用游戏剧本的映射
	scripts map[string]*model.Script
	// LLM 客户端用于 AI 玩家
	llmClient llm.Client
	logger    *zap.Logger
}

// NewServer 创建一个新的游戏服务器实例。
func NewServer(scripts map[string]*model.Script, llmClient llm.Client, logger *zap.Logger) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true // 允许所有来源进行开发，在生产环境中限制
			},
		},
		rooms:        make(map[string]*Room),
		shutdownChan: make(chan struct{}),
		scripts:      scripts,
		llmClient:    llmClient,
		logger:       logger,
	}
}

// Shutdown 优雅地关闭服务器和所有活跃房间。
func (s *Server) Shutdown() {
	close(s.shutdownChan)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, room := range s.rooms {
		room.Stop() // 发送信号给每个房间停止其游戏循环
	}
	time.Sleep(2 * time.Second) // 给予房间一些时间关闭

}

// LoggingMiddleware creates a new logger with a request_id and adds it to the context.
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctxLogger := s.logger.With(zap.String("request_id", requestID))
		ctx := logger.WithContext(r.Context(), ctxLogger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandleWebSocket 处理传入的 WebSocket 连接。
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ctxLogger := logger.FromContext(r.Context())

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctxLogger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		ctxLogger.Warn("WebSocket connection: Missing player_id query parameter.")
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Error: player_id required.")); err != nil {
			ctxLogger.Error("Error writing message", zap.Error(err))
		}
		return
	}

	clientLogger := ctxLogger.With(zap.String("playerID", playerID))
	clientLogger.Info("Player connected via WebSocket.")

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		playerID: playerID,
		room:     nil, // Set when joining a room
		logger:   clientLogger,
	}

	go client.writePump()
	client.readPump(s)
}

// HandleCreateRoom 处理创建新游戏房间的请求。
func (s *Server) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	ctxLogger := logger.FromContext(r.Context())
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ScriptID   string           `json:"script_id"`
		PlayerID   string           `json:"player_id"`
		PlayerName string           `json:"player_name"`
		PlayerRole model.PlayerRole `json:"player_role"`
		IsLLM      bool             `json:"is_llm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	script, ok := s.scripts[req.ScriptID]
	if !ok {
		http.Error(w, "Script not found", http.StatusNotFound)
		return
	}

	gameID := generateUniqueGameID() // 实现生成唯一 ID 的函数
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[gameID]; exists {
		http.Error(w, "Game ID already exists, try again", http.StatusConflict)
		return
	}

	players := make(map[string]*model.Player)
	players[req.PlayerID] = &model.Player{
		Id:                 req.PlayerID,
		Name:               req.PlayerName,
		Role:               req.PlayerRole,
		IsLlm:              req.IsLLM,
		Hand:               make([]*model.Card, 0), // 卡牌将由游戏引擎处理
		DeductionKnowledge: nil,
		LlmSessionId:       "",
	}

	gameEngine := engine.NewGameEngine(gameID, ctxLogger, script, players, s.llmClient)
	room := NewRoom(gameID, gameEngine, ctxLogger)
	s.rooms[gameID] = room

	room.Start() // 启动此房间的游戏引擎循环

	ctxLogger.Info("Room created", zap.String("gameID", gameID), zap.String("playerID", req.PlayerID), zap.String("scriptID", req.ScriptID))
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"game_id": gameID}); err != nil {
		ctxLogger.Error("Error encoding response", zap.Error(err))
	}
}

// HandleJoinRoom 处理加入现有游戏房间的请求。
func (s *Server) HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	ctxLogger := logger.FromContext(r.Context())
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GameID     string           `json:"game_id"`
		PlayerID   string           `json:"player_id"`
		PlayerName string           `json:"player_name"`
		PlayerRole model.PlayerRole `json:"player_role"`
		IsLLM      bool             `json:"is_llm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	room, ok := s.rooms[req.GameID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	room.gameEngine.GameState.Players[req.PlayerID] = &model.Player{
		Id:         req.PlayerID,
		Name:       req.PlayerName,
		Role:       req.PlayerRole,
		IsLlm:      req.IsLLM,
		Hand:       &model.CardList{}, // Cards will be handled by the model engine
		Deductions: make(map[string]string),
	}

	ctxLogger.Info("Player joined room", zap.String("playerID", req.PlayerID), zap.String("gameID", req.GameID), zap.String("role", req.PlayerRole.String()))
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Joined room successfully"}); err != nil {
		ctxLogger.Error("Error encoding response", zap.Error(err))
	}
}

// HandleListRooms 处理列出可用游戏房间的请求。
func (s *Server) HandleListRooms(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var roomList []map[string]interface{}
	for id, room := range s.rooms {
		// 只列出未满或未开始的房间
		if room.gameEngine.GameState.CurrentPhase == model.GamePhase_GAME_PHASE_MORNING { // 示例条件
			roomList = append(roomList, map[string]interface{}{
				"id":            id,
				"script_name":   room.gameEngine.GameState.Script.Name,
				"players_count": len(room.gameEngine.GameState.Players),
				"current_phase": room.gameEngine.GameState.CurrentPhase,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roomList); err != nil {
		s.logger.Error("Error encoding room list", zap.Error(err))
	}
}

// Client 表示单个 WebSocket 连接。
type Client struct {
	conn     *websocket.Conn
	send     chan []byte // 用于传出消息的带缓冲通道
	playerID string
	room     *Room // 此客户端所属的房间
	logger   *zap.Logger
}

// readPump 从 WebSocket 连接中抽取消息到房间。
func (c *Client) readPump(_ *Server) {
	defer func() {
		c.logger.Info("Player disconnected.")

		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error", zap.Error(err))
			}
			break
		}

		var action model.PlayerAction
		if err := json.Unmarshal(message, &action); err != nil {
			c.logger.Warn("Failed to parse incoming message as PlayerAction", zap.Error(err))
			continue
		}
		action.PlayerId = c.playerID

		if c.room != nil {
			c.room.gameEngine.SubmitPlayerAction(&action)
		} else {
			c.logger.Warn("Received action but not in a room.")
		}
	}
}

// writePump 将消息从房间抽取到 WebSocket 连接。
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			c.logger.Error("WebSocket write error", zap.Error(err))
			return
		}
	}
}

// Room 管理单个游戏实例及其连接的客户端。
type Room struct {
	GameID     string
	gameEngine *engine.GameEngine
	clients    map[string]*Client // 玩家 ID 到 Client 的映射
	mu         sync.RWMutex
	stopChan   chan struct{} // 用于发出房间停止信号的通道
	logger     *zap.Logger
}

// NewRoom 创建一个新的游戏房间。
func NewRoom(gameID string, ge *engine.GameEngine, logger *zap.Logger) *Room {
	return &Room{
		GameID:     gameID,
		gameEngine: ge,
		clients:    make(map[string]*Client),
		stopChan:   make(chan struct{}),
		logger:     logger.With(zap.String("gameID", gameID)), // Add gameID to all room logs
	}
}

// AddClient 将客户端添加到房间。
func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client.playerID] = client
	client.room = r // 设置客户端的房间引用
	r.logger.Info("Client added to room", zap.String("clientID", client.playerID), zap.String("roomID", r.GameID))
}

// RemoveClient 从房间中移除客户端。
func (r *Room) RemoveClient(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if client, ok := r.clients[playerID]; ok {
		close(client.send) // 关闭客户端的发送通道
		delete(r.clients, playerID)
		r.logger.Info("Client removed from room", zap.String("clientID", playerID), zap.String("roomID", r.GameID))
	}
}

// Start 启动房间的游戏引擎和事件广播。
func (r *Room) Start() {
	r.gameEngine.StartGameLoop()
	go r.broadcastGameEvents()
}

// Stop 停止房间的游戏引擎和广播。
func (r *Room) Stop() {
	r.gameEngine.StopGameLoop()
	close(r.stopChan)
	r.logger.Info("Room signaled to stop.", zap.String("roomID", r.GameID))
}

// broadcastGameEvents 监听游戏事件并将其广播给客户端。
func (r *Room) broadcastGameEvents() {
	eventChan := r.gameEngine.GetGameEvents()
	for {
		select {
		case <-r.stopChan:
			r.logger.Info("Event broadcaster stopped.", zap.String("roomID", r.GameID))
			return
		case event := <-eventChan:
			r.logger.Debug("Broadcasting event", zap.String("roomID", r.GameID), zap.String("eventType", event.Type.String()))
			r.mu.RLock()
			for playerID, client := range r.clients {
				playerView := r.gameEngine.GetPlayerView(playerID)
				msg, err := json.Marshal(playerView)
				if err != nil {
					r.logger.Error("Failed to marshal player view", zap.String("roomID", r.GameID), zap.String("playerID", playerID), zap.Error(err))
					continue
				}
				select {
				case client.send <- msg:
				default:
					r.logger.Warn("Client send channel full, dropping message.", zap.String("roomID", r.GameID), zap.String("playerID", playerID))
				}
			}
			r.mu.RUnlock()
		}
	}
}

// generateUniqueGameID 是实际 ID 生成函数的占位符。
func generateUniqueGameID() string {
	return fmt.Sprintf("game_%d", time.Now().UnixNano())
}
