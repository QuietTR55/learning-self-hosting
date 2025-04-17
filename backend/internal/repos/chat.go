package repos

import (
	"backend/internal/database"
	"context"
)

type ChatRepo struct {
	db *database.Queries
}

func NewChatRepo(db *database.Queries) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) GetRecentMessages() ([]database.Message, error) {
	return r.db.GetRecentMessages(context.Background());
}

func (r *ChatRepo) CreateMessage(message string) error {
	return r.db.CreateMessage(context.Background(), message)
}
