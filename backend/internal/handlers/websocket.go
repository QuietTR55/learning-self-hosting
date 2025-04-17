package handlers

import (
	"backend/internal/websocket"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	WebSocketManager *websocket.ClientManager
}

func NewWebSocketHandler(webSocketManager *websocket.ClientManager) *WebSocketHandler {
	return &WebSocketHandler{
		WebSocketManager: webSocketManager,
	}
}

var upgrader = gorilla.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	// Allow all origins for development (Tighten this in production!)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Log entry point
	log.Println("HandleWebSocket called by:", c.Request.RemoteAddr)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket for %s: %v", c.Request.RemoteAddr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not open websocket connection"})
		return
	}

	client := &websocket.Client{
		Conn: conn,
	}

	h.WebSocketManager.Register <- client

	go client.ReadPump(h.WebSocketManager)

	log.Println("WebSocket client connected and pumps started:", conn.RemoteAddr())
}