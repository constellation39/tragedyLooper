package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tragedylooper/internal/game/loader"

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

	// 1. Load all game data.
	gameLoader := loader.NewJSONLoader("data")

	// Initialize LLM client (e.g., OpenAI, Google Gemini)
	// This would typically involve getting API keys from environment variables.
	llmClient := llm.NewMockLLMClient() // Using a mock client for demonstration
	// llmClient := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY")) // For actual OpenAI integration

	// 2. Initialize the game server
	gameServer := server.NewServer(gameLoader, llmClient, logger)

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
	server := &http.Server{
		Addr:    port,
		Handler: loggedMux,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
