package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	model "tragedylooper/internal/game/proto/v1"

	"tragedylooper/internal/llm"
	"tragedylooper/internal/logger"
	"tragedylooper/internal/server"

	"go.uber.org/zap"
)

func main() {
	logger := logger.New()
	defer func() {
		_ = logger.Sync() // Flushes buffer, important for production
	}()

	// 1. 加载游戏剧本和资源
	// 在实际应用中，这些将从文件或数据库加载。
	// 为简单起见，我们使用占位符。
	scripts := map[string]*model.Script{
		"basic_script": {
			Id:          0,
			Name:        "Basic Tragedy",
			Description: "A simple script for testing.",
			LoopCount:   3,
			DaysPerLoop: 7,
			Characters: []*model.CharacterConfig{
				{Id: 0, Name: "boy_student", InitialLocation: model.LocationType_LOCATION_TYPE_SCHOOL, HiddenRole: model.RoleType_ROLE_TYPE_INNOCENT},
				{Id: 0, Name: "girl_student", InitialLocation: model.LocationType_LOCATION_TYPE_SCHOOL, HiddenRole: model.RoleType_ROLE_TYPE_INNOCENT},
				{Id: 0, Name: "serial_killer", InitialLocation: model.LocationType_LOCATION_TYPE_CITY, HiddenRole: model.RoleType_ROLE_TYPE_KILLER},
			},
			Tragedies: []*model.TragedyCondition{
				{
					TragedyType: model.TragedyType_TRAGEDY_TYPE_MURDER,
					Day:         3, // 谋杀可能发生在第 3 天
					Conditions: []*model.Condition{
						{CharacterId: 0, Location: model.LocationType_LOCATION_TYPE_SCHOOL, IsAlone: true},
					},
				},
			},
		},
	}

	// 初始化 LLM 客户端（例如，OpenAI，Google Gemini）
	// 这通常涉及从环境变量获取 API 密钥。
	llmClient := llm.NewMockLLMClient() // 使用模拟客户端进行演示
	// llmClient := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY")) // 用于实际的 OpenAI 集成

	// 2. 初始化游戏服务器
	gameServer := server.NewServer(scripts, llmClient, logger)

	// Create a new ServeMux to apply middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", gameServer.HandleWebSocket)
	mux.HandleFunc("/create_room", gameServer.HandleCreateRoom)
	mux.HandleFunc("/join_room", gameServer.HandleJoinRoom)
	mux.HandleFunc("/list_rooms", gameServer.HandleListRooms)

	// Apply the logging middleware
	loggedMux := gameServer.LoggingMiddleware(mux)

	port := ":8080"
	logger.Info("Server starting on port " + port)

	// In a goroutine, start the HTTP server
	go func() {
		if err := http.ListenAndServe(port, loggedMux); err != nil {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// 优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // 阻塞直到收到信号
	logger.Info("Shutting down server...")
	gameServer.Shutdown() // 执行任何清理
	logger.Info("Server gracefully stopped.")
}
