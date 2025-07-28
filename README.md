# Tragedy Looper

A Go implementation of the Tragedy Looper board game. This project is a Go-based server for the Tragedy Looper game, allowing players to connect and play through a client-server architecture. It also includes an AI opponent powered by Large Language Models (LLMs).

## Description

This project is a Go-based server for the Tragedy Looper game. The core game logic is written in Go, and it uses protobuf for data serialization and gRPC for communication between the server and clients. The game also features an AI opponent that uses LLMs to make decisions.

## Getting Started

### Prerequisites

- Go 1.x
- Git
- Buf
- protoc-gen-go
- protoc-gen-jsonschema

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/your-username/tragedyLooper.git
    cd tragedyLooper
    ```

2.  **Install tools:**
    ```sh
    make install-tools
    ```

3.  **Generate protobuf files:**
    ```sh
    make proto
    ```

4.  **Tidy up the dependencies:**
    ```sh
    go mod tidy
    ```

## Usage

### Running the application

To run the application, use the following command:

```sh
make run
```

This will start the game server.

### Building the application

To build the application, use the following command:

```sh
make build
```

This will create a binary in the `bin` directory.

### Running tests

To run the tests, use the following command:

```sh
make test
```

### Linting the code

To lint the code, use the following command:

```sh
make lint
```

### Cleaning the project

To clean the project, use the following command:

```sh
make clean
```

This will remove the `bin` directory.

## Project Structure

```
. (tragedyLooper)
|-- cmd/tragedylooper/main.go   # Main application entry point
|-- data/                         # Game data and JSON schemas
|-- internal/                     # Private application and library code
|   |-- game/                     # Core game logic and models
|   |-- llm/                      # LLM client and integration
|   `-- server/                   # Server and client connection handling
|-- pkg/
|-- proto/                        # Protobuf definitions
|-- tools/                        # Helper scripts
|-- go.mod
|-- go.sum
|-- Makefile
`-- README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request