package main

import (
	"backend/internal/dependencyinjection"
	"backend/internal/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	container := dependencyinjection.NewContainer()

	// Start the WebSocket manager's processing loop in the background
	if container.WebSocketManager == nil {
		log.Fatal("WebSocketManager is nil in container!")
	}
	go container.WebSocketManager.Run()
	log.Println("Started WebSocket Manager goroutine")

	// Assuming ChatHandler is correctly initialized
	if container.ChatHandler != nil {
		router.GET("/chat", container.ChatHandler.GetChat)
		router.POST("/chat", container.ChatHandler.SendMessage)
	} else {
		log.Println("WARN: ChatHandler is nil in container, /chat route skipped")
	}

	// Create WebSocketHandler instance and register route
	if container.WebSocketHandler != nil {
		router.GET("/ws", container.WebSocketHandler.HandleWebSocket)
	} else if container.WebSocketManager != nil {
		wsHandler := handlers.NewWebSocketHandler(container.WebSocketManager)
		router.GET("/ws", wsHandler.HandleWebSocket)
	} else {
		log.Println("WARN: WebSocketManager is nil, cannot set up /ws route")
	}
}
