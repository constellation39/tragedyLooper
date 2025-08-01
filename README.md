# Tragedy Looper

> **Note:** This project is currently under active development. The API and features are subject to change.

A Go implementation of the Tragedy Looper board game. This project is a Go-based server for the Tragedy Looper game, allowing players to connect and play through a client-server architecture. It also includes an AI opponent powered by Large Language Models (LLMs).

## Getting Started

### Prerequisites

- Go 1.x
- Git
- Buf
- protoc-gen-go

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/constellation39/tragedyLooper.git
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

## Project Structure

A brief overview of the key directories in this project.

```
. (tragedyLooper)
|-- cmd/tragedylooper/main.go   # Main application entry point.
|-- data/                         # All game data (cards, characters, scripts).
|-- cue.mod/                      # CUE definitions for game data validation.
|-- internal/                     # Private application code, not intended for import by other projects.
|   |-- game/                     # The core game engine, state management, and logic.
|   |-- llm/                      # Handles all communication with the external Large Language Model for AI.
|   `-- server/                   # The gRPC/HTTP server that handles client connections and communication.
|-- pkg/                          # Public library code, safe for other projects to import (e.g., generated protobuf code).
|-- proto/                        # The "source of truth" for data structures, defined in .proto files.
|-- tools/                        # Helper scripts and small applications for development.
|-- go.mod                        # Go module definitions.
|-- go.sum                        # Go module checksums.
|-- Makefile                      # Contains all build, test, and run commands.
`-- README.md                     # You are here.
```

## Architecture

The project's architecture is modern, robust, and well-suited for a complex, stateful game like Tragedy Looper. It is driven by several key principles, resulting in a clear and maintainable codebase.

### High-Level Design

1.  **Configuration Driven:** The entire game's content—cards, characters, scripts, and incidents—is defined in external JSON files (`data/`). This decouples the game logic from the game data, allowing for easy updates and modifications to game content without changing any code. The use of CUE (`cue.mod/`) ensures this data remains valid and consistent.

2.  **Event-Driven Core Engine:** The game's heart (`internal/game/engine/`) operates on an event-driven model. Player actions, AI decisions, and internal game mechanics do not directly modify the game state. Instead, they generate **events** (e.g., `CardPlayedEvent`). These events are processed by dedicated handlers (`internal/game/engine/eventhandler/`) which in turn trigger **effects** (`internal/game/engine/effecthandler/`). This approach elegantly handles complex chain reactions and makes the game logic easy to extend and debug.

3.  **AI/LLM Integration:** A standout feature is the integration of a Large Language Model (LLM) for AI decision-making (`internal/llm/`). The system builds a context-aware prompt from the current game state (`PlayerView`), sends it to an external LLM, and parses the response back into a valid game action. This allows for sophisticated and less predictable AI behavior.

4.  **Contract-First with Protobuf:** All data models are formally defined using Protobuf (`proto/v1/`). This creates a strict "contract" for data structures used throughout the application, from the game engine to the server API, ensuring type safety and consistency.

### System Data Flow

The following diagram illustrates the interaction between these core components. It shows the complete loop from data loading and player interaction to AI decision-making and state updates.

```mermaid
graph TD
    subgraph " "
        direction LR
        subgraph "Client"
            direction TB
            UI[Client]
        end

        subgraph "Go Application"
            direction TB
            
            subgraph "Configuration Layer"
                direction RL
                JSON[data/*.json] -->|Load & Parse| Loader(internal/game/loader)
                Schema[cue.mod/**/*.cue] -->|Validate| Loader
                Proto(proto/v1/*.proto) -->|Generate| PB(pkg/proto/v1/*.pb.go)
                Loader -->|Instantiate| PB
            end
            
            Server(internal/server/server.go)

            subgraph "Game Engine"
                direction TB
                Engine(engine.go)
                PhaseMgr(phase_manager.go)
                EventMgr -- Dispatches to --> EventHandlers(eventhandler/*)
                EffectHandlers -- Modify --> GameState
            end

            subgraph "AI/LLM Driver"
                direction TB
                GameAI(engine/ai.go)
                LLM_Client(internal/llm/client.go)
                PromptBuilder(internal/llm/prompt_builder.go)
                ActionGenerator(internal/llm/action_generator.go)
                ExternalLLM[External LLM API]

                GameAI -->|Builds Prompt| PromptBuilder
                PromptBuilder -->|Sends Request| LLM_Client
                LLM_Client -->|Calls| ExternalLLM
                ExternalLLM -->|Returns Response| LLM_Client
                LLM_Client -->|Parses Action| ActionGenerator
            end

            UI -- Player Action --> Server
            Server -- Forwards Action --> Engine
            Engine -- Triggers --> EventMgr
            
            Engine -- Requests AI Action --> GameAI
            ActionGenerator -- Returns Action --> GameAI
            GameAI -- Submits Action --> Engine

            Engine -- Generates --> View(view.go)
            View -- Player-specific State --> Server
            Server -- Sends Update --> UI
        end
    end

    style GameState fill:#lightgrey,stroke:#333,stroke-width:2px
    style JSON fill:#f9f,stroke:#333,stroke-width:2px
```

