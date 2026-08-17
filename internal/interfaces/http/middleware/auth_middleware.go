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
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

        claims, err := jwtService.ValidateAccessToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_role", claims.Role)

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

// RequireAdmin — проверяет, что пользователь является администратором (включая super_admin)
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        currentUser := user.(*domain.User)
        if !currentUser.IsAdmin() {
            c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireSuperAdmin — проверяет, что пользователь является суперадмином
func RequireSuperAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        currentUser := user.(*domain.User)
        if !currentUser.IsSuperAdmin() {
            c.JSON(http.StatusForbidden, gin.H{"error": "super admin access required"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// GetUserIDFromContext — возвращает user_id из контекста как строку
func GetUserIDFromContext(c *gin.Context) string {
    userID, exists := c.Get("user_id")
    if !exists {
        return ""
    }
    if uid, ok := userID.(uuid.UUID); ok {
        return uid.String()
    }
    return ""
}

// GetUserFromContext — возвращает пользователя из контекста
func GetUserFromContext(c *gin.Context) (*domain.User, bool) {
    user, exists := c.Get("user")
    if !exists {
        return nil, false
    }
    return user.(*domain.User), true
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