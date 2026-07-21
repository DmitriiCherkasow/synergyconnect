package domain

import (
	"time"
	"github.com/google/uuid"
)

// Comment — модель комментария
type Comment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;not null"`
	AuthorID  uuid.UUID `json:"author_id" gorm:"type:uuid;not null"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"` // Для вложенных комментариев
	Content   string    `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Связи
	Author User `json:"author" gorm:"foreignKey:AuthorID"`
	Post   Post `json:"post" gorm:"foreignKey:PostID"`
}

// IsAuthor проверяет, является ли пользователь автором комментария
func (c *Comment) IsAuthor(userID uuid.UUID) bool {
	return c.AuthorID == userID
}

// CanDelete проверяет, может ли пользователь удалять комментарий
func (c *Comment) CanDelete(userID uuid.UUID, userRole UserRole) bool {
	return c.AuthorID == userID || userRole == RoleAdmin || userRole == RoleMentor
}