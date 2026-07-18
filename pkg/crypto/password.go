package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Config — конфигурация для Argon2id
type Argon2Config struct {
	Time    uint32 // количество итераций
	Memory  uint32 // потребление памяти (KB)
	Threads uint8  // количество потоков
	KeyLen  uint32 // длина ключа
	SaltLen uint32 // длина соли
}

// DefaultArgon2Config — безопасные параметры по умолчанию
func DefaultArgon2Config() Argon2Config {
	return Argon2Config{
		Time:    3,      // 3 итерации
		Memory:  64 * 1024, // 64 MB
		Threads: 4,
		KeyLen:  32,
		SaltLen: 16,
	}
}

// HashPassword хеширует пароль с использованием Argon2id
// Возвращает строку в формате: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
func HashPassword(password string, config Argon2Config) (string, error) {
	// Генерация соли
	salt := make([]byte, config.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Хеширование пароля
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	// Формирование строки для хранения
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		config.Memory,
		config.Time,
		config.Threads,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword проверяет пароль на соответствие хешу
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Парсинг строки хеша
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	// Проверка версии
	if parts[1] != "argon2id" {
		return false, errors.New("unsupported algorithm")
	}

	// Парсинг параметров
	var version int
	var memory, time, threads int
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, errors.New("invalid parameters")
	}

	// Декодируем соль и хеш
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Хешируем пароль с теми же параметрами
	config := Argon2Config{
		Time:    uint32(time),
		Memory:  uint32(memory),
		Threads: uint8(threads),
		KeyLen:  uint32(len(hash)),
	}

	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	// Сравнение хешей (постоянное время для защиты от атак по времени)
	if subtle.ConstantTimeCompare(hash, newHash) == 1 {
		return true, nil
	}

	return false, nil
}

// MustHashPassword — хеширует пароль с параметрами по умолчанию, паникует при ошибке
func MustHashPassword(password string) string {
	hash, err := HashPassword(password, DefaultArgon2Config())
	if err != nil {
		panic(err)
	}
	return hash
}