package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// CommentRepository — репозиторий для работы с комментариями
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository создает новый репозиторий
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create создает новый комментарий
func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// FindByPostID возвращает комментарии поста
func (r *CommentRepository) FindByPostID(ctx context.Context, postID uuid.UUID) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Preload("Author").
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// FindByID ищет комментарий по ID
func (r *CommentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	var comment domain.Comment
	err := r.db.WithContext(ctx).First(&comment, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &comment, err
}

// Delete удаляет комментарий
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Comment{}, "id = ?", id).Error
}

// CountByPostID возвращает количество комментариев у поста
func (r *CommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Comment{}).Where("post_id = ?", postID).Count(&count).Error
	return count, err
}