package handlers

import (
	redisClient "backend/internal/redisclient"
	"backend/internal/repos"
	"backend/internal/websocket"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	repo *repos.ChatRepo
	ws *websocket.ClientManager
}

func NewChatHandler(repo *repos.ChatRepo, ws *websocket.ClientManager) *ChatHandler {
	return &ChatHandler{repo: repo, ws: ws}
}

func (h *ChatHandler) GetChat(c *gin.Context) {
	messages, err := h.repo.GetRecentMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}


func (h *ChatHandler) SendMessage(c *gin.Context) {
	message := c.PostForm("message")
	log.Println(message)
	err := h.repo.CreateMessage(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


	redisClient.Rdb.Publish(redisClient.Ctx, "messages", message)
	c.JSON(http.StatusOK, gin.H{"message": "Message created successfully"})
}
