package dto

import (
	"encoding/json"
	"time"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// NotificationResponse — ответ с уведомлением
type NotificationResponse struct {
	ID        uuid.UUID               `json:"id"`
	Type      domain.NotificationType `json:"type"`
	Title     string                  `json:"title"`
	Content   string                  `json:"content"`
	Link      string                  `json:"link,omitempty"`
	Payload   json.RawMessage         `json:"payload,omitempty"`
	IsRead    bool                    `json:"is_read"`
	CreatedAt time.Time               `json:"created_at"`
}

// NotificationListResponse — список уведомлений
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	UnreadCount   int64                  `json:"unread_count"`
}

// MarkReadRequest — запрос на отметку прочитанных
type MarkReadRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

// ToNotificationResponse конвертирует доменную модель в DTO
func ToNotificationResponse(notif *domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        notif.ID,
		Type:      notif.Type,
		Title:     notif.Title,
		Content:   notif.Content,
		Link:      notif.Link,
		Payload:   notif.Payload,
		IsRead:    notif.IsRead,
		CreatedAt: notif.CreatedAt,
	}
}