package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// PostRepository — репозиторий для работы с постами
type PostRepository struct {
	db *gorm.DB
}

// NewPostRepository создает новый репозиторий
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create создает новый пост
func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// FindByID ищет пост по ID
func (r *PostRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	var post domain.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Group").
		First(&post, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &post, err
}

// FindByGroupID возвращает посты группы с пагинацией
func (r *PostRepository) FindByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND visibility != ?", groupID, domain.VisibilityPrivate).
		Preload("Author").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

// FindByAuthorID возвращает посты автора
func (r *PostRepository) FindByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post
	err := r.db.WithContext(ctx).
		Where("author_id = ?", authorID).
		Preload("Author").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

// FindPublicFeed возвращает публичную ленту (с фильтром по тегам)
func (r *PostRepository) FindPublicFeed(ctx context.Context, tagSlug string, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post
	query := r.db.WithContext(ctx).
		Where("visibility = ?", domain.VisibilityPublic).
		Preload("Author").
		Preload("Tags")

	if tagSlug != "" {
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.slug = ?", tagSlug)
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

// FindFeedBySubscriptions возвращает ленту по подпискам
func (r *PostRepository) FindFeedBySubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post

	// Получаем ID пользователей, на которых подписан текущий пользователь
	var subscribedUserIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("subscriptions").
		Select("target_user_id").
		Where("subscriber_id = ? AND type = ?", userID, domain.SubscriptionUser).
		Where("target_user_id IS NOT NULL").
		Find(&subscribedUserIDs).Error
	if err != nil {
		return nil, err
	}

	// Получаем ID групп, на которые подписан текущий пользователь
	var subscribedGroupIDs []uuid.UUID
	err = r.db.WithContext(ctx).
		Table("subscriptions").
		Select("target_group_id").
		Where("subscriber_id = ? AND type = ?", userID, domain.SubscriptionGroup).
		Where("target_group_id IS NOT NULL").
		Find(&subscribedGroupIDs).Error
	if err != nil {
		return nil, err
	}

	// Если нет подписок — возвращаем пустой результат
	if len(subscribedUserIDs) == 0 && len(subscribedGroupIDs) == 0 {
		return []domain.Post{}, nil
	}

	// Строим запрос
	query := r.db.WithContext(ctx).
		Where("visibility = ?", domain.VisibilityPublic).
		Preload("Author").
		Preload("Group").
		Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	// Добавляем условия для подписок
	if len(subscribedUserIDs) > 0 {
		query = query.Or("author_id IN ?", subscribedUserIDs)
	}
	if len(subscribedGroupIDs) > 0 {
		query = query.Or("group_id IN ?", subscribedGroupIDs)
	}

	err = query.Find(&posts).Error
	return posts, err
}

// Update обновляет пост
func (r *PostRepository) Update(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// Delete удаляет пост
func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Post{}, "id = ?", id).Error
}

// AddTag добавляет тег к посту
func (r *PostRepository) AddTag(ctx context.Context, postID, tagID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		postID, tagID,
	).Error
}

// RemoveTag удаляет тег из поста
func (r *PostRepository) RemoveTag(ctx context.Context, postID, tagID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM post_tags WHERE post_id = ? AND tag_id = ?",
		postID, tagID,
	).Error
}