package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"
    "github.com/DmitriiCherkasow/synergyconnect.git/pkg/crypto"
    "github.com/google/uuid"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    email := flag.String("email", "", "Admin email")
    password := flag.String("password", "", "Admin password")
    flag.Parse()

    if *email == "" || *password == "" {
        fmt.Println("Usage: go run scripts/create_super_admin.go --email admin@example.com --password securepass")
        os.Exit(1)
    }

    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        dsn = "host=localhost user=synergy_user password=synergy_password dbname=synergy_db port=5432 sslmode=disable"
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("❌ Failed to connect to database: %v", err)
    }

    // Проверяем, существует ли уже суперадмин
    var existing domain.User
    err = db.Where("role = ?", domain.RoleSuperAdmin).First(&existing).Error
    if err == nil {
        log.Fatal("❌ Super admin already exists!")
    }

    // Хешируем пароль
    hashedPassword, err := crypto.HashPassword(*password, crypto.DefaultArgon2Config())
    if err != nil {
        log.Fatalf("❌ Failed to hash password: %v", err)
    }

    // Создаём суперадмина
    admin := &domain.User{
        ID:           uuid.New(),
        Email:        *email,
        PasswordHash: hashedPassword,
        Role:         domain.RoleSuperAdmin,
        FirstName:    "Super",
        LastName:     "Admin",
        IsVerified:   true,
        IsActive:     true,
    }

    if err := db.Create(admin).Error; err != nil {
        log.Fatalf("❌ Failed to create super admin: %v", err)
    }

    fmt.Printf("✅ Super admin created successfully!\n")
    fmt.Printf("   Email: %s\n", admin.Email)
    fmt.Printf("   Role: %s\n", admin.Role)
}