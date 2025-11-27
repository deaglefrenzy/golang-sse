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

	handlers.StartMongoWatcher(mongoChats)

	r := gin.Default()

	r.GET("/chats/stream", handlers.SSEHandler)
	r.POST("/chats", func(c *gin.Context) { handlers.InsertMessage(c, mongoChats) })
	r.GET("/chats/latest", func(c *gin.Context) { handlers.GetLatestChats(c, mongoChats) })

	r.Run(":8080")
}
