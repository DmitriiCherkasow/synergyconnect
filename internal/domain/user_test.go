package domain

import (
	"testing"
)

func TestUserRole(t *testing.T) {
	user := &User{
		Email: "test@example.com",
		Role:  RoleStudent,
	}

	if !user.IsStudent() {
		t.Error("Expected user to be student")
	}

	if user.IsAdmin() {
		t.Error("Expected user not to be admin")
	}
}

func TestUserFullName(t *testing.T) {
	user := &User{
		FirstName: "Иван",
		LastName:  "Иванов",
	}

	expected := "Иван Иванов"
	if user.FullName() != expected {
		t.Errorf("Expected %s, got %s", expected, user.FullName())
	}
}