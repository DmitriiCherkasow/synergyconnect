package database

import (
	"context"
	"errors"
	"time"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

// Create implements domain.MessageRepository
func (r *messageRepository) Create(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetByID implements domain.MessageRepository
func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	var message domain.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Where("id = ?", id).
		First(&message).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMessageNotFound
		}
		return nil, err
	}
	return &message, nil
}

// GetConversation implements domain.MessageRepository
func (r *messageRepository) GetConversation(ctx context.Context, user1ID, user2ID uuid.UUID, limit, offset int) ([]domain.Message, int64, error) {
	var messages []domain.Message
	var total int64

	query := r.db.WithContext(ctx).
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			user1ID, user2ID, user2ID, user1ID,
		)

	// Подсчитываем общее количество
	if err := query.Model(&domain.Message{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем сообщения с пагинацией
	err := query.
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	// Разворачиваем, чтобы получить в хронологическом порядке
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, err
}

// GetUnreadCount implements domain.MessageRepository
func (r *messageRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead implements domain.MessageRepository
func (r *messageRepository) MarkAsRead(ctx context.Context, messageID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// MarkAllAsRead implements domain.MessageRepository
func (r *messageRepository) MarkAllAsRead(ctx context.Context, userID, senderID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("receiver_id = ? AND sender_id = ? AND is_read = ?", userID, senderID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// GetRecentChats implements domain.MessageRepository
func (r *messageRepository) GetRecentChats(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Message, error) {
	var messages []domain.Message

	// Получаем последние сообщения для каждого собеседника
	// Используем подзапрос с оконной функцией
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (other_user_id) 
				m.* 
			FROM (
				SELECT 
					id, sender_id, receiver_id, content, is_read, read_at, created_at,
					CASE 
						WHEN sender_id = ? THEN receiver_id 
						ELSE sender_id 
					END as other_user_id
				FROM messages 
				WHERE sender_id = ? OR receiver_id = ?
				ORDER BY other_user_id, created_at DESC
			) m
			ORDER BY other_user_id, created_at DESC
			LIMIT ?
		`, userID, userID, userID, limit).
		Preload("Sender").
		Preload("Receiver").
		Find(&messages).Error

	return messages, err
}

// Delete implements domain.MessageRepository
func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Message{}, "id = ?", id).Error
}

// DeleteConversation implements domain.MessageRepository
func (r *messageRepository) DeleteConversation(ctx context.Context, user1ID, user2ID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			user1ID, user2ID, user2ID, user1ID,
		).
		Delete(&domain.Message{}).Error
}