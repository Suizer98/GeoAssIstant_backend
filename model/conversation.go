package model

import "time"

type Conversation struct {
	Id            uint        `json:"id"`
	UserId        uint        `json:"user_id"`
	ConversationId string      `json:"conversation_id"`
	ChatHistory   interface{} `json:"chat_history"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}
