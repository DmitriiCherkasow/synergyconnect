package main

import (
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
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/interfaces/http/handlers"
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

	// ============================================================
	// МИГРАЦИИ: Добавляем все модели
	// ============================================================
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Post{},
		&domain.Comment{},
		&domain.Tag{},
		&domain.Group{},
		&domain.Subscription{},
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

	// ============================================================
	// ИНИЦИАЛИЗАЦИЯ СЕРВИСОВ
	// ============================================================
	authService := application.NewAuthService(userRepo, jwtService)
	postService := application.NewPostService(postRepo, commentRepo, tagRepo)
	groupService := application.NewGroupService(groupRepo, subscriptionRepo)

	// ============================================================
	// ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ
	// ============================================================
	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService, groupService)
	commentHandler := handlers.NewCommentHandler(postService)
	groupHandler := handlers.NewGroupHandler(groupService)

	// Настраиваем роутер
	r := gin.Default()

	// ============================================================
	// НАСТРОЙКА МАРШРУТОВ (используем функцию SetupRoutes)
	// ============================================================
	http.SetupRoutes(r, authHandler, postHandler, commentHandler, groupHandler, jwtService)

	// Запускаем сервер
	port := getEnv("SERVER_PORT", "8080")
	log.Printf("✅ Server is running on http://localhost:%s", port)
	log.Fatal(r.Run(":" + port))
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}