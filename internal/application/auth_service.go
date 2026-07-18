package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/DmitriiCherkasow/synergyconnect/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect/pkg/crypto"
	"github.com/DmitriiCherkasow/synergyconnect/pkg/jwt"
)

// UserRepository — интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
}

// AuthService — сервис для аутентификации
type AuthService struct {
	userRepo   UserRepository
	jwtService *jwt.JWTService
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(userRepo UserRepository, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// RegisterRequest — данные для регистрации
type RegisterRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	// Проверяем, не существует ли уже пользователь с таким email
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := crypto.HashPassword(req.Password, crypto.DefaultArgon2Config())
	if err != nil {
		return nil, err
	}

	// Создаем пользователя
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         domain.RoleStudent, // По умолчанию — студент
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsVerified:   false,
		IsActive:     true,
	}

	// Сохраняем в БД
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginRequest — данные для входа
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResponse — результат входа
type LoginResponse struct {
	User         *domain.User
	TokenPair    *jwt.TokenPair
}

// Login выполняет вход пользователя
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Находим пользователя по email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Проверяем, активен ли пользователь
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Проверяем пароль
	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid credentials")
	}

	// Генерируем JWT-токены
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Обновляем время последнего входа
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Не критичная ошибка, логируем, но не прерываем
	}

	return &LoginResponse{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

// RefreshTokens обновляет access-токен
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	// Валидируем refresh-токен и получаем userID
	userID, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Находим пользователя
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Генерируем новую пару токенов
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// GetUserByID возвращает пользователя по ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}