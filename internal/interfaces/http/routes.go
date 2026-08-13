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
	reminderHandler *handlers.ReminderHandler,
	projectHandler *handlers.ProjectHandler,
	vacancyHandler *handlers.VacancyHandler,
	chatHandler *handlers.ChatHandler,
	notificationHandler *handlers.NotificationHandler,
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
			// Проекты (Projects)
			// ============================================
			protected.POST("/projects", projectHandler.CreateProject)
			protected.GET("/projects", projectHandler.ListProjects)
			protected.GET("/projects/:id", projectHandler.GetProject)
			protected.PUT("/projects/:id", projectHandler.UpdateProject)
			protected.DELETE("/projects/:id", projectHandler.DeleteProject)

			// Участники проектов
			protected.GET("/projects/:id/members", projectHandler.GetProjectMembers)
			protected.POST("/projects/:id/members", projectHandler.AddMember)
			protected.DELETE("/projects/:id/members/:userId", projectHandler.RemoveMember)

			// Заявки на вступление
			protected.POST("/projects/:id/applications", projectHandler.CreateApplication)
			protected.PUT("/projects/:id/applications/:applicationId", projectHandler.UpdateApplication)

			// ============================================
			// Вакансии (Vacancies)
			// ============================================
			protected.POST("/vacancies", vacancyHandler.CreateVacancy)
			protected.GET("/vacancies", vacancyHandler.ListVacancies)
			protected.GET("/vacancies/search", vacancyHandler.SearchVacancies)
			protected.GET("/vacancies/:id", vacancyHandler.GetVacancy)
			protected.PUT("/vacancies/:id", vacancyHandler.UpdateVacancy)
			protected.DELETE("/vacancies/:id", vacancyHandler.DeleteVacancy)

			// Отклики на вакансии
			protected.POST("/vacancies/:id/responses", vacancyHandler.CreateResponse)
			protected.PUT("/vacancies/responses/:responseId", vacancyHandler.UpdateResponse)
			protected.GET("/vacancies/responses/my", vacancyHandler.GetMyResponses)

			// ============================================
			// Чат (Chat)
			// ============================================
			protected.POST("/chat/messages", chatHandler.SendMessage)
			protected.GET("/chat/messages/:userId", chatHandler.GetConversation)
			protected.GET("/chat/unread/count", chatHandler.GetUnreadCount)
			protected.PUT("/chat/read", chatHandler.MarkAsRead)
			protected.GET("/chat/recent", chatHandler.GetRecentChats)
			protected.DELETE("/chat/messages/:id", chatHandler.DeleteMessage)

			// ============================================
			// WebSocket для чата (отдельный эндпоинт)
			// ============================================
			// WebSocket обрабатывается отдельно через gin
			// protected.GET("/ws/chat", chatWebSocketHandler.HandleWebSocket)

			// ============================================
			// Уведомления (Notifications)
			// ============================================
			protected.GET("/notifications", notificationHandler.GetNotifications)
			protected.GET("/notifications/unread/count", notificationHandler.GetUnreadCount)
			protected.PUT("/notifications/:id/read", notificationHandler.MarkAsRead)
			protected.PUT("/notifications/read/all", notificationHandler.MarkAllAsRead)
			protected.DELETE("/notifications/:id", notificationHandler.DeleteNotification)
			protected.DELETE("/notifications", notificationHandler.DeleteAllNotifications)

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