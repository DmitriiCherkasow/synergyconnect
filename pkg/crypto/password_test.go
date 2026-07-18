package crypto

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "mySecurePassword123!"

	// Хешируем пароль
	hash, err := HashPassword(password, DefaultArgon2Config())
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	t.Logf("Hash: %s", hash)

	// Проверяем правильный пароль
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !valid {
		t.Error("Password verification failed for correct password")
	}

	// Проверяем неправильный пароль
	valid, err = VerifyPassword("wrongPassword", hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if valid {
		t.Error("Password verification succeeded for incorrect password")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	invalidHashes := []string{
		"invalid",
		"$argon2id$invalid",
		"$argon2id$v=19$invalid",
		"$argon2id$v=19$m=65536,t=3,p=4$invalid",
	}

	for _, hash := range invalidHashes {
		_, err := VerifyPassword("password", hash)
		if err == nil {
			t.Errorf("Expected error for invalid hash: %s", hash)
		}
	}
}
