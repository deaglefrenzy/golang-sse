package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/deaglefrenzy/golang-sse/models"
)

var history []models.Chat

func main() {
	readerInput := bufio.NewReader(os.Stdin)

	// Ask username
	fmt.Print("Enter username: ")
	user, _ := readerInput.ReadString('\n')
	user = strings.TrimSpace(user)

	// Ask room name
	fmt.Print("Enter room name: ")
	room, _ := readerInput.ReadString('\n')
	room = strings.TrimSpace(room)

	fmt.Println()

	// get latest messages
	historyResp, err := http.Get("http://localhost:8080/chats/latest?room=" + room)
	if err != nil {
		panic(err)
	}
	defer historyResp.Body.Close()

	json.NewDecoder(historyResp.Body).Decode(&history)

	render(room, user)

	// connect to sse stream
	sseURL := "http://localhost:8080/chats/stream?room=" + room

	streamResp, err := http.Get(sseURL)
	if err != nil {
		panic(err)
	}
	defer streamResp.Body.Close()

	reader := bufio.NewReader(streamResp.Body)
	//fmt.Printf("Connected to SSE stream (room: %s)\n", room)

	// Goroutine: listen for SSE events
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("SSE connection closed:", err)
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, ":") {
				continue
			}
			if strings.Contains(line, "ping") {
				continue
			}

			if after, ok := strings.CutPrefix(line, "data:"); ok {
				raw := strings.TrimSpace(after)

				var msg models.Chat
				if err := json.Unmarshal([]byte(raw), &msg); err != nil {
					fmt.Println("Invalid JSON:", raw)
					continue
				}

				fmt.Printf("[%s] %s : %s\n",
					msg.CreatedAt.Format("2006-01-02 15:04:05"),
					msg.User,
					msg.Message,
				)
				history = append(history, msg)
				render(room, user)
			}
		}
	}()

	// Goroutine: send chat message
	input := bufio.NewScanner(os.Stdin)
	client := &http.Client{}

	for {
		time.Sleep(500 * time.Millisecond)
		fmt.Print("> ")
		if !input.Scan() {
			return
		}

		text := strings.TrimSpace(input.Text())
		if text == "" {
			continue
		}

		// Quit command
		if text == "/quit" || text == "/exit" {
			fmt.Println("Exiting chat...")
			return
		}

		payload := models.Chat{
			User:    user,
			Message: text,
		}

		jsonData, _ := json.Marshal(payload)

		// POST request
		req, _ := http.NewRequest(
			"POST",
			"http://localhost:8080/chats?room="+room,
			bytes.NewBuffer(jsonData),
		)

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Failed to send message:", err)
			continue
		}
		resp.Body.Close()
	}
}

func render(room string, user string) {
	fmt.Print("\033[2J\033[H")

	fmt.Printf("=== ROOM: %s ===\n\n", room)

	// show last 10 messages
	start := 0
	if len(history) > 10 {
		start = len(history) - 10
	}

	for _, msg := range history[start:] {
		fmt.Printf("[%s] %-10s : %s\n",
			msg.CreatedAt.Format("15:04:05"),
			msg.User,
			msg.Message,
		)
	}

	fmt.Println("\n-------------------")
	fmt.Printf("You (%s): ", user)
}
