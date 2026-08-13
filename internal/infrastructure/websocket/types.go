package websocket

import (
	"encoding/json"
	"time"
)

// MessageType определяет тип WebSocket сообщения
type MessageType string

const (
	// Типы сообщений от клиента
	MsgTypeChat      MessageType = "chat"
	MsgTypeRead      MessageType = "read"
	MsgTypeTyping    MessageType = "typing"
	MsgTypeStopTyping MessageType = "stop_typing"

	// Типы сообщений от сервера
	MsgTypeNewMessage  MessageType = "new_message"
	MsgTypeMessageRead MessageType = "message_read"
	MsgTypeUserTyping  MessageType = "user_typing"
	MsgTypeUserOnline  MessageType = "user_online"
	MsgTypeUserOffline MessageType = "user_offline"
	MsgTypeError       MessageType = "error"
)

// WSMessage структура WebSocket сообщения
type WSMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ChatPayload структура для чат сообщения
type ChatPayload struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"sender_id"`
	ReceiverID string   `json:"receiver_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ReadPayload структура для отметки прочитанного
type ReadPayload struct {
	MessageID string `json:"message_id"`
}

// TypingPayload структура для индикатора печатания
type TypingPayload struct {
	ReceiverID string `json:"receiver_id"`
	IsTyping   bool   `json:"is_typing"`
}

// ErrorPayload структура для ошибок
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}