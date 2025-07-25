
# Tragedy Looper

A Go implementation of the Tragedy Looper board game.

## Description

This project is a Go-based server for the Tragedy Looper game, allowing players to connect and play through a client-server architecture. It also includes an AI opponent powered by Large Language Models (LLMs).

## Getting Started

### Prerequisites

- Go 1.x

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/your-username/tragedyLooper.git
   cd tragedyLooper
   ```

2. Tidy up the dependencies:
   ```sh
   go mod tidy
   ```

### Running the application

```sh
go run ./cmd/tragedylooper
```

## Project Structure

```
. (tragedyLooper)
|-- cmd/tragedylooper/main.go   # Main application entry point
|-- internal/                     # Private application and library code
|-- pkg/
|   |-- game/                     # Core game logic and models
|   |-- llm/                      # LLM client and integration
|   `-- server/                   # Server and client connection handling
|-- go.mod
`-- README.md
```
