package domain

import (
	"time"
	"github.com/google/uuid"
)

// UserRole определяет роль пользователя в системе
type UserRole string

const (
	RoleStudent  UserRole = "student"    // Студент
	RoleMentor   UserRole = "mentor"     // Ментор/Преподаватель
	RoleEmployer UserRole = "employer"   // Работодатель
	RoleAdmin    UserRole = "admin"      // Администратор
	RoleSuperAdmin UserRole = "super_admin" // Старший Администратор
)

// User - основная бизнес-сущность пользователя
type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"not null"`
	Role         UserRole   `json:"role" gorm:"not null;default:'student'"`
	FirstName    string     `json:"first_name" gorm:"size:100"`
	LastName     string     `json:"last_name" gorm:"size:100"`
	AvatarURL    string     `json:"avatar_url,omitempty" gorm:"size:500"`
	Bio          string     `json:"bio,omitempty" gorm:"type:text"`
	IsVerified   bool       `json:"is_verified" gorm:"default:false"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// FullName возвращает полное имя пользователя
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return u.FirstName + " " + u.LastName
}

// IsStudent проверяет, является ли пользователь студентом
func (u *User) IsStudent() bool {
	return u.Role == RoleStudent
}

// IsMentor проверяет, является ли пользователь ментором
func (u *User) IsMentor() bool {
	return u.Role == RoleMentor
}

// IsEmployer проверяет, является ли пользователь работодателем
func (u *User) IsEmployer() bool {
	return u.Role == RoleEmployer
}

// IsAdmin проверяет, является ли пользователь администратором (включая super_admin)
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// IsSuperAdmin проверяет, является ли пользователь суперадминистратором
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// HasRole проверяет, имеет ли пользователь указанную роль
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

// CanManagePosts проверяет, может ли пользователь управлять постами (ментор, админ или суперадмин)
func (u *User) CanManagePosts() bool {
	return u.Role == RoleMentor || u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// CanManageUsers проверяет, может ли пользователь управлять другими пользователями (админ или суперадмин)
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// CanAssignAdmin проверяет, может ли пользователь назначать администраторов (только суперадмин)
func (u *User) CanAssignAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// CanPostJobs проверяет, может ли пользователь публиковать вакансии (работодатель, админ или суперадмин)
func (u *User) CanPostJobs() bool {
	return u.Role == RoleEmployer || u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// IsPublicProfile проверяет, виден ли профиль публично
func (u *User) IsPublicProfile() bool {
	return u.IsActive && u.IsVerified
}