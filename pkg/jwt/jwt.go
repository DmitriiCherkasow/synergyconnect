package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims — структура, хранящая информацию в JWT-токене
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwtlib.RegisteredClaims
}

// Config — конфигурация JWT
type Config struct {
	SecretKey         string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

// TokenPair — пара токенов (access + refresh)
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // время жизни access в секундах
}

// JWTService — сервис для работы с JWT
type JWTService struct {
	config Config
}

// NewJWTService создает сервис для работы с JWT
func NewJWTService(config Config) *JWTService {
	return &JWTService{config: config}
}

// GenerateTokenPair генерирует пару токенов (access + refresh)
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error) {
	accessToken, expiresIn, err := s.generateAccessToken(userID, email, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// generateAccessToken создает access-токен
func (s *JWTService) generateAccessToken(userID uuid.UUID, email, role string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.AccessExpiration)
	expiresIn := int64(s.config.AccessExpiration.Seconds())

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
			Issuer:    "synergyconnect",
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// generateRefreshToken создает refresh-токен (хранит только userID)
func (s *JWTService) generateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.RefreshExpiration)

	claims := jwtlib.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwtlib.NewNumericDate(now),
		ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		Issuer:    "synergyconnect-refresh",
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// ValidateAccessToken проверяет access-токен и возвращает Claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken проверяет refresh-токен и возвращает userID
func (s *JWTService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &jwtlib.RegisteredClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*jwtlib.RegisteredClaims); ok && token.Valid {
		return uuid.Parse(claims.Subject)
	}

	return uuid.Nil, errors.New("invalid refresh token")
}
