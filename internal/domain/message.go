package domain

import (
    "time"
    "github.com/google/uuid"
)

type Message struct {
    ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
    SenderID   uuid.UUID  `json:"sender_id" gorm:"type:uuid;not null"`
    Sender     User       `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
    ReceiverID uuid.UUID  `json:"receiver_id" gorm:"type:uuid;not null"`
    Receiver   User       `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
    Content    string     `json:"content" gorm:"type:text;not null"`
    IsRead     bool       `json:"is_read" gorm:"default:false"`
    ReadAt     *time.Time `json:"read_at,omitempty"`
    CreatedAt  time.Time  `json:"created_at"`
}

type MessageResponse struct {
    ID         uuid.UUID  `json:"id"`
    SenderID   uuid.UUID  `json:"sender_id"`
    SenderName string     `json:"sender_name"`
    Content    string     `json:"content"`
    IsRead     bool       `json:"is_read"`
    CreatedAt  time.Time  `json:"created_at"`
}