package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/infrastructure/database"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/infrastructure/email"
	ws "github.com/DmitriiCherkasow/synergyconnect.git/internal/infrastructure/websocket"
	httpRoutes "github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/handlers"
	wsHandlers "github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/websocket"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/worker"
	"github.com/DmitriiCherkasow/synergyconnect.git/pkg/jwt"
	"github.com/DmitriiCherkasow/synergyconnect.git/pkg/totp"

	//_ "github.com/DmitriiCherkasow/synergyconnect.git/docs" // Swagger docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           SynergyConnect API
// @version         1.0.0
// @description     Университетская социальная сеть с полным набором функций
// @termsOfService  https://synergyconnect.com/terms

// @contact.name   API Support
// @contact.email  support@synergyconnect.com

// @license.name   MIT
// @license.url    https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	log.Println("🚀 SynergyConnect starting...")

	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, using environment variables")
	}

	// Подключаемся к БД
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=synergy_user password=synergy_password dbname=synergy_db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected")

	// Миграции
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Post{},
		&domain.Comment{},
		&domain.Tag{},
		&domain.Group{},
		&domain.Subscription{},
		&domain.Board{},
		&domain.Sticker{},
		&domain.Reminder{},
		&domain.ReminderEmail{},
		&domain.Project{},
		&domain.ProjectMember{},
		&domain.ProjectApplication{},
		&domain.Vacancy{},
		&domain.VacancyResponse{},
		&domain.Message{},
		&domain.Notification{},
		&domain.UserTwoFA{},
		&domain.UserDevice{},
	); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migrated")

	// Настраиваем JWT
	accessExpiration, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRATION", "24h"))
	if err != nil {
		accessExpiration = 24 * time.Hour
	}
	refreshExpiration, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRATION", "720h"))
	if err != nil {
		refreshExpiration = 720 * time.Hour
	}

	jwtConfig := jwt.Config{
		SecretKey:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		AccessExpiration:   accessExpiration,
		RefreshExpiration:  refreshExpiration,
	}
	jwtService := jwt.NewJWTService(jwtConfig)

	// ============================================================
	// ИНИЦИАЛИЗАЦИЯ РЕПОЗИТОРИЕВ
	// ============================================================
	userRepo := database.NewUserRepository(db)
	postRepo := database.NewPostRepository(db)
	commentRepo := database.NewCommentRepository(db)
	groupRepo := database.NewGroupRepository(db)
	subscriptionRepo := database.NewSubscriptionRepository(db)
	tagRepo := database.NewTagRepository(db)

	boardRepo := database.NewBoardRepository(db)
	stickerRepo := database.NewStickerRepository(db)
	reminderRepo := database.NewReminderRepository(db)
	reminderEmailRepo := database.NewReminderEmailRepository(db)

	projectRepo := database.NewProjectRepository(db)
	vacancyRepo := database.NewVacancyRepository(db)
	messageRepo := database.NewMessageRepository(db)
	notificationRepo := database.NewNotificationRepository(db)
	twofaRepo := database.NewTwoFARepository(db)

	// ============================================================
	// TOTP ПРОВАЙДЕР
	// ============================================================
	totpProvider := totp.NewTOTPProvider("SynergyConnect")

	// ============================================================
	// ИНИЦИАЛИЗАЦИЯ СЕРВИСОВ
	// ============================================================
	authService := application.NewAuthService(userRepo, jwtService)
	postService := application.NewPostService(postRepo, commentRepo, tagRepo)
	groupService := application.NewGroupService(groupRepo, subscriptionRepo)

	boardService := application.NewBoardService(boardRepo, stickerRepo, reminderRepo)
	stickerService := application.NewStickerService(stickerRepo, boardRepo, reminderRepo)
	reminderService := application.NewReminderService(reminderRepo, stickerRepo)

	projectService := application.NewProjectService(projectRepo)
	vacancyService := application.NewVacancyService(vacancyRepo)
	chatService := application.NewChatService(messageRepo, userRepo)
	notificationService := application.NewNotificationService(notificationRepo, userRepo)
	twofaService := application.NewTwoFAService(twofaRepo, userRepo, totpProvider)

	// ============================================================
	// АДМИН СЕРВИС
	// ============================================================
	adminService := application.NewAdminService(
		userRepo,
		postRepo,
		commentRepo,
		projectRepo,
		vacancyRepo,
	)

	// ============================================================
	// WebSocket HUB
	// ============================================================
	wsHub := ws.NewHub()
	go wsHub.Run()

	
	// ============================================================
	// ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ
	// ============================================================
	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService, groupService)
	commentHandler := handlers.NewCommentHandler(postService)
	groupHandler := handlers.NewGroupHandler(groupService)

	boardHandler := handlers.NewBoardHandler(boardService, stickerService)
	stickerHandler := handlers.NewStickerHandler(stickerService)
	reminderHandler := handlers.NewReminderHandler(reminderService)

	projectHandler := handlers.NewProjectHandler(projectService)
	vacancyHandler := handlers.NewVacancyHandler(vacancyService)
	chatHandler := handlers.NewChatHandler(chatService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	twofaHandler := handlers.NewTwoFAHandler(twofaService)
	adminHandler := handlers.NewAdminHandler(adminService)

	_ = wsHandlers.NewChatWebSocketHandler(wsHub, chatService)

	// ============================================================
	// EMAIL КОНФИГУРАЦИЯ И ВОРКЕР
	// ============================================================
	emailConfig := email.Config{
		Host:     getEnv("SMTP_HOST", "localhost"),
		Port:     1025,
		Username: getEnv("SMTP_USER", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("FROM_EMAIL", "test@synergyconnect.com"),
		FromName: "SynergyConnect",
		UseTLS:   false,
	}
	emailService := email.NewService(emailConfig)

	reminderWorker := worker.NewReminderWorker(
		reminderRepo,
		stickerRepo,
		boardRepo,
		userRepo,
		reminderEmailRepo,
		emailService,
		1*time.Minute,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go reminderWorker.Start(ctx)

	// ============================================================
	// НАСТРОЙКА РОУТЕРА
	// ============================================================
	r := gin.Default()

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Настройка маршрутов
	httpRoutes.SetupRoutes(
		r,
		authHandler,
		postHandler,
		commentHandler,
		groupHandler,
		boardHandler,
		stickerHandler,
		reminderHandler,
		projectHandler,
		vacancyHandler,
		chatHandler,
		notificationHandler,
		twofaHandler,
		adminHandler,
		jwtService,
		twofaService,
	)

	// WebSocket эндпоинт
	r.GET("/ws/chat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "WebSocket endpoint"})
	})

	// ============================================================
	// ЗАПУСК СЕРВЕРА
	// ============================================================
	port := getEnv("SERVER_PORT", "8080")
	log.Printf("✅ Server is running on http://localhost:%s", port)
	log.Printf("📚 Swagger UI: http://localhost:%s/swagger/index.html", port)
	log.Printf("📧 Reminder worker is running (checking every 1 minute)")

	defer func() {
		log.Println("🛑 Shutting down reminder worker...")
		cancel()
		time.Sleep(2 * time.Second)
		log.Println("✅ Reminder worker stopped")
	}()

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}