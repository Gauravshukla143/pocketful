package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole defines the role of a user in the system.
type UserRole string

const (
	RoleUser  UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"
)

// User represents a registered user in the system.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"    json:"id"`
	Email        string             `bson:"email"            json:"email"`
	PasswordHash string             `bson:"password_hash"    json:"-"`
	Role         UserRole           `bson:"role"             json:"role"`
	CreatedAt    time.Time          `bson:"created_at"       json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"       json:"updated_at"`
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
