package dto

import (
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// UserListResponse — список пользователей
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
}

// ToUserResponse конвертирует доменную модель в DTO
func ToUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:         user.ID.String(),
		Email:      user.Email,
		Role:       string(user.Role),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		AvatarURL:  user.AvatarURL,
		IsVerified: user.IsVerified,
		IsActive:   user.IsActive,
		Bio:        user.Bio,
	}
}