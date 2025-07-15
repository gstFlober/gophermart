package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gophemart/internal/app/entity"
	"gophemart/internal/app/repository"
	"gophemart/pkg/logger"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials format")
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (*entity.User, error) {
	logger.Info().
		Str("method", "Register").
		Str("login", login).
		Msg("Starting user registration")

	existingUser, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil && !errors.Is(err, repository.ErrRocordNotFound) {
		logger.Error().
			Err(err).
			Str("login", login).
			Msg("Error checking user existence")
		return nil, err
	}
	if existingUser != nil {
		logger.Warn().
			Str("login", login).
			Msg("User already exists, registration aborted")
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Msg("Password hashing failed")
		return nil, err
	}
	user := &entity.User{
		ID:           uint(uuid.New().ID()),
		Login:        login,
		PasswordHash: string(hashedPassword),
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Msg("Failed to create user in database")
		return nil, err
	}

	logger.Info().
		Str("user_id", fmt.Sprintf("%d", user.ID)).
		Str("login", login).
		Msg("User successfully registered")
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*entity.User, error) {
	logger.Info().
		Str("method", "Login").
		Str("login", login).
		Msg("Attempting user login")
	user, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Msg("Database error during login")
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		logger.Warn().
			Str("user_id", fmt.Sprintf("%d", user.ID)).
			Str("login", login).
			Msg("Invalid password provided")
		return nil, ErrInvalidCredentials
	}
	logger.Info().
		Str("user_id", fmt.Sprintf("%d", user.ID)).
		Str("login", login).
		Msg("User successfully authenticated")

	return user, nil
}
