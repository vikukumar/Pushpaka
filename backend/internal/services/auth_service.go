package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/pushpaka/internal/config"
	"github.com/yourusername/pushpaka/internal/models"
	"github.com/yourusername/pushpaka/internal/repositories"
)

var (
	ErrUserExists         = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	userRepo *repositories.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hash),
		APIKey:       uuid.New().String(),
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid subject claim")
	}
	return userID, nil
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(s.cfg.JWTExpiry) * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
