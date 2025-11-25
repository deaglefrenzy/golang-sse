package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx := r.Context()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Client disconnected")
				return

			default:
				num := rand.Intn(1000)
				fmt.Println("Number sent by server:", num)

				fmt.Fprintf(w, "%d\n", num)
				w.(http.Flusher).Flush()
				time.Sleep(1 * time.Second)
			}
		}
	})

	fmt.Println("Server running on :8080 streaming numbers")
	http.ListenAndServe(":8080", nil)
}
