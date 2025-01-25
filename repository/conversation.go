package repository

import (
	"database/sql"
	"geoai-app/model"
)

type ConversationRepositoryInterface interface {
	GetConversationsByUserID(userID string) ([]model.Conversation, error)
	CreateConversation(conversation *model.Conversation) error
}

type ConversationRepository struct {
	DB *sql.DB
}

func NewConversationRepository(db *sql.DB) ConversationRepositoryInterface {
	return &ConversationRepository{DB: db}
}

// GetConversationsByUserId retrieves all conversations for a user
func (r *ConversationRepository) GetConversationsByUserID(userID string) ([]model.Conversation, error) {
	rows, err := r.DB.Query(
		"SELECT id, user_id, conversation_id, chat_history, created_at, updated_at FROM conversations WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []model.Conversation
	for rows.Next() {
		var conversation model.Conversation
		err := rows.Scan(&conversation.ID, &conversation.UserID, &conversation.ConversationID, &conversation.ChatHistory, &conversation.CreatedAt, &conversation.UpdatedAt)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}
	return conversations, nil
}

// CreateConversation inserts a new conversation into the database
func (r *ConversationRepository) CreateConversation(conversation *model.Conversation) error {
	err := r.DB.QueryRow(
		"INSERT INTO conversations (user_id, conversation_id, chat_history) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		conversation.UserID, conversation.ConversationID, conversation.ChatHistory,
	).Scan(&conversation.ID, &conversation.CreatedAt, &conversation.UpdatedAt)
	return err
}
