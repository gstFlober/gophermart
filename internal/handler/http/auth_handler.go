package http

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"gophemart/internal/app/service"
	"gophemart/internal/handler/http/dto"
	"gophemart/pkg/jwt"
	"gophemart/pkg/logger"
	"net/http"
	"time"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtManager  *jwt.Manager
}

func NewAuthHandler(authService *service.AuthService, jwtManager *jwt.Manager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtManager:  jwtManager,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	logger.Info().
		Str("handler", "Register").
		Str("ip", c.RealIP()).
		Msg("Registration request received")

	req := new(dto.RegisterRequest)

	if err := c.Bind(req); err != nil {
		logger.Error().
			Err(err).
			Str("ip", c.RealIP()).
			Msg("Failed to bind registration request")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	logger.Info().
		Str("login", req.Login).
		Str("ip", c.RealIP()).
		Msg("Attempting user registration")

	ctx := c.Request().Context()
	user, err := h.authService.Register(ctx, req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			logger.Warn().
				Err(err).
				Str("login", req.Login).
				Str("ip", c.RealIP()).
				Msg("Registration failed - user already exists")
			return echo.NewHTTPError(http.StatusConflict, "user already exists")
		default:
			logger.Error().
				Err(err).
				Str("login", req.Login).
				Str("ip", c.RealIP()).
				Msg("Internal server error during registration")
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
	}
	token, err := h.jwtManager.GenerateToken(user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", fmt.Sprintf("%d", user.ID)).
			Str("login", req.Login).
			Msg("Failed to generate JWT token")
		return echo.NewHTTPError(http.StatusInternalServerError, "")
	}
	h.setAuthCookie(c, token)
	logger.Info().
		Str("user_id", fmt.Sprintf("%d", user.ID)).
		Str("login", req.Login).
		Msg("User registered successfully")
	return c.JSON(http.StatusOK, dto.RegisterResponse{Token: token})

}
func (h *AuthHandler) Login(c echo.Context) error {
	logger.Info().
		Str("handler", "Login").
		Str("ip", c.RealIP()).
		Msg("Login request received")

	req := new(dto.LoginRequest)
	if err := c.Bind(req); err != nil {
		logger.Error().
			Err(err).
			Str("ip", c.RealIP()).
			Msg("Failed to bind login request")
		return c.JSON(http.StatusBadRequest, "")
	}
	logger.Info().
		Str("login", req.Login).
		Str("ip", c.RealIP()).
		Msg("Attempting user login")
	ctx := c.Request().Context()
	user, err := h.authService.Login(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			logger.Warn().
				Str("login", req.Login).
				Str("ip", c.RealIP()).
				Msg("Invalid login credentials")
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		logger.Error().
			Err(err).
			Str("login", req.Login).
			Str("ip", c.RealIP()).
			Msg("Internal server error during login")
		return echo.NewHTTPError(http.StatusInternalServerError, "authentication error")
	}

	token, err := h.jwtManager.GenerateToken(user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", fmt.Sprintf("%d", user.ID)).
			Str("login", req.Login).
			Msg("Failed to generate JWT token")
		return c.JSON(http.StatusInternalServerError, "")
	}
	h.setAuthCookie(c, token)
	logger.Info().
		Str("user_id", fmt.Sprintf("%d", user.ID)).
		Str("login", req.Login).
		Msg("User logged in successfully")
	return c.JSON(http.StatusOK, dto.LoginResponse{Token: token})
}
func (h *AuthHandler) setAuthCookie(c echo.Context, token string) {
	cookie := new(http.Cookie)
	cookie.Name = authCookieName
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = false
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = time.Now().Add(24 * time.Hour)

	c.SetCookie(cookie)

	logger.Debug().
		Str("cookie_name", authCookieName).
		Time("expires", cookie.Expires).
		Msg("Authentication cookie set")
}
