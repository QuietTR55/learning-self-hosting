package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
}

type ClientManager struct {
	clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	mu         sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte, 256),
	}
}

func (m *ClientManager) Run() {
	log.Println("WebSocket Manager Run() loop started...")
	for {
		select {
		case client := <-m.Register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()
			log.Printf("Client registered: %p", client.Conn.RemoteAddr())

		case client := <-m.Unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				client.Conn.Close()
				log.Printf("Client unregistered: %p", client.Conn.RemoteAddr())
			}
			m.mu.Unlock()

		case message := <-m.Broadcast:
			m.mu.Lock()
			currentClients := make([]*Client, 0, len(m.clients))
			for client := range m.clients {
				currentClients = append(currentClients, client)
			}
			m.mu.Unlock()

			for _, client := range currentClients {
				err := client.Conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error broadcasting to client %p: %v", client.Conn.RemoteAddr(), err)
				}
			}
		}
	}
}

func (c *Client) ReadPump(manager *ClientManager) {
	defer func() {
		manager.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(256)
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); return nil })

	for {
		messageType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for %p: %v", c.Conn.RemoteAddr(), err)
			} else {
				log.Printf("WebSocket connection closed gracefully: %p", c.Conn.RemoteAddr())
			}
			break
		}

		log.Printf("Received message type %d from %p: %s", messageType, c.Conn.RemoteAddr(), message)

		manager.Broadcast <- message
	}
}
