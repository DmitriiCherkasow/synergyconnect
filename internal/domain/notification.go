package domain

import (
    "encoding/json"
    "time"
    "github.com/google/uuid"
)

type NotificationType string

const (
    NotificationTypeProjectInvite      NotificationType = "project_invite"
    NotificationTypeApplicationStatus  NotificationType = "application_status"
    NotificationTypeNewMember          NotificationType = "new_member"
    NotificationTypeNewMessage         NotificationType = "new_message"
    NotificationTypeVacancyResponse    NotificationType = "vacancy_response"
    NotificationTypeVacancyStatus      NotificationType = "vacancy_status"
    NotificationTypeComment            NotificationType = "comment"
    NotificationTypeSubscribe          NotificationType = "subscribe"
    NotificationTypeReminder           NotificationType = "reminder"
)

type Notification struct {
    ID        uuid.UUID        `json:"id" gorm:"type:uuid;primary_key"`
    UserID    uuid.UUID        `json:"user_id" gorm:"type:uuid;not null"`
    User      User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Type      NotificationType `json:"type" gorm:"size:50;not null"`
    Title     string           `json:"title" gorm:"size:255;not null"`
    Content   string           `json:"content" gorm:"type:text"`
    Link      string           `json:"link" gorm:"size:500"`
    Payload   json.RawMessage  `json:"payload" gorm:"type:jsonb"`
    IsRead    bool             `json:"is_read" gorm:"default:false"`
    ReadAt    *time.Time       `json:"read_at,omitempty"`
    CreatedAt time.Time        `json:"created_at"`
}

type NotificationPayload struct {
    ProjectID   uuid.UUID `json:"project_id,omitempty"`
    VacancyID   uuid.UUID `json:"vacancy_id,omitempty"`
    MessageID   uuid.UUID `json:"message_id,omitempty"`
    SenderID    uuid.UUID `json:"sender_id,omitempty"`
    SenderName  string    `json:"sender_name,omitempty"`
    ActionBy    uuid.UUID `json:"action_by,omitempty"`
    ActionName  string    `json:"action_name,omitempty"`
}