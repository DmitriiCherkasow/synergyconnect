package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// StickerService — сервис для работы со стикерами
type StickerService struct {
	stickerRepo  StickerRepository
	boardRepo    BoardRepository
	reminderRepo ReminderRepository
}

// NewStickerService создает новый сервис
func NewStickerService(
	stickerRepo StickerRepository,
	boardRepo BoardRepository,
	reminderRepo ReminderRepository,
) *StickerService {
	return &StickerService{
		stickerRepo:  stickerRepo,
		boardRepo:    boardRepo,
		reminderRepo: reminderRepo,
	}
}

// CreateStickerRequest — запрос на создание стикера
type CreateStickerRequest struct {
	BoardID     uuid.UUID
	AuthorID    uuid.UUID
	Title       string
	Content     string
	Color       domain.StickerColor
	Priority    domain.StickerPriority
	PositionX   int
	PositionY   int
	Width       int
	Height      int
	DueDate     *time.Time
}

// CreateSticker создает новый стикер
func (s *StickerService) CreateSticker(ctx context.Context, req CreateStickerRequest) (*domain.Sticker, error) {
	// Проверяем существование доски
	board, err := s.boardRepo.FindByID(ctx, req.BoardID)
	if err != nil {
		return nil, err
	}
	if board == nil {
		return nil, errors.New("board not found")
	}

	// Создаем стикер
	sticker := &domain.Sticker{
		BoardID:     req.BoardID,
		AuthorID:    req.AuthorID,
		Title:       req.Title,
		Content:     req.Content,
		Color:       req.Color,
		Priority:    req.Priority,
		PositionX:   req.PositionX,
		PositionY:   req.PositionY,
		Width:       req.Width,
		Height:      req.Height,
		DueDate:     req.DueDate,
		IsCompleted: false,
	}

	if err := s.stickerRepo.Create(ctx, sticker); err != nil {
		return nil, err
	}

	return sticker, nil
}

// GetSticker возвращает стикер по ID
func (s *StickerService) GetSticker(ctx context.Context, stickerID uuid.UUID) (*domain.Sticker, error) {
	return s.stickerRepo.FindByID(ctx, stickerID)
}

// GetBoardStickers возвращает все стикеры доски
func (s *StickerService) GetBoardStickers(ctx context.Context, boardID uuid.UUID) ([]domain.Sticker, error) {
	return s.stickerRepo.FindByBoardID(ctx, boardID)
}

// UpdateStickerRequest — запрос на обновление стикера
type UpdateStickerRequest struct {
	StickerID   uuid.UUID
	UserID      uuid.UUID
	UserRole    domain.UserRole
	Title       *string
	Content     *string
	Color       *domain.StickerColor
	Priority    *domain.StickerPriority
	Width       *int
	Height      *int
	DueDate     *time.Time
}

// UpdateSticker обновляет стикер
func (s *StickerService) UpdateSticker(ctx context.Context, req UpdateStickerRequest) (*domain.Sticker, error) {
	sticker, err := s.stickerRepo.FindByID(ctx, req.StickerID)
	if err != nil {
		return nil, err
	}
	if sticker == nil {
		return nil, errors.New("sticker not found")
	}

	if !sticker.CanEdit(req.UserID, req.UserRole) {
		return nil, errors.New("access denied")
	}

	if req.Title != nil {
		sticker.Title = *req.Title
	}
	if req.Content != nil {
		sticker.Content = *req.Content
	}
	if req.Color != nil {
		sticker.Color = *req.Color
	}
	if req.Priority != nil {
		sticker.Priority = *req.Priority
	}
	if req.Width != nil {
		sticker.Width = *req.Width
	}
	if req.Height != nil {
		sticker.Height = *req.Height
	}
	if req.DueDate != nil {
		sticker.DueDate = req.DueDate
	}

	if err := s.stickerRepo.Update(ctx, sticker); err != nil {
		return nil, err
	}

	return sticker, nil
}

// DeleteSticker удаляет стикер и все его напоминания
func (s *StickerService) DeleteSticker(ctx context.Context, stickerID, userID uuid.UUID, userRole domain.UserRole) error {
	sticker, err := s.stickerRepo.FindByID(ctx, stickerID)
	if err != nil {
		return err
	}
	if sticker == nil {
		return errors.New("sticker not found")
	}

	if !sticker.CanEdit(userID, userRole) {
		return errors.New("access denied")
	}

	// Удаляем все напоминания
	if err := s.reminderRepo.DeleteByStickerID(ctx, stickerID); err != nil {
		return err
	}

	return s.stickerRepo.Delete(ctx, stickerID)
}

// ToggleComplete переключает статус выполнения стикера
func (s *StickerService) ToggleComplete(ctx context.Context, stickerID, userID uuid.UUID, userRole domain.UserRole) error {
	sticker, err := s.stickerRepo.FindByID(ctx, stickerID)
	if err != nil {
		return err
	}
	if sticker == nil {
		return errors.New("sticker not found")
	}

	if !sticker.CanEdit(userID, userRole) {
		return errors.New("access denied")
	}

	if sticker.IsCompleted {
		// Возвращаем в работу: просто снимаем отметку
		sticker.IsCompleted = false
		return s.stickerRepo.Update(ctx, sticker)
	}

	return s.stickerRepo.MarkComplete(ctx, stickerID)
}

// UpdatePosition обновляет позицию стикера
func (s *StickerService) UpdatePosition(ctx context.Context, stickerID, userID uuid.UUID, userRole domain.UserRole, x, y int) error {
	sticker, err := s.stickerRepo.FindByID(ctx, stickerID)
	if err != nil {
		return err
	}
	if sticker == nil {
		return errors.New("sticker not found")
	}

	if !sticker.CanEdit(userID, userRole) {
		return errors.New("access denied")
	}

	return s.stickerRepo.UpdatePosition(ctx, stickerID, x, y)
}

