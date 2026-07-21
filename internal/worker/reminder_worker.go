package worker

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/internal/infrastructure/email"
)

// ReminderRepository — интерфейс для работы с напоминаниями
type ReminderRepository interface {
	FindPending(ctx context.Context) ([]domain.Reminder, error)
	MarkAsSent(ctx context.Context, id uuid.UUID) error
}

// StickerRepository — интерфейс для работы со стикерами
type StickerRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Sticker, error)
}

// BoardRepository — интерфейс для работы с досками
type BoardRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error)
}

// UserRepository — интерфейс для работы с пользователями
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// ReminderEmailRepository — интерфейс для логирования писем
type ReminderEmailRepository interface {
	Create(ctx context.Context, log *domain.ReminderEmail) error
}

// ReminderWorker — фоновый воркер для отправки напоминаний
type ReminderWorker struct {
	reminderRepo      ReminderRepository
	stickerRepo       StickerRepository
	boardRepo         BoardRepository
	userRepo          UserRepository
	reminderEmailRepo ReminderEmailRepository
	emailService      *email.Service
	interval          time.Duration
	stopCh            chan struct{}
}

// NewReminderWorker создает новый воркер
func NewReminderWorker(
	reminderRepo ReminderRepository,
	stickerRepo StickerRepository,
	boardRepo BoardRepository,
	userRepo UserRepository,
	reminderEmailRepo ReminderEmailRepository,
	emailService *email.Service,
	interval time.Duration,
) *ReminderWorker {
	return &ReminderWorker{
		reminderRepo:      reminderRepo,
		stickerRepo:       stickerRepo,
		boardRepo:         boardRepo,
		userRepo:          userRepo,
		reminderEmailRepo: reminderEmailRepo,
		emailService:      emailService,
		interval:          interval,
		stopCh:            make(chan struct{}),
	}
}

// Start запускает воркер
func (w *ReminderWorker) Start(ctx context.Context) {
	log.Println("📧 Reminder worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processReminders(ctx)
		case <-w.stopCh:
			log.Println("📧 Reminder worker stopped")
			return
		case <-ctx.Done():
			log.Println("📧 Reminder worker stopped by context")
			return
		}
	}
}

// Stop останавливает воркер
func (w *ReminderWorker) Stop() {
	close(w.stopCh)
}

// processReminders обрабатывает все ожидающие напоминания
func (w *ReminderWorker) processReminders(ctx context.Context) {
	log.Println("📧 Checking pending reminders...")

	reminders, err := w.reminderRepo.FindPending(ctx)
	if err != nil {
		log.Printf("❌ Failed to fetch pending reminders: %v", err)
		return
	}

	if len(reminders) == 0 {
		log.Println("📧 No pending reminders")
		return
	}

	log.Printf("📧 Found %d pending reminders", len(reminders))

	for _, reminder := range reminders {
		w.processReminder(ctx, &reminder)
	}
}

// processReminder обрабатывает одно напоминание
func (w *ReminderWorker) processReminder(ctx context.Context, reminder *domain.Reminder) {
	log.Printf("📧 Processing reminder %s for sticker %s", reminder.ID.String(), reminder.StickerID.String())

	// Получаем стикер
	sticker, err := w.stickerRepo.FindByID(ctx, reminder.StickerID)
	if err != nil {
		log.Printf("❌ Failed to fetch sticker %s: %v", reminder.StickerID.String(), err)
		w.logEmail(ctx, reminder, "", "failed to fetch sticker")
		return
	}
	if sticker == nil {
		log.Printf("❌ Sticker %s not found", reminder.StickerID.String())
		w.logEmail(ctx, reminder, "", "sticker not found")
		w.reminderRepo.MarkAsSent(ctx, reminder.ID)
		return
	}

	// Получаем доску
	board, err := w.boardRepo.FindByID(ctx, sticker.BoardID)
	if err != nil {
		log.Printf("❌ Failed to fetch board %s: %v", sticker.BoardID.String(), err)
		w.logEmail(ctx, reminder, "", "failed to fetch board")
		return
	}
	if board == nil {
		log.Printf("❌ Board %s not found", sticker.BoardID.String())
		w.logEmail(ctx, reminder, "", "board not found")
		w.reminderRepo.MarkAsSent(ctx, reminder.ID)
		return
	}

	// Получаем пользователя
	user, err := w.userRepo.FindByID(ctx, reminder.UserID)
	if err != nil {
		log.Printf("❌ Failed to fetch user %s: %v", reminder.UserID.String(), err)
		w.logEmail(ctx, reminder, "", "failed to fetch user")
		return
	}
	if user == nil {
		log.Printf("❌ User %s not found", reminder.UserID.String())
		w.logEmail(ctx, reminder, "", "user not found")
		w.reminderRepo.MarkAsSent(ctx, reminder.ID)
		return
	}

	// Отправляем email
	err = w.emailService.SendReminder(user.Email, user.FullName(), reminder, sticker, board)
	if err != nil {
		log.Printf("❌ Failed to send email for reminder %s: %v", reminder.ID.String(), err)
		w.logEmail(ctx, reminder, user.Email, err.Error())
		return
	}

	// Логируем успешную отправку
	w.logEmail(ctx, reminder, user.Email, "")

	// Отмечаем напоминание как отправленное
	if err := w.reminderRepo.MarkAsSent(ctx, reminder.ID); err != nil {
		log.Printf("❌ Failed to mark reminder %s as sent: %v", reminder.ID.String(), err)
	}

	log.Printf("✅ Reminder %s sent to %s", reminder.ID.String(), user.Email)

	// Создаем следующее напоминание для recurrence
	if reminder.Recurrence != domain.RecurrenceOneTime {
		next := reminder.CreateNextReminder()
		if next != nil {
			// TODO: Создать следующее напоминание через reminderRepo.Create()
			log.Printf("📧 Would create next reminder for %s (recurrence: %s)", sticker.ID.String(), reminder.Recurrence)
		}
	}
}

// logEmail логирует отправку email
func (w *ReminderWorker) logEmail(ctx context.Context, reminder *domain.Reminder, to, errMsg string) {
	status := domain.EmailStatusSent
	if errMsg != "" {
		status = domain.EmailStatusFailed
	}

	logEntry := &domain.ReminderEmail{
		ReminderID: reminder.ID,
		UserID:     reminder.UserID,
		EmailTo:    to,
		Subject:    "🔔 SynergyConnect — Напоминание о задаче",
		Status:     status,
		Error:      errMsg,
	}

	if err := w.reminderEmailRepo.Create(ctx, logEntry); err != nil {
		log.Printf("❌ Failed to log email for reminder %s: %v", reminder.ID.String(), err)
	}
}