package database

import (
    "context"
    "errors"
    "strings"

    "github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// UserRepository — репозиторий для работы с пользователями
type UserRepository struct {
    db *gorm.DB
}

// NewUserRepository создает новый репозиторий
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// FindByEmail ищет пользователя по email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// FindByID ищет пользователя по ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

// UpdateLastLogin обновляет время последнего входа
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).
        Model(&domain.User{}).
        Where("id = ?", userID).
        Update("last_login_at", gorm.Expr("NOW()")).Error
}

// List возвращает список пользователей с фильтрацией
func (r *UserRepository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int64, error) {
    var users []domain.User
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.User{})

    if filter.Search != "" {
        search := "%" + strings.ToLower(filter.Search) + "%"
        query = query.Where(
            "LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
            search, search, search,
        )
    }

    if filter.Role != nil {
        query = query.Where("role = ?", *filter.Role)
    }

    if filter.IsActive != nil {
        query = query.Where("is_active = ?", *filter.IsActive)
    }

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if filter.Limit > 0 {
        query = query.Limit(filter.Limit)
    }
    if filter.Offset > 0 {
        query = query.Offset(filter.Offset)
    }

    err := query.Order("created_at DESC").Find(&users).Error
    return users, total, err
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

// FindSuperAdmin находит суперадмина
func (r *UserRepository) FindSuperAdmin(ctx context.Context) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).
        Where("role = ?", domain.RoleSuperAdmin).
        First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// CountAdmins считает количество администраторов (включая super_admin)
func (r *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&domain.User{}).
        Where("role IN ?", []domain.UserRole{domain.RoleAdmin, domain.RoleSuperAdmin}).
        Count(&count).Error
    return count, err
}

// FindAdmins находит всех администраторов
func (r *UserRepository) FindAdmins(ctx context.Context) ([]domain.User, error) {
    var users []domain.User
    err := r.db.WithContext(ctx).
        Where("role IN ?", []domain.UserRole{domain.RoleAdmin, domain.RoleSuperAdmin}).
        Find(&users).Error
    return users, err
}

// PromoteToAdmin повышает пользователя до администратора
func (r *UserRepository) PromoteToAdmin(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).
        Model(&domain.User{}).
        Where("id = ?", userID).
        Update("role", domain.RoleAdmin).Error
}

// DemoteFromAdmin понижает администратора до студента
func (r *UserRepository) DemoteFromAdmin(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).
        Model(&domain.User{}).
        Where("id = ? AND role = ?", userID, domain.RoleAdmin).
        Update("role", domain.RoleStudent).Error
}

// Count implements domain.UserRepository
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error
    return count, err
}

// CountActive implements domain.UserRepository
func (r *UserRepository) CountActive(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.User{}).Where("is_active = ?", true).Count(&count).Error
    return count, err
}

// CountNewToday implements domain.UserRepository
func (r *UserRepository) CountNewToday(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.User{}).
        Where("created_at >= CURRENT_DATE").
        Count(&count).Error
    return count, err
}