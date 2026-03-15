package models

import "time"

type User struct {
	ID           string    `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	Name         string    `db:"name"          json:"name"`
	PasswordHash string    `db:"password_hash" json:"-"`
	APIKey       string    `db:"api_key"       json:"-"`
	Role         string    `db:"role"          json:"role"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Name     string `json:"name"     binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
