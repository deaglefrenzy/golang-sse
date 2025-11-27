package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/deaglefrenzy/golang-sse/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertMessage(c *gin.Context, col *mongo.Collection) {
	room := c.Query("room")
	var message models.Chat

	if err := c.BindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message.CreatedAt = time.Now()
	message.Room = room

	ctx := c.Request.Context()
	_, err := col.InsertOne(ctx, message)
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notice":  "Message Sent",
		"message": message,
	})
}

func GetLatestChats(c *gin.Context, col *mongo.Collection) {
	room := c.Query("room")
	ctx := c.Request.Context()

	limit := 5

	filter := bson.M{}
	if room != "" {
		filter["room"] = room
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}
	defer cursor.Close(ctx)

	var messages []models.Chat
	if err := cursor.All(ctx, &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode messages"})
		return
	}

	// reverse order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, messages)
}
