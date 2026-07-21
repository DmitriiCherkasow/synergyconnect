package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// BoardRepository — интерфейс для работы с досками
type BoardRepository interface {
	Create(ctx context.Context, board *domain.Board) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error)
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]domain.Board, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]domain.Board, error)
	Update(ctx context.Context, board *domain.Board) error
	Delete(ctx context.Context, id uuid.UUID) error
	Archive(ctx context.Context, id uuid.UUID) error
	Unarchive(ctx context.Context, id uuid.UUID) error
}

// StickerRepository — интерфейс для работы со стикерами
type StickerRepository interface {
	Create(ctx context.Context, sticker *domain.Sticker) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Sticker, error)
	FindByBoardID(ctx context.Context, boardID uuid.UUID) ([]domain.Sticker, error)
	FindByAuthorID(ctx context.Context, authorID uuid.UUID) ([]domain.Sticker, error)
	Update(ctx context.Context, sticker *domain.Sticker) error
	Delete(ctx context.Context, id uuid.UUID) error
	MarkComplete(ctx context.Context, id uuid.UUID) error
	UpdatePosition(ctx context.Context, id uuid.UUID, x, y int) error
}

// ReminderRepository — интерфейс для работы с напоминаниями
type ReminderRepository interface {
	Create(ctx context.Context, reminder *domain.Reminder) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Reminder, error)
	FindPending(ctx context.Context) ([]domain.Reminder, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Reminder, error)
	FindByStickerID(ctx context.Context, stickerID uuid.UUID) ([]domain.Reminder, error)
	MarkAsSent(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByStickerID(ctx context.Context, stickerID uuid.UUID) error
}

// BoardService — сервис для работы с досками
type BoardService struct {
	boardRepo    BoardRepository
	stickerRepo  StickerRepository
	reminderRepo ReminderRepository
}

// NewBoardService создает новый сервис
func NewBoardService(
	boardRepo BoardRepository,
	stickerRepo StickerRepository,
	reminderRepo ReminderRepository,
) *BoardService {
	return &BoardService{
		boardRepo:    boardRepo,
		stickerRepo:  stickerRepo,
		reminderRepo: reminderRepo,
	}
}

// CreateBoardRequest — запрос на создание доски
type CreateBoardRequest struct {
	OwnerID     uuid.UUID
	GroupID     *uuid.UUID
	Title       string
	Description string
	ColorHex    string
}

// CreateBoard создает новую доску
func (s *BoardService) CreateBoard(ctx context.Context, req CreateBoardRequest) (*domain.Board, error) {
	board := &domain.Board{
		OwnerID:     req.OwnerID,
		GroupID:     req.GroupID,
		Title:       req.Title,
		Description: req.Description,
		ColorHex:    req.ColorHex,
		IsArchived:  false,
	}

	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

// GetBoard возвращает доску по ID с проверкой прав
func (s *BoardService) GetBoard(ctx context.Context, boardID, userID uuid.UUID, userRole domain.UserRole, userGroupIDs []uuid.UUID) (*domain.Board, error) {
	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board == nil {
		return nil, errors.New("board not found")
	}

	if !board.CanView(userID, userRole, userGroupIDs) {
		return nil, errors.New("access denied")
	}

	return board, nil
}

// GetUserBoards возвращает все доски пользователя (личные + групповые)
func (s *BoardService) GetUserBoards(ctx context.Context, userID uuid.UUID, userGroupIDs []uuid.UUID) ([]domain.Board, error) {
	var boards []domain.Board

	// Личные доски
	personalBoards, err := s.boardRepo.FindByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	boards = append(boards, personalBoards...)

	// Групповые доски
	for _, groupID := range userGroupIDs {
		groupBoards, err := s.boardRepo.FindByGroupID(ctx, groupID)
		if err != nil {
			continue
		}
		boards = append(boards, groupBoards...)
	}

	return boards, nil
}

// UpdateBoard обновляет доску
func (s *BoardService) UpdateBoard(ctx context.Context, boardID, userID uuid.UUID, userRole domain.UserRole, title, description, colorHex *string) (*domain.Board, error) {
	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board == nil {
		return nil, errors.New("board not found")
	}

	if !board.CanEdit(userID, userRole) {
		return nil, errors.New("access denied")
	}

	if title != nil {
		board.Title = *title
	}
	if description != nil {
		board.Description = *description
	}
	if colorHex != nil {
		board.ColorHex = *colorHex
	}

	if err := s.boardRepo.Update(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

// DeleteBoard удаляет доску (каскадно)
func (s *BoardService) DeleteBoard(ctx context.Context, boardID, userID uuid.UUID, userRole domain.UserRole) error {
	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return err
	}
	if board == nil {
		return errors.New("board not found")
	}

	if !board.CanEdit(userID, userRole) {
		return errors.New("access denied")
	}

	// Удаляем все напоминания, связанные со стикерами доски
	stickers, err := s.stickerRepo.FindByBoardID(ctx, boardID)
	if err != nil {
		return err
	}
	for _, sticker := range stickers {
		if err := s.reminderRepo.DeleteByStickerID(ctx, sticker.ID); err != nil {
			continue
		}
	}

	return s.boardRepo.Delete(ctx, boardID)
}

// ArchiveBoard архивирует доску
func (s *BoardService) ArchiveBoard(ctx context.Context, boardID uuid.UUID) error {
	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return err
	}
	if board == nil {
		return errors.New("board not found")
	}

	return s.boardRepo.Archive(ctx, boardID)
}

// UnarchiveBoard разархивирует доску
func (s *BoardService) UnarchiveBoard(ctx context.Context, boardID uuid.UUID) error {
	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return err
	}
	if board == nil {
		return errors.New("board not found")
	}

	return s.boardRepo.Unarchive(ctx, boardID)
}