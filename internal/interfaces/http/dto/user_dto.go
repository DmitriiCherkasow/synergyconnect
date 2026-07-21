package dto

import (
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// ToUserResponse преобразует доменную модель User в UserResponse
func ToUserResponse(user *domain.User) UserResponse {
	if user == nil {
		return UserResponse{}
	}
	return UserResponse{
		ID:         user.ID.String(),
		Email:      user.Email,
		Role:       string(user.Role),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		AvatarURL:  user.AvatarURL,
		IsVerified: user.IsVerified,
	}
}