package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// CreateBoardRequest — запрос на создание доски
type CreateBoardRequest struct {
	GroupID     *uuid.UUID `json:"group_id,omitempty"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	ColorHex    string     `json:"color_hex" binding:"omitempty,hexcolor"`
}

// UpdateBoardRequest — запрос на обновление доски
type UpdateBoardRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ColorHex    *string `json:"color_hex" binding:"omitempty,hexcolor"`
}

// BoardResponse — ответ с данными доски
type BoardResponse struct {
	ID          string             `json:"id"`
	Owner       UserResponse       `json:"owner"`
	Group       *GroupResponse     `json:"group,omitempty"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	ColorHex    string             `json:"color_hex"`
	IsArchived  bool               `json:"is_archived"`
	Stickers    []StickerResponse  `json:"stickers,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// StickerResponse — ответ с данными стикера
type StickerResponse struct {
	ID          string     `json:"id"`
	Author      UserResponse `json:"author"`
	BoardID     string     `json:"board_id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Color       string     `json:"color"`
	Priority    string     `json:"priority"`
	PositionX   int        `json:"position_x"`
	PositionY   int        `json:"position_y"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	IsCompleted bool       `json:"is_completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateStickerRequest — запрос на создание стикера
type CreateStickerRequest struct {
	Title    string     `json:"title"`
	Content  string     `json:"content" binding:"required"`
	Color    string     `json:"color" binding:"omitempty,oneof=yellow blue green red purple orange gray"`
	Priority string     `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	PositionX int        `json:"position_x"`
	PositionY int        `json:"position_y"`
	Width    int        `json:"width"`
	Height   int        `json:"height"`
	DueDate  *time.Time `json:"due_date,omitempty"`
}

// UpdateStickerRequest — запрос на обновление стикера
type UpdateStickerRequest struct {
	Title       *string     `json:"title"`
	Content     *string     `json:"content"`
	Color       *string     `json:"color" binding:"omitempty,oneof=yellow blue green red purple orange gray"`
	Priority    *string     `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	Width       *int        `json:"width"`
	Height      *int        `json:"height"`
	DueDate     *time.Time  `json:"due_date,omitempty"`
}

// UpdateStickerPositionRequest — запрос на обновление позиции стикера
type UpdateStickerPositionRequest struct {
	PositionX int `json:"position_x" binding:"required"`
	PositionY int `json:"position_y" binding:"required"`
}

// ToBoardResponse преобразует доменную модель в ответ
func ToBoardResponse(board *domain.Board, stickers []domain.Sticker) BoardResponse {
	resp := BoardResponse{
		ID:          board.ID.String(),
		Owner:       ToUserResponse(&board.Owner),
		Title:       board.Title,
		Description: board.Description,
		ColorHex:    board.ColorHex,
		IsArchived:  board.IsArchived,
		CreatedAt:   board.CreatedAt,
		UpdatedAt:   board.UpdatedAt,
	}

	if board.Group != nil {
		groupResp := ToGroupResponse(board.Group)
		resp.Group = &groupResp
	}

	for _, sticker := range stickers {
		resp.Stickers = append(resp.Stickers, ToStickerResponse(&sticker))
	}

	return resp
}

// ToStickerResponse преобразует доменную модель стикера в ответ
func ToStickerResponse(sticker *domain.Sticker) StickerResponse {
	return StickerResponse{
		ID:          sticker.ID.String(),
		Author:      ToUserResponse(&sticker.Author),
		BoardID:     sticker.BoardID.String(),
		Title:       sticker.Title,
		Content:     sticker.Content,
		Color:       string(sticker.Color),
		Priority:    string(sticker.Priority),
		PositionX:   sticker.PositionX,
		PositionY:   sticker.PositionY,
		Width:       sticker.Width,
		Height:      sticker.Height,
		DueDate:     sticker.DueDate,
		IsCompleted: sticker.IsCompleted,
		CreatedAt:   sticker.CreatedAt,
		UpdatedAt:   sticker.UpdatedAt,
	}
}