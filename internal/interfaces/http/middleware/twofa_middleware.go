package middleware

import (
	"net/http"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TwoFAMiddleware проверяет, что 2FA пройдена
func TwoFAMiddleware(twofaService *application.TwoFAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, есть ли флаг в контексте, что 2FA уже пройдена
		if _, ok := c.Get("2fa_verified"); ok {
			c.Next()
			return
		}

		userIDStr := GetUserIDFromContext(c)
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			c.Abort()
			return
		}

		// Проверяем, включена ли 2FA для пользователя
		enabled, err := twofaService.IsEnabled(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check 2FA status"})
			c.Abort()
			return
		}

		// Если 2FA не включена, пропускаем
		if !enabled {
			c.Next()
			return
		}

		// Проверяем, есть ли в запросе заголовок X-2FA-Code
		code := c.GetHeader("X-2FA-Code")
		if code == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "2FA required",
				"code":  "2FA_REQUIRED",
			})
			c.Abort()
			return
		}

		// Проверяем код
		valid, err := twofaService.VerifyCode(c.Request.Context(), userID, code)
		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid 2FA code",
				"code":  "INVALID_2FA_CODE",
			})
			c.Abort()
			return
		}

		// Помечаем, что 2FA пройдена
		c.Set("2fa_verified", true)
		c.Next()
	}
}

// Require2FA проверяет, что 2FA включена (для эндпоинтов, требующих 2FA)
func Require2FA(twofaService *application.TwoFAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := GetUserIDFromContext(c)
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			c.Abort()
			return
		}

		enabled, err := twofaService.IsEnabled(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check 2FA status"})
			c.Abort()
			return
		}

		if !enabled {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "2FA is required for this action",
				"code":  "2FA_REQUIRED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}