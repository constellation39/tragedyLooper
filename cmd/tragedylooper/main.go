package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"tragedylooper/internal/game/loader"
	model "tragedylooper/internal/game/proto/v1"

	"tragedylooper/internal/llm"
	"tragedylooper/internal/logger"
	"tragedylooper/internal/server"

	"go.uber.org/zap"
)

func loadScripts(path string, logger *zap.Logger) map[string]*model.Script {
	scripts := make(map[string]*model.Script)
	files, err := filepath.Glob(filepath.Join(path, "*.json"))
	if err != nil {
		logger.Fatal("Failed to glob scripts", zap.Error(err))
	}

	for _, file := range files {
		script, err := loader.LoadScript(file)
		if err != nil {
			logger.Warn("Failed to load script", zap.String("file", file), zap.Error(err))
			continue
		}
		scriptID := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		scripts[scriptID] = script
		logger.Info("Loaded script", zap.String("id", scriptID))
	}
	return scripts
}

func main() {
	logger := logger.New()
	defer func() {
		_ = logger.Sync() // Flushes buffer, important for production
	}()

	// 1. Load game scripts and resources.
	scripts := loadScripts("data/scripts", logger)

	// Initialize LLM client (e.g., OpenAI, Google Gemini)
	// This would typically involve getting API keys from environment variables.
	llmClient := llm.NewMockLLMClient() // Using a mock client for demonstration
	// llmClient := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY")) // For actual OpenAI integration

	// 2. Initialize the game server
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
			log.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Block until a signal is received
	logger.Info("Shutting down server...")
	gameServer.Shutdown() // Perform any cleanup
	logger.Info("Server gracefully stopped.")
}
