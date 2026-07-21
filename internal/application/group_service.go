package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
)

// GroupRepository — интерфейс для работы с группами
type GroupRepository interface {
	Create(ctx context.Context, group *domain.Group) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Group, error)
	GetTree(ctx context.Context) ([]domain.Group, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Group, error)
	Update(ctx context.Context, group *domain.Group) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SubscriptionRepository — интерфейс для работы с подписками
type SubscriptionRepo interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, subscriberID, targetID uuid.UUID, subType domain.SubscriptionType) error
	FindBySubscriber(ctx context.Context, subscriberID uuid.UUID) ([]domain.Subscription, error)
	Exists(ctx context.Context, subscriberID, targetID uuid.UUID, subType domain.SubscriptionType) (bool, error)
}

// GroupService — сервис для работы с группами
type GroupService struct {
	groupRepo     GroupRepository
	subscriptionRepo SubscriptionRepo
}

// NewGroupService создает новый сервис
func NewGroupService(groupRepo GroupRepository, subscriptionRepo SubscriptionRepo) *GroupService {
	return &GroupService{
		groupRepo:     groupRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// GetGroupTree возвращает дерево групп
func (s *GroupService) GetGroupTree(ctx context.Context) ([]domain.Group, error) {
	return s.groupRepo.GetTree(ctx)
}

// GetGroupBySlug возвращает группу по slug
func (s *GroupService) GetGroupBySlug(ctx context.Context, slug string) (*domain.Group, error) {
	return s.groupRepo.FindBySlug(ctx, slug)
}

// SubscribeToGroup подписывает пользователя на группу
func (s *GroupService) SubscribeToGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	// Проверяем, существует ли группа
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	// Проверяем, не подписан ли уже
	exists, err := s.subscriptionRepo.Exists(ctx, userID, groupID, domain.SubscriptionGroup)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already subscribed to this group")
	}

	sub := &domain.Subscription{
		SubscriberID:  userID,
		TargetGroupID: &groupID,
		Type:          domain.SubscriptionGroup,
	}

	return s.subscriptionRepo.Create(ctx, sub)
}

// UnsubscribeFromGroup отписывает пользователя от группы
func (s *GroupService) UnsubscribeFromGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	return s.subscriptionRepo.Delete(ctx, userID, groupID, domain.SubscriptionGroup)
}

// GetUserGroupIDs возвращает ID всех групп, на которые подписан пользователь
func (s *GroupService) GetUserGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	subs, err := s.subscriptionRepo.FindBySubscriber(ctx, userID)
	if err != nil {
		return nil, err
	}

	var groupIDs []uuid.UUID
	for _, sub := range subs {
		if sub.Type == domain.SubscriptionGroup && sub.TargetGroupID != nil {
			groupIDs = append(groupIDs, *sub.TargetGroupID)
		}
	}
	return groupIDs, nil
}