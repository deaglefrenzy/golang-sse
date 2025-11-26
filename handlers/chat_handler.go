package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/deaglefrenzy/golang-sse/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InsertMessage(c *gin.Context, col *mongo.Collection, room string) {
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
