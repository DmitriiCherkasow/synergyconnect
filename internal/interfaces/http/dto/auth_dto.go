package dto

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// RegisterRequest — запрос на регистрацию
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Validate проверяет корректность запроса
func (r *RegisterRequest) Validate() error {
	return validate.Struct(r)
}

// LoginRequest — запрос на вход
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Validate проверяет корректность запроса
func (r *LoginRequest) Validate() error {
	return validate.Struct(r)
}

// RefreshTokenRequest — запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse — ответ при успешной аутентификации
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

// UserResponse — информация о пользователе
type UserResponse struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	IsVerified bool   `json:"is_verified"`
}