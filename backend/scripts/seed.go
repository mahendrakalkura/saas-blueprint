package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/database"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
)

func main() {
	log.Println("Starting database seeding...")

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	ctx := context.Background()

	// Create repositories
	userRepo := repository.NewUserRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)

	// Create demo users
	users := []struct {
		Email     string
		Password  string
		FirstName string
		LastName  string
	}{
		{"admin@example.com", "password123", "Admin", "User"},
		{"john@example.com", "password123", "John", "Doe"},
		{"jane@example.com", "password123", "Jane", "Smith"},
	}

	log.Println("Creating demo users...")

	for _, u := range users {
		// Check if user already exists
		existing, _ := userRepo.GetByEmail(ctx, u.Email)
		if existing != nil {
			log.Printf("User %s already exists, skipping", u.Email)
			continue
		}

		// Hash password
		passwordHash, err := auth.HashPassword(u.Password)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", u.Email, err)
			continue
		}

		// Create user
		user := &struct {
			ID            string
			Email         string
			PasswordHash  string
			FirstName     *string
			LastName      *string
			EmailVerified bool
			IsActive      bool
			CreatedAt     time.Time
			UpdatedAt     time.Time
		}{
			ID:            uuid.New().String(),
			Email:         u.Email,
			PasswordHash:  passwordHash,
			FirstName:     &u.FirstName,
			LastName:      &u.LastName,
			EmailVerified: true, // Pre-verified for demo
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		query := `
			INSERT INTO users (id, email, password_hash, first_name, last_name, email_verified, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		_, err = db.DB.ExecContext(ctx, query,
			user.ID,
			user.Email,
			user.PasswordHash,
			user.FirstName,
			user.LastName,
			user.EmailVerified,
			user.IsActive,
			user.CreatedAt,
			user.UpdatedAt,
		)

		if err != nil {
			log.Printf("Failed to create user %s: %v", u.Email, err)
			continue
		}

		log.Printf("✓ Created user: %s (password: password123)", u.Email)
	}

	log.Println("\nSeeding completed!")
	log.Println("\nDemo credentials:")
	log.Println("  Email: admin@example.com")
	log.Println("  Password: password123")
	log.Println("")
	log.Println("  Email: john@example.com")
	log.Println("  Password: password123")
	log.Println("")
	log.Println("  Email: jane@example.com")
	log.Println("  Password: password123")

	// Avoid unused variable warning
	_ = sessionRepo
	_ = fmt.Sprint("")
}
