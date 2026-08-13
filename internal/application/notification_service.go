package application

import (
	"context"
	"encoding/json"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/google/uuid"
)

// NotificationService — сервис для работы с уведомлениями
type NotificationService struct {
	notifRepo domain.NotificationRepository
	userRepo  domain.UserRepository
	// wsHub     *websocket.Hub // добавим позже для реального времени
}

// NewNotificationService создает новый сервис уведомлений
func NewNotificationService(
	notifRepo domain.NotificationRepository,
	userRepo domain.UserRepository,
) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
		userRepo:  userRepo,
	}
}

// CreateNotificationRequest — данные для создания уведомления
type CreateNotificationRequest struct {
	UserID  uuid.UUID
	Type    domain.NotificationType
	Title   string
	Content string
	Link    string
	Payload interface{}
}

// CreateNotification создает уведомление для пользователя
func (s *NotificationService) CreateNotification(ctx context.Context, req CreateNotificationRequest) (*domain.Notification, error) {
	var payloadBytes json.RawMessage
	if req.Payload != nil {
		bytes, err := json.Marshal(req.Payload)
		if err != nil {
			return nil, err
		}
		payloadBytes = json.RawMessage(bytes)
	}

	notification := &domain.Notification{
		UserID:  req.UserID,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
		Link:    req.Link,
		Payload: payloadBytes,
		IsRead:  false,
	}

	if err := s.notifRepo.Create(ctx, notification); err != nil {
		return nil, err
	}

	// TODO: Отправить через WebSocket, если пользователь онлайн
	// if s.wsHub != nil {
	//     s.wsHub.SendToUser(req.UserID.String(), &websocket.WSMessage{
	//         Type: websocket.MsgTypeNewNotification,
	//         Payload: notification,
	//     })
	// }

	return notification, nil
}

// GetUserNotifications возвращает уведомления пользователя
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.notifRepo.GetByUser(ctx, userID, limit, offset)
}

// GetUnreadCount возвращает количество непрочитанных уведомлений
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notifRepo.GetUnreadCount(ctx, userID)
}

// MarkAsRead отмечает уведомление как прочитанное
func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	notif, err := s.notifRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notif == nil {
		return domain.ErrNotificationNotFound
	}
	if notif.UserID != userID {
		return domain.ErrForbidden
	}
	return s.notifRepo.MarkAsRead(ctx, id)
}

// MarkAllAsRead отмечает все уведомления как прочитанные
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

// DeleteNotification удаляет уведомление
func (s *NotificationService) DeleteNotification(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	notif, err := s.notifRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notif == nil {
		return domain.ErrNotificationNotFound
	}
	if notif.UserID != userID {
		return domain.ErrForbidden
	}
	return s.notifRepo.Delete(ctx, id)
}

// DeleteAllByUser удаляет все уведомления пользователя
func (s *NotificationService) DeleteAllByUser(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.DeleteAllByUser(ctx, userID)
}

// ============================================================
// ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ДЛЯ ИНТЕГРАЦИИ
// ============================================================

// NotifyProjectInvite уведомляет о приглашении в проект
func (s *NotificationService) NotifyProjectInvite(ctx context.Context, userID, projectID uuid.UUID, projectTitle string, invitedBy string) error {
	payload := map[string]interface{}{
		"project_id": projectID.String(),
		"invited_by": invitedBy,
	}
	_, err := s.CreateNotification(ctx, CreateNotificationRequest{
		UserID:  userID,
		Type:    domain.NotificationTypeProjectInvite,
		Title:   "Приглашение в проект",
		Content: "Вас пригласили в проект \"" + projectTitle + "\"",
		Link:    "/projects/" + projectID.String(),
		Payload: payload,
	})
	return err
}

// NotifyApplicationStatus уведомляет об изменении статуса заявки
func (s *NotificationService) NotifyApplicationStatus(ctx context.Context, userID, projectID uuid.UUID, projectTitle, status string) error {
	var title, content string
	switch status {
	case "accepted":
		title = "Заявка принята ✅"
		content = "Ваша заявка в проект \"" + projectTitle + "\" принята"
	case "rejected":
		title = "Заявка отклонена ❌"
		content = "Ваша заявка в проект \"" + projectTitle + "\" отклонена"
	default:
		return nil
	}

	payload := map[string]interface{}{
		"project_id": projectID.String(),
		"status":     status,
	}
	_, err := s.CreateNotification(ctx, CreateNotificationRequest{
		UserID:  userID,
		Type:    domain.NotificationTypeApplicationStatus,
		Title:   title,
		Content: content,
		Link:    "/projects/" + projectID.String(),
		Payload: payload,
	})
	return err
}

// NotifyNewMessage уведомляет о новом сообщении
func (s *NotificationService) NotifyNewMessage(ctx context.Context, userID, senderID uuid.UUID, senderName, messagePreview string) error {
	payload := map[string]interface{}{
		"sender_id": senderID.String(),
		"sender_name": senderName,
	}
	_, err := s.CreateNotification(ctx, CreateNotificationRequest{
		UserID:  userID,
		Type:    domain.NotificationTypeNewMessage,
		Title:   "Новое сообщение",
		Content: senderName + ": " + messagePreview,
		Link:    "/chat",
		Payload: payload,
	})
	return err
}

// NotifyVacancyResponse уведомляет о новом отклике на вакансию
func (s *NotificationService) NotifyVacancyResponse(ctx context.Context, employerID, vacancyID uuid.UUID, vacancyTitle, userName string) error {
	payload := map[string]interface{}{
		"vacancy_id": vacancyID.String(),
		"user_name":  userName,
	}
	_, err := s.CreateNotification(ctx, CreateNotificationRequest{
		UserID:  employerID,
		Type:    domain.NotificationTypeVacancyResponse,
		Title:   "Новый отклик на вакансию",
		Content: userName + " откликнулся на \"" + vacancyTitle + "\"",
		Link:    "/vacancies/" + vacancyID.String() + "/responses",
		Payload: payload,
	})
	return err
}

// NotifyComment уведомляет о новом комментарии
func (s *NotificationService) NotifyComment(ctx context.Context, userID uuid.UUID, postID uuid.UUID, authorName, commentPreview string) error {
	payload := map[string]interface{}{
		"post_id": postID.String(),
		"author":  authorName,
	}
	_, err := s.CreateNotification(ctx, CreateNotificationRequest{
		UserID:  userID,
		Type:    domain.NotificationTypeComment,
		Title:   "Новый комментарий",
		Content: authorName + " прокомментировал: " + commentPreview,
		Link:    "/posts/" + postID.String(),
		Payload: payload,
	})
	return err
}

// NotifySubscribe уведомляет о новой подписке
func (s *NotificationService) NotifySubscribe(ctx context.Context, userID uuid.UUID, subscriberName string) error {
	_, err := s.CreateNotification(ctx, CreateNotificationRequest{
		UserID:  userID,
		Type:    domain.NotificationTypeSubscribe,
		Title:   "Новый подписчик",
		Content: subscriberName + " подписался на вас",
		Link:    "/profile",
		Payload: nil,
	})
	return err
}