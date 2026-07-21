package handlers

import (

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/application"
)

// BoardHandler — обработчик для досок
type BoardHandler struct {
	boardService *application.BoardService
}

// NewBoardHandler создает новый обработчик
func NewBoardHandler(boardService *application.BoardService) *BoardHandler {
	return &BoardHandler{
		boardService: boardService,
	}
}

// TODO: Добавить методы в Шаге 2.3