package application

import (
	"context"
	"errors"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// ChatService — сервис для работы с чатом
type ChatService struct {
	messageRepo domain.MessageRepository
	userRepo    domain.UserRepository
}

// NewChatService создает новый сервис чата
func NewChatService(messageRepo domain.MessageRepository, userRepo domain.UserRepository) *ChatService {
	return &ChatService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

// SendMessageRequest — данные для отправки сообщения
type SendMessageRequest struct {
	ReceiverID uuid.UUID
	Content    string
}

// SendMessage отправляет сообщение
func (s *ChatService) SendMessage(ctx context.Context, senderID uuid.UUID, req SendMessageRequest) (*domain.Message, error) {
	if senderID == req.ReceiverID {
		return nil, domain.ErrCannotSendToSelf
	}

	if req.Content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	// Проверяем, что получатель существует
	receiver, err := s.userRepo.FindByID(ctx, req.ReceiverID)
	if err != nil {
		return nil, err
	}
	if receiver == nil {
		return nil, domain.ErrUserNotFound
	}

	message := &domain.Message{
		ID:         uuid.New(),
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		IsRead:     false,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

// GetConversation возвращает переписку между двумя пользователями
func (s *ChatService) GetConversation(ctx context.Context, user1ID, user2ID uuid.UUID, limit, offset int) ([]domain.Message, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.GetConversation(ctx, user1ID, user2ID, limit, offset)
}

// GetUnreadCount возвращает количество непрочитанных сообщений
func (s *ChatService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.messageRepo.GetUnreadCount(ctx, userID)
}

// GetUnreadCountFromUser возвращает количество непрочитанных сообщений от конкретного пользователя
func (s *ChatService) GetUnreadCountFromUser(ctx context.Context, userID, senderID uuid.UUID) (int64, error) {
	return s.messageRepo.GetUnreadCountFromUser(ctx, userID, senderID)
}

// MarkAsRead отмечает сообщение как прочитанное
func (s *ChatService) MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return domain.ErrMessageNotFound
	}

	// Только получатель может отметить сообщение как прочитанное
	if message.ReceiverID != userID {
		return domain.ErrForbidden
	}

	return s.messageRepo.MarkAsRead(ctx, messageID)
}

// MarkAllAsRead отмечает все сообщения от пользователя как прочитанные
func (s *ChatService) MarkAllAsRead(ctx context.Context, userID, senderID uuid.UUID) error {
	return s.messageRepo.MarkAllAsRead(ctx, userID, senderID)
}

// GetRecentChats возвращает последние чаты пользователя
func (s *ChatService) GetRecentChats(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	return s.messageRepo.GetRecentChats(ctx, userID, limit)
}

// DeleteMessage удаляет сообщение
func (s *ChatService) DeleteMessage(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return domain.ErrMessageNotFound
	}

	// Только отправитель может удалить сообщение
	if message.SenderID != userID {
		return domain.ErrForbidden
	}

	return s.messageRepo.Delete(ctx, messageID)
}

// GetMessage возвращает сообщение по ID
func (s *ChatService) GetMessage(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	return s.messageRepo.GetByID(ctx, id)
}