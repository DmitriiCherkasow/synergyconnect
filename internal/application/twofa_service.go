package application

import (
	"context"
	"errors"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
	"github.com/DmitriiCherkasow/synergyconnect.git/pkg/totp"
	"github.com/google/uuid"
)

// TwoFAService — сервис для работы с 2FA
type TwoFAService struct {
	twofaRepo   domain.TwoFARepository
	userRepo    domain.UserRepository
	totpProvider *totp.TOTPProvider
}

// NewTwoFAService создает новый сервис 2FA
func NewTwoFAService(
	twofaRepo domain.TwoFARepository,
	userRepo domain.UserRepository,
	totpProvider *totp.TOTPProvider,
) *TwoFAService {
	return &TwoFAService{
		twofaRepo:   twofaRepo,
		userRepo:    userRepo,
		totpProvider: totpProvider,
	}
}

// Enable2FARequest — данные для включения 2FA
type Enable2FARequest struct {
	Code string
}

// Enable2FAResponse — ответ при включении 2FA
type Enable2FAResponse struct {
	Secret        string   `json:"secret"`
	QRCode        []byte   `json:"qr_code"`
	RecoveryCodes []string `json:"recovery_codes"`
}

// Initiate2FA начинает процесс включения 2FA
func (s *TwoFAService) Initiate2FA(ctx context.Context, userID uuid.UUID) (*Enable2FAResponse, error) {
	// Проверяем, не включена ли уже 2FA
	existing, err := s.twofaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.Enabled {
		return nil, domain.ErrTwoFAAlreadyEnabled
	}

	// Генерируем секрет
	secret := s.totpProvider.GenerateSecret()

	// Получаем пользователя
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Генерируем QR-код
	qrCode, err := s.totpProvider.GenerateQRCode(secret, user.Email)
	if err != nil {
		return nil, err
	}

	// Генерируем коды восстановления
	recoveryCodes, err := s.totpProvider.GenerateRecoveryCodes(10)
	if err != nil {
		return nil, err
	}

	// Сохраняем секрет и коды восстановления (но не включаем 2FA до верификации)
	twofa := &domain.UserTwoFA{
		UserID:        userID,
		Secret:        secret,
		Enabled:       false,
		RecoveryCodes: recoveryCodes,
	}
	if err := s.twofaRepo.CreateOrUpdate(ctx, twofa); err != nil {
		return nil, err
	}

	return &Enable2FAResponse{
		Secret:        secret,
		QRCode:        qrCode,
		RecoveryCodes: recoveryCodes,
	}, nil
}

// VerifyAndEnable2FA проверяет код и включает 2FA
func (s *TwoFAService) VerifyAndEnable2FA(ctx context.Context, userID uuid.UUID, code string) error {
	// Получаем данные 2FA
	twofa, err := s.twofaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if twofa == nil {
		return errors.New("2FA not initiated")
	}
	if twofa.Enabled {
		return domain.ErrTwoFAAlreadyEnabled
	}

	// Проверяем код
	if !s.totpProvider.Verify(twofa.Secret, code) {
		return domain.ErrInvalidCode
	}

	// Включаем 2FA
	return s.twofaRepo.Enable(ctx, userID, twofa.Secret)
}

// Disable2FA отключает 2FA
func (s *TwoFAService) Disable2FA(ctx context.Context, userID uuid.UUID, code string) error {
	// Получаем данные 2FA
	twofa, err := s.twofaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if twofa == nil || !twofa.Enabled {
		return domain.ErrTwoFANotEnabled
	}

	// Проверяем код
	if !s.totpProvider.Verify(twofa.Secret, code) {
		return domain.ErrInvalidCode
	}

	return s.twofaRepo.Disable(ctx, userID)
}

// VerifyCode проверяет TOTP код без изменения статуса
func (s *TwoFAService) VerifyCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	twofa, err := s.twofaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	if twofa == nil || !twofa.Enabled {
		return false, domain.ErrTwoFANotEnabled
	}

	return s.totpProvider.Verify(twofa.Secret, code), nil
}

// IsEnabled проверяет, включена ли 2FA для пользователя
func (s *TwoFAService) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	return s.twofaRepo.IsEnabled(ctx, userID)
}

// GetRecoveryCodes возвращает коды восстановления
func (s *TwoFAService) GetRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	twofa, err := s.twofaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if twofa == nil {
		return nil, domain.ErrTwoFANotEnabled
	}
	return twofa.RecoveryCodes, nil
}

// RegenerateRecoveryCodes генерирует новые коды восстановления
func (s *TwoFAService) RegenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	// Проверяем код
	valid, err := s.VerifyCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, domain.ErrInvalidCode
	}

	// Генерируем новые коды
	newCodes, err := s.totpProvider.GenerateRecoveryCodes(10)
	if err != nil {
		return nil, err
	}

	// Обновляем коды в БД
	if err := s.twofaRepo.UpdateRecoveryCodes(ctx, userID, newCodes); err != nil {
		return nil, err
	}

	return newCodes, nil
}