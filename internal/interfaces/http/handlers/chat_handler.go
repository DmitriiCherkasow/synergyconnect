package handlers

import (
	"net/http"
	"strconv"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/dto"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatHandler — обработчик для чата
type ChatHandler struct {
	chatService *application.ChatService
	// notificationService *application.NotificationService // добавим позже
}

// NewChatHandler создает новый обработчик
func NewChatHandler(chatService *application.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// getUserID извлекает ID пользователя из контекста
func (h *ChatHandler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// SendMessage — отправка сообщения
// @Summary Отправка сообщения
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SendMessageRequest true "Данные сообщения"
// @Success 201 {object} dto.MessageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/chat/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receiver_id"})
		return
	}

	message, err := h.chatService.SendMessage(c.Request.Context(), userID, application.SendMessageRequest{
		ReceiverID: receiverID,
		Content:    req.Content,
	})
	if err != nil {
		if err == domain.ErrCannotSendToSelf {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot send message to yourself"})
			return
		}
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "receiver not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToMessageResponse(message))
}

// GetConversation — получение переписки с пользователем
// @Summary Получение переписки
// @Tags chat
// @Security BearerAuth
// @Produce json
// @Param userId path string true "ID пользователя"
// @Param limit query int false "Лимит" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} dto.ConversationResponse
// @Router /api/v1/chat/messages/{userId} [get]
func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	otherUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	messages, total, err := h.chatService.GetConversation(c.Request.Context(), userID, otherUserID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	unreadCount, err := h.chatService.GetUnreadCountFromUser(c.Request.Context(), userID, otherUserID)
	if err != nil {
		unreadCount = 0
	}

	response := dto.ConversationResponse{
		Messages:    make([]dto.MessageResponse, len(messages)),
		Total:       total,
		UnreadCount: unreadCount,
	}

	for i, msg := range messages {
		response.Messages[i] = dto.ToMessageResponse(&msg)
	}

	c.JSON(http.StatusOK, response)
}

// GetUnreadCount — получение количества непрочитанных сообщений
// @Summary Количество непрочитанных сообщений
// @Tags chat
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]int64
// @Router /api/v1/chat/unread/count [get]
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.chatService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkAsRead — отметить сообщение как прочитанное
// @Summary Отметить сообщение как прочитанное
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MarkReadRequest true "Данные для отметки"
// @Success 200 {object} map[string]string
// @Router /api/v1/chat/read [put]
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.MessageReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MessageID != "" {
		messageID, err := uuid.Parse(req.MessageID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
			return
		}
		err = h.chatService.MarkAsRead(c.Request.Context(), messageID, userID)
		if err != nil {
			if err == domain.ErrMessageNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			if err == domain.ErrForbidden {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if req.SenderID != "" {
		senderID, err := uuid.Parse(req.SenderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sender_id"})
			return
		}
		err = h.chatService.MarkAllAsRead(c.Request.Context(), userID, senderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id or sender_id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message(s) marked as read"})
}

// GetRecentChats — получение последних чатов
// @Summary Получение последних чатов
// @Tags chat
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Лимит" default(20)
// @Success 200 {object} dto.RecentChatsResponse
// @Router /api/v1/chat/recent [get]
func (h *ChatHandler) GetRecentChats(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages, err := h.chatService.GetRecentChats(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.RecentChatsResponse{
		Chats: make([]dto.RecentChatResponse, 0, len(messages)),
	}

	for _, msg := range messages {
		// Определяем собеседника
		var otherUser *domain.User
		if msg.SenderID == userID {
			otherUser = &msg.Receiver
		} else {
			otherUser = &msg.Sender
		}

		// Получаем количество непрочитанных от этого пользователя
		unreadCount, _ := h.chatService.GetUnreadCountFromUser(c.Request.Context(), userID, otherUser.ID)

		chat := dto.RecentChatResponse{
			User: dto.UserResponse{
				ID:         otherUser.ID.String(),
				Email:      otherUser.Email,
				FirstName:  otherUser.FirstName,
				LastName:   otherUser.LastName,
				AvatarURL:  otherUser.AvatarURL,
				IsVerified: otherUser.IsVerified,
			},
			LastMessage: dto.ToMessageResponse(&msg),
			UnreadCount: unreadCount,
		}
		response.Chats = append(response.Chats, chat)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteMessage — удаление сообщения
// @Summary Удаление сообщения
// @Tags chat
// @Security BearerAuth
// @Param id path string true "ID сообщения"
// @Success 204 "No Content"
// @Router /api/v1/chat/messages/{id} [delete]
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	err = h.chatService.DeleteMessage(c.Request.Context(), messageID, userID)
	if err != nil {
		if err == domain.ErrMessageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}