package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// PostRepository — интерфейс для работы с постами
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]domain.Post, error)
	FindByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Post, error)
	FindPublicFeed(ctx context.Context, tagSlug string, limit, offset int) ([]domain.Post, error)
	FindFeedBySubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddTag(ctx context.Context, postID, tagID uuid.UUID) error
	RemoveTag(ctx context.Context, postID, tagID uuid.UUID) error
}

// CommentRepository — интерфейс для работы с комментариями
type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]domain.Comment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
}

// TagRepository — интерфейс для работы с тегами
type TagRepository interface {
	FindBySlug(ctx context.Context, slug string) (*domain.Tag, error)
	FindOrCreate(ctx context.Context, name string) (*domain.Tag, error)
}

// PostService — сервис для работы с постами
type PostService struct {
	postRepo    PostRepository
	commentRepo CommentRepository
	tagRepo     TagRepository
}

// NewPostService создает новый сервис
func NewPostService(
	postRepo PostRepository,
	commentRepo CommentRepository,
	tagRepo TagRepository,
) *PostService {
	return &PostService{
		postRepo:    postRepo,
		commentRepo: commentRepo,
		tagRepo:     tagRepo,
	}
}

// CreatePostRequest — запрос на создание поста
type CreatePostRequest struct {
	AuthorID   uuid.UUID
	GroupID    *uuid.UUID
	Title      string
	Content    string
	Category   domain.PostCategory
	Visibility domain.PostVisibility
	TagNames   []string
}

// CreatePost создает новый пост
func (s *PostService) CreatePost(ctx context.Context, req CreatePostRequest) (*domain.Post, error) {
	// Создаем пост
	post := &domain.Post{
		AuthorID:   req.AuthorID,
		GroupID:    req.GroupID,
		Title:      req.Title,
		Content:    req.Content,
		Category:   req.Category,
		Visibility: req.Visibility,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	// Добавляем теги
	for _, tagName := range req.TagNames {
		tag, err := s.tagRepo.FindOrCreate(ctx, tagName)
		if err != nil {
			return nil, err
		}
		if err := s.postRepo.AddTag(ctx, post.ID, tag.ID); err != nil {
			return nil, err
		}
	}

	return post, nil
}

// GetPost возвращает пост по ID с проверкой прав
func (s *PostService) GetPost(ctx context.Context, postID, userID uuid.UUID, userRole domain.UserRole, userGroupIDs []uuid.UUID) (*domain.Post, error) {
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	// Проверяем права на просмотр
	if !post.CanView(userID, userRole, userGroupIDs) {
		return nil, errors.New("access denied")
	}

	return post, nil
}

// UpdatePostRequest — запрос на обновление поста
type UpdatePostRequest struct {
	PostID     uuid.UUID
	UserID     uuid.UUID
	UserRole   domain.UserRole
	Title      *string
	Content    *string
	Category   *domain.PostCategory
	Visibility *domain.PostVisibility
}

// UpdatePost обновляет пост
func (s *PostService) UpdatePost(ctx context.Context, req UpdatePostRequest) (*domain.Post, error) {
	post, err := s.postRepo.FindByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	// Проверяем права
	if !post.CanEdit(req.UserID, req.UserRole) {
		return nil, errors.New("access denied")
	}

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.Category != nil {
		post.Category = *req.Category
	}
	if req.Visibility != nil {
		post.Visibility = *req.Visibility
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// DeletePost удаляет пост
func (s *PostService) DeletePost(ctx context.Context, postID, userID uuid.UUID, userRole domain.UserRole) error {
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}

	if !post.CanDelete(userID, userRole) {
		return errors.New("access denied")
	}

	return s.postRepo.Delete(ctx, postID)
}

// GetPublicFeed возвращает публичную ленту
func (s *PostService) GetPublicFeed(ctx context.Context, tagSlug string, limit, offset int) ([]domain.Post, error) {
	return s.postRepo.FindPublicFeed(ctx, tagSlug, limit, offset)
}

// GetFeedBySubscriptions возвращает ленту подписок
func (s *PostService) GetFeedBySubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	return s.postRepo.FindFeedBySubscriptions(ctx, userID, limit, offset)
}

// AddComment добавляет комментарий к посту
func (s *PostService) AddComment(ctx context.Context, postID, authorID uuid.UUID, content string, parentID *uuid.UUID) (*domain.Comment, error) {
	// Проверяем, существует ли пост
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	comment := &domain.Comment{
		PostID:   postID,
		AuthorID: authorID,
		ParentID: parentID,
		Content:  content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComments возвращает комментарии поста
func (s *PostService) GetComments(ctx context.Context, postID uuid.UUID) ([]domain.Comment, error) {
	return s.commentRepo.FindByPostID(ctx, postID)
}

// DeleteComment удаляет комментарий
func (s *PostService) DeleteComment(ctx context.Context, commentID, userID uuid.UUID, userRole domain.UserRole) error {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return errors.New("comment not found")
	}

	if !comment.CanDelete(userID, userRole) {
		return errors.New("access denied")
	}

	return s.commentRepo.Delete(ctx, commentID)
}