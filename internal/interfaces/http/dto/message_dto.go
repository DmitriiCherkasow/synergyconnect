package dto

import (
	"time"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// SendMessageRequest — запрос на отправку сообщения
type SendMessageRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// MessageResponse — ответ с сообщением
type MessageResponse struct {
	ID         string     `json:"id"`
	SenderID   string     `json:"sender_id"`
	ReceiverID string     `json:"receiver_id"`
	Content    string     `json:"content"`
	IsRead     bool       `json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
}

// MessageListResponse — список сообщений
type MessageListResponse struct {
	Messages    []MessageResponse `json:"messages"`
	Total       int64             `json:"total"`
	UnreadCount int64             `json:"unread_count,omitempty"`
}

// ConversationResponse — переписка с пользователем
type ConversationResponse struct {
	User        UserResponse      `json:"user"`
	Messages    []MessageResponse `json:"messages"`
	Total       int64             `json:"total"`
	UnreadCount int64             `json:"unread_count"`
}

// RecentChatResponse — последний чат с пользователем
type RecentChatResponse struct {
	User        UserResponse    `json:"user"`
	LastMessage MessageResponse `json:"last_message"`
	UnreadCount int64           `json:"unread_count"`
}

// RecentChatsResponse — список последних чатов
type RecentChatsResponse struct {
	Chats []RecentChatResponse `json:"chats"`
}

// MessageReadRequest — запрос на отметку прочитанных (для чата)
type MessageReadRequest struct {
	MessageID string `json:"message_id"`
	SenderID  string `json:"sender_id,omitempty"`
}

// ToMessageResponse конвертирует доменную модель в DTO
func ToMessageResponse(msg *domain.Message) MessageResponse {
	return MessageResponse{
		ID:         msg.ID.String(),
		SenderID:   msg.SenderID.String(),
		ReceiverID: msg.ReceiverID.String(),
		Content:    msg.Content,
		IsRead:     msg.IsRead,
		CreatedAt:  msg.CreatedAt,
		ReadAt:     msg.ReadAt,
	}
}