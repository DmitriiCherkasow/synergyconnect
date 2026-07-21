package http

import (
	"github.com/gin-gonic/gin"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/handlers"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/middleware"
	"github.com/DmitriiCherkasow/synergyconnect.git/pkg/jwt"
)

// SetupRoutes настраивает маршруты
func SetupRoutes(
	r *gin.Engine,
	authHandler *handlers.AuthHandler,
	postHandler *handlers.PostHandler,
	commentHandler *handlers.CommentHandler,
	groupHandler *handlers.GroupHandler,
	boardHandler *handlers.BoardHandler,
	stickerHandler *handlers.StickerHandler,
	reminderHandler *handlers.ReminderHandler, // ← ДОБАВИТЬ
	jwtService *jwt.JWTService,
) {
	// Health-check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "synergyconnect",
			"version": "0.3.0",
		})
	})

	api := r.Group("/api/v1")
	{
		// Публичные эндпоинты
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Публичные посты и группы (без авторизации)
		api.GET("/posts/public", postHandler.GetPublicFeed)
		api.GET("/groups/tree", groupHandler.GetGroupTree)

		// Защищенные эндпоинты
		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(jwtService))
		{
			// ============================================
			// Посты
			// ============================================
			protected.POST("/posts", postHandler.CreatePost)
			protected.GET("/posts/:id", postHandler.GetPost)
			protected.PUT("/posts/:id", postHandler.UpdatePost)
			protected.DELETE("/posts/:id", postHandler.DeletePost)
			protected.GET("/posts/feed", postHandler.GetFeed)

			// ============================================
			// Комментарии
			// ============================================
			protected.POST("/posts/:postId/comments", commentHandler.AddComment)
			protected.DELETE("/comments/:id", commentHandler.DeleteComment)

			// ============================================
			// Группы и подписки
			// ============================================
			protected.POST("/groups/:id/subscribe", groupHandler.SubscribeToGroup)
			protected.DELETE("/groups/:id/unsubscribe", groupHandler.UnsubscribeFromGroup)

			// ============================================
			// Доски (Boards)
			// ============================================
			protected.POST("/boards", boardHandler.CreateBoard)
			protected.GET("/boards", boardHandler.GetUserBoards)
			protected.GET("/boards/:id", boardHandler.GetBoard)
			protected.PUT("/boards/:id", boardHandler.UpdateBoard)
			protected.DELETE("/boards/:id", boardHandler.DeleteBoard)
			protected.PATCH("/boards/:id/archive", boardHandler.ArchiveBoard)
			protected.PATCH("/boards/:id/unarchive", boardHandler.UnarchiveBoard)

			// ============================================
			// Стикеры (Stickers)
			// ============================================
			protected.POST("/boards/:boardId/stickers", stickerHandler.CreateSticker)
			protected.GET("/stickers/:id", stickerHandler.GetSticker)
			protected.PUT("/stickers/:id", stickerHandler.UpdateSticker)
			protected.DELETE("/stickers/:id", stickerHandler.DeleteSticker)
			protected.PATCH("/stickers/:id/toggle-complete", stickerHandler.ToggleComplete)
			protected.PATCH("/stickers/:id/position", stickerHandler.UpdatePosition)

			// ============================================
			// Напоминания (Reminders)
			// ============================================
			protected.POST("/stickers/:stickerId/reminders", reminderHandler.CreateReminder)
			protected.GET("/reminders", reminderHandler.GetUserReminders)
			protected.DELETE("/reminders/:id", reminderHandler.DeleteReminder)
			protected.PATCH("/reminders/:id/snooze", reminderHandler.SnoozeReminder)

			// ============================================
			// Профиль
			// ============================================
			protected.GET("/profile", func(c *gin.Context) {
				userID := middleware.GetUserIDFromContext(c)
				c.JSON(200, gin.H{
					"message": "Authenticated access",
					"user_id": userID,
				})
			})
		}
	}
}