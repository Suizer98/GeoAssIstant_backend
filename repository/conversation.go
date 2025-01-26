package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"geoai-app/model"
)

type ConversationRepositoryInterface interface {
	GetConversationsByUserID(userID string) ([]model.Conversation, error)
	CreateConversation(conversation *model.Conversation) error
	GetLatestConversationByUserID(userID int) (*model.Conversation, error)
	UpdateConversation(conversation *model.Conversation) error
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

func (r *ConversationRepository) CreateConversation(conversation *model.Conversation) error {
	// Marshal chat history to JSON
	chatHistoryJSON, err := json.Marshal(conversation.ChatHistory)
	if err != nil {
		fmt.Printf("Error marshaling chat history: %v\n", err)
		return err
	}

	// Log the conversation and marshaled chat history
	fmt.Printf("Saving new conversation to database: %+v\n", conversation)
	fmt.Printf("Marshaled chat history for SQL: %s\n", string(chatHistoryJSON))

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

func (r *ConversationRepository) GetLatestConversationByUserID(userID int) (*model.Conversation, error) {
	row := r.DB.QueryRow(
		"SELECT id, user_id, conversation_id, chat_history, created_at, updated_at FROM conversations WHERE user_id = $1 ORDER BY updated_at DESC LIMIT 1",
		userID,
	)

	var conversation model.Conversation
	err := row.Scan(&conversation.ID, &conversation.UserID, &conversation.ConversationID, &conversation.ChatHistory, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // No existing conversation
	}
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *ConversationRepository) UpdateConversation(conversation *model.Conversation) error {
	// Marshal chat history to JSON
	chatHistoryJSON, err := json.Marshal(conversation.ChatHistory)
	if err != nil {
		fmt.Printf("Error marshaling chat history: %v\n", err)
		return err
	}

	// Log the marshaled chat history
	fmt.Printf("Marshaled chat history for update: %s\n", string(chatHistoryJSON))

	// Update the conversation in the database
	_, err = r.DB.Exec(
		"UPDATE conversations SET chat_history = $1::JSONB, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		chatHistoryJSON, conversation.ID,
	)
	if err != nil {
		fmt.Printf("SQL Error while updating conversation: %v\n", err)
	}
	return err
}
