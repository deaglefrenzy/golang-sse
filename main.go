package main

import (
	"log"

	"github.com/deaglefrenzy/golang-sse/db"
	"github.com/deaglefrenzy/golang-sse/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	mongoClient, ctx, cancel, err := db.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	defer cancel()
	defer mongoClient.Disconnect(ctx)

	mdb := mongoClient.Database("chatroom")
	mongoChats := mdb.Collection("chats")

	r := gin.Default()

	r.GET("/chats/stream", func(c *gin.Context) {
		room := c.Query("room")
		handlers.StreamChats(c, mongoChats, room)
	})

	r.POST("/chats", func(c *gin.Context) {
		room := c.Query("room")
		handlers.InsertMessage(c, mongoChats, room)
	})

	r.GET("/chats/latest", func(c *gin.Context) {
		room := c.Query("room")
		handlers.GetLatestChats(c, mongoChats, room)
	})

	r.Run(":8080")
}
