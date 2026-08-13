package dto

// Initiate2FAResponse — ответ при инициализации 2FA
type Initiate2FAResponse struct {
	Secret        string `json:"secret"`
	QRCode        string `json:"qr_code"` // base64
	RecoveryCodes []string `json:"recovery_codes"`
}

// Enable2FARequest — запрос на включение 2FA
type Enable2FARequest struct {
	Code string `json:"code" binding:"required"`
}

// Disable2FARequest — запрос на отключение 2FA
type Disable2FARequest struct {
	Code string `json:"code" binding:"required"`
}

// Verify2FARequest — запрос на проверку кода
type Verify2FARequest struct {
	Code string `json:"code" binding:"required"`
}

// RegenerateCodesRequest — запрос на перегенерацию кодов восстановления
type RegenerateCodesRequest struct {
	Code string `json:"code" binding:"required"`
}