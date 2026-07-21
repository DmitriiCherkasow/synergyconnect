package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/pkg/jwt"
)

// JWTAuthMiddleware — проверяет JWT-токен и добавляет пользователя в контекст
func JWTAuthMiddleware(jwtService *jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат Bearer
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Валидируем токен
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)        // uuid.UUID
		c.Set("user_email", claims.Email)      // string
		c.Set("user_role", claims.Role)        // string

		// Полная структура пользователя
		c.Set("user", &domain.User{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  domain.UserRole(claims.Role),
		})

		c.Next()
	}
}

// RequireRole — проверяет, что пользователь имеет одну из указанных ролей
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		currentUser := user.(*domain.User)

		for _, role := range allowedRoles {
			if currentUser.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

// GetUserIDFromContext — возвращает user_id из контекста как UUID
func GetUserIDFromContext(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	// user_id сохранён как uuid.UUID, преобразуем в строку
	if uid, ok := userID.(uuid.UUID); ok {
		return uid.String()
	}

	return ""
}

// GetUserIDFromContextAsUUID — возвращает user_id из контекста как uuid.UUID
func GetUserIDFromContextAsUUID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, nil
	}

	if uid, ok := userID.(uuid.UUID); ok {
		return uid, nil
	}

	return uuid.Nil, nil
}