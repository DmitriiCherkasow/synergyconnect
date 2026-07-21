package domain

import (
	"time"
	"github.com/google/uuid"
)

// GroupType — тип группы
type GroupType string

const (
	GroupTypeUniversity GroupType = "university"
	GroupTypeFaculty    GroupType = "faculty"
	GroupTypeDepartment GroupType = "department"
	GroupTypeCourse     GroupType = "course"
	GroupTypeStudyGroup GroupType = "study_group"
)

// Group — модель учебной группы
type Group struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string     `json:"name" gorm:"size:255;not null"`
	Slug        string     `json:"slug" gorm:"size:255;unique;not null"`
	Type        GroupType  `json:"type" gorm:"not null"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"`
	Description string     `json:"description" gorm:"type:text"`
	ColorHex    string     `json:"color_hex" gorm:"size:7;default:'#3498db'"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Связи
	Parent   *Group   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Group  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// IsRoot проверяет, является ли группа корневой (университет)
func (g *Group) IsRoot() bool {
	return g.ParentID == nil && g.Type == GroupTypeUniversity
}

// FullPath возвращает полный путь к группе (для отображения)
func (g *Group) FullPath(parentNames ...string) string {
	// Будет реализовано в сервисе
	return g.Name
}