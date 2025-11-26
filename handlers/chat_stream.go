package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deaglefrenzy/golang-sse/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StreamChats(c *gin.Context, col *mongo.Collection, room string) {
	// SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := c.Request.Context()

	match := bson.D{
		{Key: "operationType", Value: "insert"},
	}
	if room != "" {
		match = append(match, bson.E{Key: "fullDocument.room", Value: room})
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
	}

	// change stream
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	changeStream, err := col.Watch(ctx, pipeline, opts)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to watch change stream")
		return
	}
	defer changeStream.Close(ctx)

	// SSE loop
	for changeStream.Next(ctx) {
		var event models.ChangeEvent
		if err := changeStream.Decode(&event); err != nil {
			continue
		}

		jsonBytes, err := json.Marshal(event.FullDocument)
		if err != nil {
			continue
		}
		fmt.Fprintf(c.Writer, "data: %s\n\n", jsonBytes)
		c.Writer.Flush()
	}
}
