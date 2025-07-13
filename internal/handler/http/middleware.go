package http

import (
	"fmt"
	"github.com/labstack/echo"
	"gophemart/pkg/jwt"
	"gophemart/pkg/logger"
	"net/http"
	"strconv"
	"time"
)

var (
	authCookieName = "auth_token"
	userIDKey      = "userID"
	user_id        = "user_id"
)

func AuthMiddleware(jwtManager *jwt.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			path := c.Path()
			method := c.Request().Method
			ip := c.RealIP()
			cookie, err := c.Cookie(authCookieName)
			if err != nil {
				logger.Error().
					Err(err).
					Str("path", path).
					Str("method", method).
					Str("ip", ip).
					Msg("Auth cookie not found in request")

				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "authentication required",
				})
			}

			claims, err := jwtManager.ValidateToken(cookie.Value)
			if err != nil {
				logger.Error().
					Err(err).
					Str("path", path).
					Str("method", method).
					Str("ip", ip).
					Msg("JWT token validation failed")

				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "invalid token",
				})
			}
			var userID string

			switch v := claims[user_id].(type) {
			case string:
				userID = v
			case float64:
				userID = strconv.FormatFloat(v, 'f', 0, 64)
			case int, int32, int64:
				userID = fmt.Sprintf("%d", v)
			default:
				logger.Error().
					Str("path", path).
					Str("method", method).
					Str("ip", ip).
					Str("type", fmt.Sprintf("%T", claims[user_id])).
					Interface("claims", claims).
					Msg("Invalid userID type in token claims")

				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "invalid user in token",
				})
			}
			if userID == "" {
				logger.Error().
					Str("path", path).
					Str("method", method).
					Str("ip", ip).
					Interface("claims", claims).
					Msg("Empty userID in token claims")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "invalid user in token",
				})
			}
			c.Set(userIDKey, userID)

			logger.Info().
				Str("user_id", userID).
				Str("path", path).
				Str("method", method).
				Str("ip", ip).
				Dur("duration_ms", time.Since(start)).
				Msg("User authenticated successfully")

			return next(c)
		}
	}
}
