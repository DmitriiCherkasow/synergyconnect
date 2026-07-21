package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// CreatePostRequest — запрос на создание поста
type CreatePostRequest struct {
	GroupID    *uuid.UUID `json:"group_id,omitempty"`
	Title      string     `json:"title" binding:"required"`
	Content    string     `json:"content" binding:"required"`
	Category   string     `json:"category" binding:"required,oneof=announcement project question discussion event job"`
	Visibility string     `json:"visibility" binding:"required,oneof=public group private"`
	Tags       []string   `json:"tags"`
}

// UpdatePostRequest — запрос на обновление поста
type UpdatePostRequest struct {
	Title      *string `json:"title"`
	Content    *string `json:"content"`
	Category   *string `json:"category" binding:"omitempty,oneof=announcement project question discussion event job"`
	Visibility *string `json:"visibility" binding:"omitempty,oneof=public group private"`
}

// PostResponse — ответ с данными поста
type PostResponse struct {
	ID         string              `json:"id"`
	Author     UserResponse        `json:"author"`
	Group      *GroupResponse      `json:"group,omitempty"`
	Title      string              `json:"title"`
	Content    string              `json:"content"`
	Category   string              `json:"category"`
	Visibility string              `json:"visibility"`
	IsPinned   bool                `json:"is_pinned"`
	Tags       []TagResponse       `json:"tags"`
	Comments   []CommentResponse   `json:"comments,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// CommentResponse — ответ с данными комментария
type CommentResponse struct {
	ID        string       `json:"id"`
	Author    UserResponse `json:"author"`
	Content   string       `json:"content"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// TagResponse — ответ с данными тега
type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// GroupResponse — ответ с данными группы
type GroupResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Type        string  `json:"type"`
	ParentID    *string `json:"parent_id,omitempty"`
	Description string  `json:"description"`
	ColorHex    string  `json:"color_hex"`
	Children    []GroupResponse `json:"children,omitempty"`
}

// ToPostResponse преобразует доменную модель в ответ
func ToPostResponse(post *domain.Post, comments []domain.Comment, tags []domain.Tag) PostResponse {
	resp := PostResponse{
		ID:         post.ID.String(),
		Author:     ToUserResponse(&post.Author),
		Title:      post.Title,
		Content:    post.Content,
		Category:   string(post.Category),
		Visibility: string(post.Visibility),
		IsPinned:   post.IsPinned,
		CreatedAt:  post.CreatedAt,
		UpdatedAt:  post.UpdatedAt,
	}

	if post.Group != nil {
		groupResp := ToGroupResponse(post.Group)
		resp.Group = &groupResp
	}

	// Теги
	for _, tag := range tags {
		resp.Tags = append(resp.Tags, TagResponse{
			ID:   tag.ID.String(),
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	// Комментарии
	for _, comment := range comments {
		resp.Comments = append(resp.Comments, CommentResponse{
			ID:        comment.ID.String(),
			Author:    ToUserResponse(&comment.Author),
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	return resp
}

// ToGroupResponse преобразует доменную модель группы в ответ
func ToGroupResponse(group *domain.Group) GroupResponse {
	resp := GroupResponse{
		ID:          group.ID.String(),
		Name:        group.Name,
		Slug:        group.Slug,
		Type:        string(group.Type),
		Description: group.Description,
		ColorHex:    group.ColorHex,
	}

	if group.ParentID != nil {
		parentID := group.ParentID.String()
		resp.ParentID = &parentID
	}

	for _, child := range group.Children {
		resp.Children = append(resp.Children, ToGroupResponse(&child))
	}

	return resp
}