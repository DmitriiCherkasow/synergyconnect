package handlers

import (

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
)

// StickerHandler — обработчик для стикеров
type StickerHandler struct {
	stickerService *application.StickerService
}

// NewStickerHandler создает новый обработчик
func NewStickerHandler(stickerService *application.StickerService) *StickerHandler {
	return &StickerHandler{
		stickerService: stickerService,
	}
}

// TODO: Добавить методы в Шаге 2.3