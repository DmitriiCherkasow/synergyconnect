package domain

import (
	"time"
	"github.com/google/uuid"
)

// StickerPriority — приоритет стикера
type StickerPriority string

const (
	PriorityLow    StickerPriority = "low"
	PriorityMedium StickerPriority = "medium"
	PriorityHigh   StickerPriority = "high"
	PriorityUrgent StickerPriority = "urgent"
)

// StickerColor — цвет стикера
type StickerColor string

const (
	ColorYellow  StickerColor = "yellow"
	ColorBlue    StickerColor = "blue"
	ColorGreen   StickerColor = "green"
	ColorRed     StickerColor = "red"
	ColorPurple  StickerColor = "purple"
	ColorOrange  StickerColor = "orange"
	ColorGray    StickerColor = "gray"
)

// Sticker — модель стикера на доске
type Sticker struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BoardID     uuid.UUID       `json:"board_id" gorm:"type:uuid;not null"`
	AuthorID    uuid.UUID       `json:"author_id" gorm:"type:uuid;not null"`
	Title       string          `json:"title" gorm:"size:255"`
	Content     string          `json:"content" gorm:"type:text;not null"`
	Color       StickerColor    `json:"color" gorm:"default:'yellow'"`
	Priority    StickerPriority `json:"priority" gorm:"default:'medium'"`
	PositionX   int             `json:"position_x" gorm:"default:0"`
	PositionY   int             `json:"position_y" gorm:"default:0"`
	Width       int             `json:"width" gorm:"default:200"`
	Height      int             `json:"height" gorm:"default:150"`
	DueDate     *time.Time      `json:"due_date,omitempty"`
	IsCompleted bool            `json:"is_completed" gorm:"default:false"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	// Связи
	Board  Board `json:"board" gorm:"foreignKey:BoardID"`
	Author User  `json:"author" gorm:"foreignKey:AuthorID"`
}

// IsAuthor проверяет, является ли пользователь автором стикера
func (s *Sticker) IsAuthor(userID uuid.UUID) bool {
	return s.AuthorID == userID
}

// CanEdit проверяет, может ли пользователь редактировать стикер
func (s *Sticker) CanEdit(userID uuid.UUID, userRole UserRole) bool {
	return s.AuthorID == userID || userRole == RoleAdmin
}

// IsOverdue проверяет, просрочен ли стикер
func (s *Sticker) IsOverdue() bool {
	if s.DueDate == nil || s.IsCompleted {
		return false
	}
	return time.Now().After(*s.DueDate)
}

// DaysUntilDue возвращает количество дней до дедлайна
func (s *Sticker) DaysUntilDue() int {
	if s.DueDate == nil || s.IsCompleted {
		return -1
	}
	days := int(time.Until(*s.DueDate).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}