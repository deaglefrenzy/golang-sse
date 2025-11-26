package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/deaglefrenzy/golang-sse/models"
)

func main() {
	room := "lobby"
	resp, err := http.Get("http://localhost:8080/chats/stream?room=" + room)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	fmt.Printf("Connected to room: %s\n", room)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed:", err)
			return
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
		}
	}
}
