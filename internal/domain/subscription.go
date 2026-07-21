package domain

import (
	"time"
	"github.com/google/uuid"
)

// SubscriptionType — тип подписки
type SubscriptionType string

const (
	SubscriptionUser  SubscriptionType = "user"  // Подписка на пользователя
	SubscriptionGroup SubscriptionType = "group" // Подписка на группу
)

// Subscription — модель подписки
type Subscription struct {
	ID         uuid.UUID          `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SubscriberID uuid.UUID        `json:"subscriber_id" gorm:"type:uuid;not null"`
	TargetUserID  *uuid.UUID      `json:"target_user_id,omitempty" gorm:"type:uuid"`
	TargetGroupID *uuid.UUID      `json:"target_group_id,omitempty" gorm:"type:uuid"`
	Type          SubscriptionType `json:"type" gorm:"not null"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`

	// Связи
	Subscriber User  `json:"subscriber" gorm:"foreignKey:SubscriberID"`
	TargetUser *User `json:"target_user,omitempty" gorm:"foreignKey:TargetUserID"`
	TargetGroup *Group `json:"target_group,omitempty" gorm:"foreignKey:TargetGroupID"`
}