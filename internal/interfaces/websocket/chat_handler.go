package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/infrastructure/websocket"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatWebSocketHandler — обработчик WebSocket для чата
type ChatWebSocketHandler struct {
	hub         *websocket.Hub
	chatService *application.ChatService
}

// NewChatWebSocketHandler создает новый WebSocket обработчик
func NewChatWebSocketHandler(
	hub *websocket.Hub,
	chatService *application.ChatService,
) *ChatWebSocketHandler {
	return &ChatWebSocketHandler{
		hub:         hub,
		chatService: chatService,
	}
}

// HandleWebSocket обрабатывает WebSocket соединение
func (h *ChatWebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Получаем user_id из контекста (уже проверен middleware)
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	// Обновляем соединение до WebSocket
	err = h.hub.UpgradeToWebSocket(c.Writer, c.Request, userID.String())
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed"})
		return
	}

	// Обработчик сообщений для этого клиента будет установлен в Hub
	h.hub.SetMessageHandler(h.handleMessage)
}

// handleMessage обрабатывает входящие WebSocket сообщения
func (h *ChatWebSocketHandler) handleMessage(client *websocket.Client, msg *websocket.WSMessage) {
	// Получаем userID из Client (используем метод, если он есть, или храним отдельно)
	userID, err := uuid.Parse(client.GetUserID())
	if err != nil {
		client.SendError("invalid_user", "Invalid user ID")
		return
	}

	switch msg.Type {
	case websocket.MsgTypeChat:
		h.handleChatMessage(client, userID, msg)
	case websocket.MsgTypeRead:
		h.handleReadMessage(client, userID, msg)
	case websocket.MsgTypeTyping:
		h.handleTypingMessage(client, userID, msg, true)
	case websocket.MsgTypeStopTyping:
		h.handleTypingMessage(client, userID, msg, false)
	default:
		client.SendError("unknown_type", "Unknown message type")
	}
}

// handleChatMessage обрабатывает чат сообщение
func (h *ChatWebSocketHandler) handleChatMessage(client *websocket.Client, userID uuid.UUID, msg *websocket.WSMessage) {
	var payload websocket.ChatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		client.SendError("invalid_payload", "Invalid chat payload")
		return
	}

	receiverID, err := uuid.Parse(payload.ReceiverID)
	if err != nil {
		client.SendError("invalid_receiver", "Invalid receiver ID")
		return
	}

	// Сохраняем сообщение в БД
	message, err := h.chatService.SendMessage(context.Background(), userID, application.SendMessageRequest{
		ReceiverID: receiverID,
		Content:    payload.Content,
	})
	if err != nil {
		if err == domain.ErrCannotSendToSelf {
			client.SendError("cannot_send_to_self", "Cannot send message to yourself")
			return
		}
		if err == domain.ErrUserNotFound {
			client.SendError("receiver_not_found", "Receiver not found")
			return
		}
		client.SendError("send_error", err.Error())
		return
	}

	// Отправляем подтверждение отправителю
	response := &websocket.WSMessage{
		Type: websocket.MsgTypeNewMessage,
		Payload: json.RawMessage(`{
			"id":"` + message.ID.String() + `",
			"sender_id":"` + message.SenderID.String() + `",
			"receiver_id":"` + message.ReceiverID.String() + `",
			"content":"` + message.Content + `",
			"created_at":"` + message.CreatedAt.Format("2006-01-02T15:04:05Z") + `"
		}`),
	}
	client.SendMessage(response)

	// Отправляем сообщение получателю, если он онлайн
	h.hub.SendToUser(receiverID.String(), response)
}

// handleReadMessage обрабатывает отметку прочитанного
func (h *ChatWebSocketHandler) handleReadMessage(client *websocket.Client, userID uuid.UUID, msg *websocket.WSMessage) {
	var payload websocket.ReadPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		client.SendError("invalid_payload", "Invalid read payload")
		return
	}

	messageID, err := uuid.Parse(payload.MessageID)
	if err != nil {
		client.SendError("invalid_message_id", "Invalid message ID")
		return
	}

	err = h.chatService.MarkAsRead(context.Background(), messageID, userID)
	if err != nil {
		if err == domain.ErrMessageNotFound {
			client.SendError("message_not_found", "Message not found")
			return
		}
		if err == domain.ErrForbidden {
			client.SendError("forbidden", "You cannot read this message")
			return
		}
		client.SendError("read_error", err.Error())
		return
	}

	// Уведомляем отправителя, что сообщение прочитано
	message, _ := h.chatService.GetMessage(context.Background(), messageID)
	if message != nil {
		h.hub.SendToUser(message.SenderID.String(), &websocket.WSMessage{
			Type: websocket.MsgTypeMessageRead,
			Payload: json.RawMessage(`{"message_id":"` + payload.MessageID + `"}`),
		})
	}
}

// handleTypingMessage обрабатывает индикатор печатания
func (h *ChatWebSocketHandler) handleTypingMessage(client *websocket.Client, userID uuid.UUID, msg *websocket.WSMessage, isTyping bool) {
	var payload websocket.TypingPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		client.SendError("invalid_payload", "Invalid typing payload")
		return
	}

	// Отправляем индикатор получателю
	typingMsg := &websocket.WSMessage{
		Type: websocket.MsgTypeUserTyping,
		Payload: json.RawMessage(`{
			"sender_id":"` + client.GetUserID() + `",
			"is_typing":` + map[bool]string{true: "true", false: "false"}[isTyping] + `
		}`),
	}
	h.hub.SendToUser(payload.ReceiverID, typingMsg)
}