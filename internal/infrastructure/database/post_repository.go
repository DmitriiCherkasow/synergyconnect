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

// List implements domain.PostRepository
func (r *PostRepository) List(ctx context.Context, filter domain.PostFilter) ([]domain.Post, int64, error) {
    var posts []domain.Post
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.Post{}).Preload("Author").Preload("Group")

    if filter.AuthorID != nil {
        query = query.Where("author_id = ?", *filter.AuthorID)
    }
    if filter.GroupID != nil {
        query = query.Where("group_id = ?", *filter.GroupID)
    }
    if filter.Category != nil {
        query = query.Where("category = ?", *filter.Category)
    }
    if filter.Visibility != nil {
        query = query.Where("visibility = ?", *filter.Visibility)
    }
    if filter.Search != "" {
        query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
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

    err := query.Order("created_at DESC").Find(&posts).Error
    return posts, total, err
}

// Count implements domain.PostRepository
func (r *PostRepository) Count(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.Post{}).Count(&count).Error
    return count, err
}

// GetByID implements domain.PostRepository
func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
    var post domain.Post
    err := r.db.WithContext(ctx).
        Preload("Author").
        Preload("Group").
        Where("id = ?", id).
        First(&post).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &post, err
}

// ListByAuthor implements domain.PostRepository
func (r *PostRepository) ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Post, int64, error) {
    var posts []domain.Post
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.Post{}).Where("author_id = ?", authorID)

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }

    err := query.Preload("Author").Preload("Group").Order("created_at DESC").Find(&posts).Error
    return posts, total, err
}

// ListByGroup implements domain.PostRepository
func (r *PostRepository) ListByGroup(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]domain.Post, int64, error) {
    var posts []domain.Post
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.Post{}).Where("group_id = ?", groupID)

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }

    err := query.Preload("Author").Preload("Group").Order("created_at DESC").Find(&posts).Error
    return posts, total, err
}

// GetPublicFeed implements domain.PostRepository
func (r *PostRepository) GetPublicFeed(ctx context.Context, limit, offset int) ([]domain.Post, int64, error) {
    var posts []domain.Post
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.Post{}).Where("visibility = ?", domain.VisibilityPublic)

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }

    err := query.Preload("Author").Preload("Group").Order("created_at DESC").Find(&posts).Error
    return posts, total, err
}

// GetFeed implements domain.PostRepository
func (r *PostRepository) GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Post, int64, error) {
    var posts []domain.Post
    var total int64

    // Получаем ID групп, на которые подписан пользователь
    var groupIDs []uuid.UUID
    r.db.WithContext(ctx).Model(&domain.Subscription{}).
        Where("user_id = ? AND subscribable_type = ?", userID, "group").
        Pluck("subscribable_id", &groupIDs)

    query := r.db.WithContext(ctx).Model(&domain.Post{}).
        Where("visibility = ? OR author_id = ? OR group_id IN ?",
            domain.VisibilityPublic,
            userID,
            groupIDs,
        )

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }

    err := query.Preload("Author").Preload("Group").Order("created_at DESC").Find(&posts).Error
    return posts, total, err
}