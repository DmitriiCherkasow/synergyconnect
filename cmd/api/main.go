package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 SynergyConnect starting...")

	// Создаем роутер Gin
	r := gin.Default()

	// Health-check эндпоинт
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "synergyconnect",
			"version": "0.1.0",
		})
	})

	// Группа API v1
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}

	log.Println("✅ Server is running on http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}
