package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/deaglefrenzy/golang-sse/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rooms = make(map[string]*Streamer)
var roomsMu sync.Mutex

func StartMongoWatcher(col *mongo.Collection) {
	// building pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: "insert"},
		}}},
	}

	// mongo change stream
	ctx := context.Background()
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	stream, err := col.Watch(ctx, pipeline, opts)
	if err != nil {
		log.Fatal("cannot start change stream:", err)
	}

	// put the stream into a goroutine
	go func() {
		defer stream.Close(ctx)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			// send heartbeats to all rooms
			select {
			case <-ticker.C:
				roomsMu.Lock()
				for _, b := range rooms {
					b.Stream([]byte(`{"type":"ping"}`))
				}
				roomsMu.Unlock()

			default:
				// TryNext is a non-blocking mongo read (instead of Next)
				if !stream.TryNext(ctx) {
					time.Sleep(30 * time.Millisecond)
					continue
				}

				var event models.ChangeEvent
				stream.Decode(&event)

				room := event.FullDocument.Room
				if room == "" {
					continue
				}

				b := GetRoom(room) // get the clients/streamer from specified room

				// send to the room
				msg, _ := json.Marshal(event.FullDocument)
				b.Stream(msg) // stream message to all client sse
			}
		}
	}()
}

func SSEHandler(c *gin.Context) {
	room := c.Query("room") // get room from sse URL
	if room == "" {
		c.String(400, "room required")
		return
	}

	// SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// get streamer from specified room
	b := GetRoom(room)

	// create new client
	ch := make(chan []byte, 10)
	b.AddClient(ch)
	defer b.RemoveClient(ch)

	// loop for watching the channels
	for {
		select {
		case <-c.Request.Context().Done():
			return

		case msg := <-ch: // if theres message in channel then flush data
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			c.Writer.Flush()
		}
	}
}

func GetRoom(room string) *Streamer {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if b, ok := rooms[room]; ok { // room exists, return client
		return b
	}

	b := NewStreamer() // room doesn't exists yet
	rooms[room] = b    // create new one in a map
	return b           // return the new room
}
