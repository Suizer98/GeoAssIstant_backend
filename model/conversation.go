package model

import "time"

type Conversation struct {
	ID             uint        `json:"id"`
	UserID         uint        `json:"user_id"`
	ConversationID string      `json:"conversation_id"`
	ChatHistory    interface{} `json:"chat_history"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}
