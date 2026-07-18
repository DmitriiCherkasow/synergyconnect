package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/DmitriiCherkasow/synergyconnect/internal/application"
	"github.com/DmitriiCherkasow/synergyconnect/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect/internal/infrastructure/database"
	"github.com/DmitriiCherkasow/synergyconnect/internal/interfaces/http/handlers"
	"github.com/DmitriiCherkasow/synergyconnect/internal/interfaces/http/middleware"
	"github.com/DmitriiCherkasow/synergyconnect/pkg/jwt"
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

	// Автомиграция
	if err := db.AutoMigrate(&domain.User{}); err != nil {
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

	// Инициализируем репозитории и сервисы
	userRepo := database.NewUserRepository(db)
	authService := application.NewAuthService(userRepo, jwtService)

	// Инициализируем обработчики
	authHandler := handlers.NewAuthHandler(authService)

	// Настраиваем роутер
	r := gin.Default()

	// Health-check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "synergyconnect",
			"version": "0.1.0",
		})
	})

	// Группа API v1
	api := r.Group("/api/v1")
	{
		// Публичные эндпоинты (без JWT)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Защищенные эндпоинты (с JWT)
		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(jwtService))
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID := middleware.GetUserIDFromContext(c)
				c.JSON(200, gin.H{
					"message": "Authenticated access",
					"user_id": userID,
				})
			})
		}
	}

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