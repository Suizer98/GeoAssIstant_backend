package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"geoai-app/model"
)

type ConversationRepositoryInterface interface {
	GetConversationsByUserID(userID string) ([]model.Conversation, error)
	GetConversationByUUID(uuid string) (*model.Conversation, error)
	CreateConversation(conversation *model.Conversation) error
	UpdateConversation(conversation *model.Conversation) error
}

type ConversationRepository struct {
	DB *sql.DB
}

func NewConversationRepository(db *sql.DB) ConversationRepositoryInterface {
	return &ConversationRepository{DB: db}
}

// GetConversationsByUserID retrieves all conversations for a user
func (r *ConversationRepository) GetConversationsByUserID(userID string) ([]model.Conversation, error) {
	rows, err := r.DB.Query(
		"SELECT id, user_id, conversation_id, chat_history, created_at, updated_at FROM conversations WHERE user_id = $1 ORDER BY created_at ASC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []model.Conversation
	for rows.Next() {
		var conversation model.Conversation
		var chatHistory string

		err := rows.Scan(&conversation.ID, &conversation.UserID, &conversation.ConversationID, &chatHistory, &conversation.CreatedAt, &conversation.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON chat history
		err = json.Unmarshal([]byte(chatHistory), &conversation.ChatHistory)
		if err != nil {
			return nil, err
		}

		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

// GetConversationByUUID retrieves a conversation by its UUID
func (r *ConversationRepository) GetConversationByUUID(uuid string) (*model.Conversation, error) {
	row := r.DB.QueryRow(
		"SELECT id, user_id, conversation_id, chat_history, created_at, updated_at FROM conversations WHERE conversation_id = $1",
		uuid,
	)

	var conversation model.Conversation
	var chatHistory string

	err := row.Scan(&conversation.ID, &conversation.UserID, &conversation.ConversationID, &chatHistory, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // No existing conversation
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON chat history
	err = json.Unmarshal([]byte(chatHistory), &conversation.ChatHistory)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

// CreateConversation creates a new conversation in the database
func (r *ConversationRepository) CreateConversation(conversation *model.Conversation) error {
	// Marshal chat history to JSON
	chatHistoryJSON, err := json.Marshal(conversation.ChatHistory)
	if err != nil {
		fmt.Printf("Error marshaling chat history: %v\n", err)
		return err
	}

	// Insert into database
	err = r.DB.QueryRow(
		"INSERT INTO conversations (user_id, conversation_id, chat_history) VALUES ($1, $2, $3::JSONB) RETURNING id, created_at, updated_at",
		conversation.UserID, conversation.ConversationID, chatHistoryJSON,
	).Scan(&conversation.ID, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err != nil {
		fmt.Printf("SQL Error while creating conversation: %v\n", err)
	}
	return err
}

// UpdateConversation updates an existing conversation in the database
func (r *ConversationRepository) UpdateConversation(conversation *model.Conversation) error {
	// Marshal chat history to JSON
	chatHistoryJSON, err := json.Marshal(conversation.ChatHistory)
	if err != nil {
		fmt.Printf("Error marshaling chat history: %v\n", err)
		return err
	}

	// Update the conversation in the database
	_, err = r.DB.Exec(
		"UPDATE conversations SET chat_history = $1::JSONB, updated_at = CURRENT_TIMESTAMP WHERE conversation_id = $2",
		chatHistoryJSON, conversation.ConversationID,
	)
	if err != nil {
		fmt.Printf("SQL Error while updating conversation: %v\n", err)
	}
	return err
}
