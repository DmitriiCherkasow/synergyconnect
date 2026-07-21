package domain

import (
	"time"
	"github.com/google/uuid"
)

// Board — модель рабочей доски
type Board struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerID     uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	GroupID     *uuid.UUID `json:"group_id,omitempty" gorm:"type:uuid"`
	Title       string     `json:"title" gorm:"size:255;not null"`
	Description string     `json:"description" gorm:"type:text"`
	ColorHex    string     `json:"color_hex" gorm:"size:7;default:'#3498db'"`
	IsArchived  bool       `json:"is_archived" gorm:"default:false"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Связи
	Owner User  `json:"owner" gorm:"foreignKey:OwnerID"`
	Group *Group `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	Stickers []Sticker `json:"stickers,omitempty" gorm:"foreignKey:BoardID"`
}

// IsOwner проверяет, является ли пользователь владельцем доски
func (b *Board) IsOwner(userID uuid.UUID) bool {
	return b.OwnerID == userID
}

// CanView проверяет, может ли пользователь просматривать доску
func (b *Board) CanView(userID uuid.UUID, userRole UserRole, userGroupIDs []uuid.UUID) bool {
	// Владелец всегда может видеть
	if b.OwnerID == userID {
		return true
	}
	// Админ всегда может видеть
	if userRole == RoleAdmin {
		return true
	}
	// Если доска групповая - проверяем членство
	if b.GroupID != nil {
		for _, gid := range userGroupIDs {
			if gid == *b.GroupID {
				return true
			}
		}
	}
	return false
}

// CanEdit проверяет, может ли пользователь редактировать доску
func (b *Board) CanEdit(userID uuid.UUID, userRole UserRole) bool {
	return b.OwnerID == userID || userRole == RoleAdmin
}