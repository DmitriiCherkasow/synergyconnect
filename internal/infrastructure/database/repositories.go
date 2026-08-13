package database

import (
    "github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
    "gorm.io/gorm"
)

// Repositories holds all repository instances
type Repositories struct {
    Project      domain.ProjectRepository
    Vacancy      domain.VacancyRepository
    Message      domain.MessageRepository
    Notification domain.NotificationRepository
    TwoFA        domain.TwoFARepository
    Device       domain.DeviceRepository
}

// NewRepositories creates all repository instances
func NewRepositories(db *gorm.DB) *Repositories {
    return &Repositories{
        Project:      NewProjectRepository(db),
        Vacancy:      NewVacancyRepository(db),
        Message:      NewMessageRepository(db),
        Notification: NewNotificationRepository(db),
        TwoFA:        NewTwoFARepository(db),
        Device:       NewDeviceRepository(db),
    }
}