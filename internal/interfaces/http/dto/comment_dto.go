package dto

import (
	"time"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// CommentResponse — ответ с комментарием
type CommentResponse struct {
	ID        string       `json:"id"`
	PostID    string       `json:"post_id"`
	AuthorID  string       `json:"author_id"`
	Author    UserResponse `json:"author,omitempty"`
	ParentID  *string      `json:"parent_id,omitempty"`
	Content   string       `json:"content"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// CommentListResponse — список комментариев
type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int64             `json:"total"`
}

// ToCommentResponse конвертирует доменную модель в DTO
func ToCommentResponse(comment *domain.Comment) CommentResponse {
	resp := CommentResponse{
		ID:        comment.ID.String(),
		PostID:    comment.PostID.String(),
		AuthorID:  comment.AuthorID.String(),
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	if comment.ParentID != nil {
		parentID := comment.ParentID.String()
		resp.ParentID = &parentID
	}

	if comment.Author.ID != uuid.Nil {
		resp.Author = ToUserResponse(&comment.Author)
	}

	return resp
}