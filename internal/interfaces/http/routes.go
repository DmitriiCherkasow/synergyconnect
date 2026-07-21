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
	jwtService *jwt.JWTService,
) {
	// Health-check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "synergyconnect",
			"version": "0.2.0",
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
			// Посты
			protected.POST("/posts", postHandler.CreatePost)
			protected.GET("/posts/:id", postHandler.GetPost)
			protected.PUT("/posts/:id", postHandler.UpdatePost)
			protected.DELETE("/posts/:id", postHandler.DeletePost)
			protected.GET("/posts/feed", postHandler.GetFeed)

			// Комментарии
			protected.POST("/posts/:postId/comments", commentHandler.AddComment)
			protected.DELETE("/comments/:id", commentHandler.DeleteComment)

			// Группы
			protected.POST("/groups/:id/subscribe", groupHandler.SubscribeToGroup)
			protected.DELETE("/groups/:id/unsubscribe", groupHandler.UnsubscribeFromGroup)

			// Профиль
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