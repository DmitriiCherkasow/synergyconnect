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

// List implements domain.CommentRepository
func (r *CommentRepository) List(ctx context.Context, filter domain.CommentFilter) ([]domain.Comment, int64, error) {
    var comments []domain.Comment
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.Comment{})

    if filter.PostID != nil {
        query = query.Where("post_id = ?", *filter.PostID)
    }
    if filter.AuthorID != nil {
        query = query.Where("author_id = ?", *filter.AuthorID)
    }
    if filter.Search != "" {
        query = query.Where("content ILIKE ?", "%"+filter.Search+"%")
    }

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if filter.Limit > 0 {
        query = query.Limit(filter.Limit)
    }
    if filter.Offset > 0 {
        query = query.Offset(filter.Offset)
    }

    err := query.Order("created_at DESC").Find(&comments).Error
    return comments, total, err
}

// Count implements domain.CommentRepository
func (r *CommentRepository) Count(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.Comment{}).Count(&count).Error
    return count, err
}

// GetByID implements domain.CommentRepository
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
    var comment domain.Comment
    err := r.db.WithContext(ctx).
        Preload("Author").
        Where("id = ?", id).
        First(&comment).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &comment, err
}

// Update implements domain.CommentRepository
func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
    return r.db.WithContext(ctx).Save(comment).Error
}

// ListByAuthor implements domain.CommentRepository
func (r *CommentRepository) ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Comment, int64, error) {
    var comments []domain.Comment
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.Comment{}).Where("author_id = ?", authorID)

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }

    err := query.Preload("Author").Order("created_at DESC").Find(&comments).Error
    return comments, total, err
}

// ListByPost implements domain.CommentRepository
func (r *CommentRepository) ListByPost(ctx context.Context, postID uuid.UUID) ([]domain.Comment, error) {
    var comments []domain.Comment
    err := r.db.WithContext(ctx).
        Where("post_id = ?", postID).
        Preload("Author").
        Order("created_at ASC").
        Find(&comments).Error
    return comments, err
}