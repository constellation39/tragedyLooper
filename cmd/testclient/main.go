package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tragedylooper/internal/game/proto/v1"

	"github.com/gorilla/websocket"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "http service address")
	playerID := flag.Int("player_id", 1, "player's ID")
	playerName := flag.String("player_name", "Test Player", "player's name")
	gameID := flag.String("game_id", "", "game ID to join")
	scriptID := flag.String("script_id", "first_steps", "script ID to create a new game")
	isLLM := flag.Bool("is_llm", false, "is this player an LLM")
	playerRole := flag.String("player_role", "PROTAGONIST", "player's role")
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws", RawQuery: fmt.Sprintf("player_id=%d", *playerID)}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// Create or join a room
	if *gameID == "" {
		createRoom(*addr, *scriptID, int32(*playerID), *playerName, *playerRole, *isLLM)
	} else {
		joinRoom(*addr, *gameID, int32(*playerID), *playerName, *playerRole, *isLLM)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter commands (e.g., 'play_card 1', 'use_ability 2', 'make_guess ...'):")

	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		parts := strings.SplitN(text, " ", 2)
		command := parts[0]
		var payload string
		if len(parts) > 1 {
			payload = parts[1]
		}

		var action *v1.PlayerActionPayload
		switch command {
		case "play_card":
			action = &v1.PlayerActionPayload{
				Payload: &v1.PlayerActionPayload_PlayCard{
					PlayCard: &v1.PlayCardPayload{CardId: payload},
				},
			}
		case "use_ability":
			action = &v1.PlayerActionPayload{
				Payload: &v1.PlayerActionPayload_UseAbility{
					UseAbility: &v1.UseAbilityPayload{AbilityId: payload},
				},
			}
		case "make_guess":
			// Example: make_guess CULPRIT:Alice,INCIDENT:Murder
			// You might need a more robust parser for a real client
			action = &v1.PlayerActionPayload{
				Payload: &v1.PlayerActionPayload_MakeGuess{
					MakeGuess: &v1.MakeGuessPayload{ /* Parse payload */ },
				},
			}
		default:
			log.Printf("Unknown command: %s", command)
			continue
		}

		data, err := json.Marshal(action)
		if err != nil {
			log.Println("marshal:", err)
			continue
		}

		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func createRoom(addr, scriptID string, playerID int32, playerName, playerRole string, isLLM bool) {
	role, ok := v1.PlayerRole_value[playerRole]
	if !ok {
		log.Fatalf("Invalid player role: %s", playerRole)
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"script_id":   scriptID,
		"player_id":   playerID,
		"player_name": playerName,
		"player_role": v1.PlayerRole(role),
		"is_llm":      isLLM,
	})

	resp, err := http.Post("http://"+addr+"/create_room", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		log.Fatal("create room failed:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("create room failed with status: %s", resp.Status)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	log.Printf("Room created with ID: %s", result["game_id"])
}

func joinRoom(addr, gameID string, playerID int32, playerName, playerRole string, isLLM bool) {
	role, ok := v1.PlayerRole_value[playerRole]
	if !ok {
		log.Fatalf("Invalid player role: %s", playerRole)
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"game_id":     gameID,
		"player_id":   playerID,
		"player_name": playerName,
		"player_role": v1.PlayerRole(role),
		"is_llm":      isLLM,
	})

	resp, err := http.Post("http://"+addr+"/join_room", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		log.Fatal("join room failed:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("join room failed with status: %s", resp.Status)
	}

	log.Printf("Joined room %s", gameID)
}
