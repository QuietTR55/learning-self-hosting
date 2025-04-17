package dependencyinjection

import (
	"backend/internal/database"
	"backend/internal/handlers"
	"backend/internal/repos"
	"backend/internal/websocket"
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	DbPool *pgxpool.Pool
	Queries *database.Queries
	ChatRepo *repos.ChatRepo
	ChatHandler *handlers.ChatHandler
	WebSocketManager *websocket.ClientManager
	WebSocketHandler *handlers.WebSocketHandler
}

func NewContainer() *Container {
	postgreConnectionUrl := os.Getenv("POSTGRES_URL")
	if postgreConnectionUrl == "" {
		log.Fatal("POSTGRES_URL environment variable not set")
	}

	dbPool, err := pgxpool.New(context.Background(), postgreConnectionUrl)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	queries := database.New(dbPool)
	chatRepo := repos.NewChatRepo(queries)
	websocketManager := websocket.NewClientManager()
	chatHandler := handlers.NewChatHandler(chatRepo, websocketManager)
	websocketHandler := handlers.NewWebSocketHandler(websocketManager)
	return &Container{
		DbPool: dbPool,
		Queries: queries,
		ChatRepo: chatRepo,
		ChatHandler: chatHandler,
		WebSocketManager: websocketManager,
		WebSocketHandler: websocketHandler,
	}
}
