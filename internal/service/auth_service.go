package service

import (
	"context"
	"errors"
	"time"

	"pocketful/internal/config"
	"pocketful/internal/models"
	"pocketful/internal/repository"
	"pocketful/internal/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles user registration and login logic.
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register creates a new user with a hashed password.
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	// Check if email already exists
	_, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("email already registered")
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         models.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.userRepo.Create(ctx, user)
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(
		user.ID.Hex(),
		user.Email,
		string(user.Role),
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTExpiryHours,
	)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{Token: token, User: *user}, nil
}

// SeedAdminUser creates a default admin user if none exists.
func (s *AuthService) SeedAdminUser(ctx context.Context, email, password string) error {
	exists, err := s.userRepo.ExistsAdminUser(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil // Admin already seeded
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = s.userRepo.Create(ctx, admin)
	return err
}
