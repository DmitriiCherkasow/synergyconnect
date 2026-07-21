package domain

import (
	"time"
	"github.com/google/uuid"
)

// PostVisibility — тип видимости поста
type PostVisibility string

const (
	VisibilityPublic  PostVisibility = "public"   // Виден всем
	VisibilityGroup   PostVisibility = "group"    // Виден только участникам группы
	VisibilityPrivate PostVisibility = "private"  // Только автору
)

// PostCategory — категория поста
type PostCategory string

const (
	CategoryAnnouncement PostCategory = "announcement" // Объявление
	CategoryProject      PostCategory = "project"      // Проект
	CategoryQuestion     PostCategory = "question"     // Вопрос
	CategoryDiscussion   PostCategory = "discussion"   // Обсуждение
	CategoryEvent        PostCategory = "event"        // Мероприятие
	CategoryJob          PostCategory = "job"          // Вакансия
)

// Post — модель поста
type Post struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AuthorID   uuid.UUID      `json:"author_id" gorm:"type:uuid;not null"`
	GroupID    *uuid.UUID     `json:"group_id,omitempty" gorm:"type:uuid"`
	Title      string         `json:"title" gorm:"size:255;not null"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	Category   PostCategory   `json:"category" gorm:"default:'discussion'"`
	Visibility PostVisibility `json:"visibility" gorm:"default:'public'"`
	IsPinned   bool           `json:"is_pinned" gorm:"default:false"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Связи (для GORM)
	Author User `json:"author" gorm:"foreignKey:AuthorID"`
	Group  *Group `json:"group,omitempty" gorm:"foreignKey:GroupID"`
}

// IsAuthor проверяет, является ли пользователь автором поста
func (p *Post) IsAuthor(userID uuid.UUID) bool {
	return p.AuthorID == userID
}

// CanView проверяет, может ли пользователь просматривать пост
func (p *Post) CanView(userID uuid.UUID, userRole UserRole, userGroupIDs []uuid.UUID) bool {
	switch p.Visibility {
	case VisibilityPublic:
		return true
	case VisibilityPrivate:
		return p.AuthorID == userID
	case VisibilityGroup:
		if p.GroupID == nil {
			return false
		}
		// Проверяем, состоит ли пользователь в этой группе
		for _, gid := range userGroupIDs {
			if gid == *p.GroupID {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// CanEdit проверяет, может ли пользователь редактировать пост
func (p *Post) CanEdit(userID uuid.UUID, userRole UserRole) bool {
	// Автор или администратор
	return p.AuthorID == userID || userRole == RoleAdmin
}

// CanDelete проверяет, может ли пользователь удалять пост
func (p *Post) CanDelete(userID uuid.UUID, userRole UserRole) bool {
	// Автор, ментор (для группы) или администратор
	return p.AuthorID == userID || userRole == RoleAdmin || userRole == RoleMentor
}