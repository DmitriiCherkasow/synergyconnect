package main

import (
	"context"
	"log"
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
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/handlers"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/worker"
	"github.com/DmitriiCherkasow/synergyconnect.git/pkg/jwt"
)

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

	// Репозитории для Спринта 2
	boardRepo := database.NewBoardRepository(db)
	stickerRepo := database.NewStickerRepository(db)
	reminderRepo := database.NewReminderRepository(db)
	reminderEmailRepo := database.NewReminderEmailRepository(db) 

	// ============================================================
	// ИНИЦИАЛИЗАЦИЯ СЕРВИСОВ
	// ============================================================
	authService := application.NewAuthService(userRepo, jwtService)
	postService := application.NewPostService(postRepo, commentRepo, tagRepo)
	groupService := application.NewGroupService(groupRepo, subscriptionRepo)

	// Сервисы для Спринта 2
	boardService := application.NewBoardService(boardRepo, stickerRepo, reminderRepo)
	stickerService := application.NewStickerService(stickerRepo, boardRepo, reminderRepo)
	reminderService := application.NewReminderService(reminderRepo, stickerRepo)

	// ============================================================
	// ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ
	// ============================================================
	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService, groupService)
	commentHandler := handlers.NewCommentHandler(postService)
	groupHandler := handlers.NewGroupHandler(groupService)

	// Обработчики для Спринта 2
	boardHandler := handlers.NewBoardHandler(boardService, stickerService)
	stickerHandler := handlers.NewStickerHandler(stickerService)
	reminderHandler := handlers.NewReminderHandler(reminderService)

	// ============================================================
	// EMAIL КОНФИГУРАЦИЯ И ВОРКЕР
	// ============================================================
	emailConfig := email.Config{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     587,
		Username: getEnv("SMTP_USER", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("FROM_EMAIL", ""),
		FromName: "SynergyConnect",
		UseTLS:   true,
	}
	emailService := email.NewService(emailConfig)

	// Инициализация воркера
	reminderWorker := worker.NewReminderWorker(
		reminderRepo,
		stickerRepo,
		boardRepo,
		userRepo,
		reminderEmailRepo,
		emailService,
		1*time.Minute,
	)

	// Запуск воркера в отдельной горутине
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go reminderWorker.Start(ctx)

	// ============================================================
	// НАСТРОЙКА РОУТЕРА И ЗАПУСК СЕРВЕРА
	// ============================================================
	r := gin.Default()

	// Настройка маршрутов
	http.SetupRoutes(
		r,
		authHandler,
		postHandler,
		commentHandler,
		groupHandler,
		boardHandler,
		stickerHandler,
		reminderHandler,
		jwtService,
	)

	// Запускаем сервер
	port := getEnv("SERVER_PORT", "8080")
	log.Printf("✅ Server is running on http://localhost:%s", port)
	log.Printf("📧 Reminder worker is running (checking every 1 minute)")

	// Graceful shutdown для воркера при завершении сервера
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

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}