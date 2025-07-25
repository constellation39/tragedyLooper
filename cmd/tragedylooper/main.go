package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tragedylooper/pkg/game/model"
	"tragedylooper/pkg/llm"
	"tragedylooper/pkg/server"
)

func main() {
	// 1. 加载游戏剧本和资源
	// 在实际应用中，这些将从文件或数据库加载。
	// 为简单起见，我们使用占位符。
	scripts := map[string]model.Script{
		"basic_script": {
			ID:          "basic_script",
			Name:        "Basic Tragedy",
			Description: "A simple script for testing.",
			LoopCount:   3,
			DaysPerLoop: 7,
			Characters: []model.CharacterConfig{
				{CharacterID: "boy_student", InitialLocation: model.LocationSchool, HiddenRole: model.RoleInnocent},
				{CharacterID: "girl_student", InitialLocation: model.LocationSchool, HiddenRole: model.RoleInnocent},
				{CharacterID: "serial_killer", InitialLocation: model.LocationCity, HiddenRole: model.RoleKiller, IsCulpritFor: model.TragedyMurder},
			},
			Tragedies: []model.TragedyCondition{
				{
					TragedyType: model.TragedyMurder,
					Day:         3, // 谋杀可能发生在第 3 天
					CulpritID:   "serial_killer",
					Conditions: []model.Condition{
						{CharacterID: "boy_student", Location: model.LocationSchool, IsAlone: true},
					},
					TargetRule: model.TargetRuleSpecificCharacter,
				},
			},
		},
	}

	// 初始化 LLM 客户端（例如，OpenAI，Google Gemini）
	// 这通常涉及从环境变量获取 API 密钥。
	llmClient := llm.NewMockLLMClient() // 使用模拟客户端进行演示
	// llmClient := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY")) // 用于实际的 OpenAI 集成

	// 2. 初始化游戏服务器
	gameServer := server.NewServer(scripts, llmClient)

	// 设置 WebSocket 连接的 HTTP 服务器
	http.HandleFunc("/ws", gameServer.HandleWebSocket)
	http.HandleFunc("/create_room", gameServer.HandleCreateRoom)
	http.HandleFunc("/join_room", gameServer.HandleJoinRoom)
	http.HandleFunc("/list_rooms", gameServer.HandleListRooms)

	port := ":8080"
	log.Printf("Server starting on port %s", port)

	// 在协程中启动 HTTP 服务器
	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// 优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // 阻塞直到收到信号
	log.Println("Shutting down server...")
	gameServer.Shutdown() // 执行任何清理
	log.Println("Server gracefully stopped.")
}