### Lifecycle of a Player Action

To understand how the engine works in detail, let's trace the complete lifecycle of a single player action. The project is built around an event-driven game engine that processes player actions and manages game state through a series of immutable events and chain reactions.

```mermaid
graph TD
    subgraph "Phase 1: Setup & Data Loading"
        direction LR
        A1["proto/v1/*.proto<br><b>(Data Structure Contract)</b>"] -- "1. Compile with protoc" --> A2["internal/game/proto/v1/*.pb.go<br><b>(Generated Go Code)</b>"]
        B1["data/*.json<br><b>(Game Scripts, Cards, Characters)</b>"] -- "2. On App Startup" --> B2["internal/game/loader/loader.go<br><b>(Config Loader)</b>"]
        B2 -- "3. Parse JSON to Go structs" --> A2
    end

    subgraph "Phase 2: Application Runtime Environment"
        C1["cmd/tragedylooper/main.go<br><b>(Entry Point)</b>"] -- "4. Initialize & Start" --> C2["internal/server/server.go<br><b>(gRPC/HTTP Server)</b>"]
        C2 -- "5. Create & Hold" --> C3["<b>Game Engine Instance<br>(internal/game/engine)</b>"]
        B2 -- "6. Inject Loaded Config Data" --> C3
    end

    subgraph "Phase 3: Engine Core Loop - Lifecycle of a Single Action"
        D1["Player Client<br><b>(External Player)</b>"] -- "7. Send Action<br><i>(e.g., PlayCardPayload)</i>" --> C2
        C2 -- "8. Validate & Forward to Engine" --> E1{"<b>Engine Coordinator<br>(engine.go)</b>"}

        subgraph "Internal Engine Components"
            direction LR
            F1["<b>GameState</b><br><i>(The single source of truth)</i>"]
            F2["<b>Event Queue</b><br><i>(FIFO)</i>"]
            
            subgraph "Logic Processors"
                direction TB
                G1["actions.go"]
                G2["effects.go"]
                G3["triggers.go"]
                G4["event_handlers.go"]
                G5["phase_handlers.go"]
                G6["view.go"]
            end
        end

        E1 -- "9. Dispatch to<br><b>Action Handler</b>" --> G1
        G1 -- "10. [Read-Only] Validate action legality" --> F1
        G1 -- "11. Create initial event<br><i>(e.g., CardPlayedEvent)</i>" --> F2

        E1 -- "12. Dequeue event for processing" --> F2
        E1 -- "13. Find & invoke<br><b>Effect Handler</b>" --> G2
        G2 -- "14. <b>[Write] Modify state</b><br><i>(e.g., Paranoia +1)</i>" --> F1

        F1 -- "15. State changed!<br><b>Trigger Chain Reaction Check</b>" --> H1{"<b>Post-Change<br>Chain Reaction</b>"}

        H1 -- "16a. Check all<br><b>Triggers</b>" --> G3
        G3 -- "17. [Read-Only] Evaluate trigger conditions" --> F1
        G3 -- "18. If met, create new event" --> F2

        H1 -- "16b. Notify all<br><b>Event Handlers</b>" --> G4
        G4 -- "19. [Read-Only] Respond to specific event" --> F1
        G4 -- "20. May create new event" --> F2

        F2 -- "21. <b>Loop!</b><br>If queue is not empty, return to step 12" --> E1

        E1 -- "22. Queue empty, call<br><b>Phase Handler</b>" --> G5
        G5 -- "23. [Write] May advance day/phase" --> F1

        E1 -- "24. Generate new view for player<br>call <b>View Generator</b>" --> G6
        G6 -- "25. [Read-Only] Extract player-visible info" --> F1
        G6 -- "26. Create PlayerView object" --> I1["<b>PlayerView</b>"]

        I1 -- "27. Return to server" --> C2
        C2 -- "28. Send to client" --> D1
    end
```