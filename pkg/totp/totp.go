package totp

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/xlzd/gotp"
	"github.com/skip2/go-qrcode"
)

// TOTPProvider — провайдер для работы с TOTP
type TOTPProvider struct {
	issuer string
}

// NewTOTPProvider создает новый TOTP провайдер
func NewTOTPProvider(issuer string) *TOTPProvider {
	return &TOTPProvider{
		issuer: issuer,
	}
}

// GenerateSecret генерирует секретный ключ
func (p *TOTPProvider) GenerateSecret() string {
	secret := gotp.RandomSecret(20)
	return secret
}

// GenerateQRCode генерирует QR-код для Google Authenticator
func (p *TOTPProvider) GenerateQRCode(secret, email string) ([]byte, error) {
	otpURL := gotp.NewDefaultTOTP(secret).ProvisioningUri(email, p.issuer)
	
	qr, err := qrcode.Encode(otpURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	
	return qr, nil
}

// Verify проверяет TOTP код
func (p *TOTPProvider) Verify(secret, code string) bool {
	totp := gotp.NewDefaultTOTP(secret)
	return totp.Verify(code, time.Now().Unix())
}

// GenerateRecoveryCodes генерирует коды восстановления
func (p *TOTPProvider) GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Генерируем 8-значный код
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		// Преобразуем в шестнадцатеричную строку и берём первые 8 символов
		codes[i] = fmt.Sprintf("%X", bytes)[:8]
	}
	return codes, nil
}

// GetProvisioningURI возвращает URI для подключения
func (p *TOTPProvider) GetProvisioningURI(secret, email string) string {
	return gotp.NewDefaultTOTP(secret).ProvisioningUri(email, p.issuer)
}