package domain

import (
	"github.com/google/uuid"
)

// Tag — модель тега
type Tag struct {
	ID   uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name string    `json:"name" gorm:"size:50;unique;not null"`
	Slug string    `json:"slug" gorm:"size:50;unique;not null"`

	// Связь Many-to-Many с постами
	Posts []Post `json:"posts,omitempty" gorm:"many2many:post_tags;"`
}