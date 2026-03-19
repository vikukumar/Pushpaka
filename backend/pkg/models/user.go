package models

import "github.com/vikukumar/Pushpaka/pkg/basemodel"

type User struct {
	basemodel.BaseModel
	Email        string `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	Name         string `gorm:"type:varchar(255);not null" json:"name"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	APIKey       string `gorm:"uniqueIndex;type:varchar(255)" json:"-"`
	Role         string `gorm:"type:varchar(50);default:'user'" json:"role"`
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
