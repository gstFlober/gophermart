package jwt

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenFlow(t *testing.T) {
	secret := "super-secret-key"
	duration := 15 * time.Minute

	manager := NewManager(secret, duration)

	userID := uint(12345)
	token, err := manager.GenerateToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	_, err = manager.ValidateToken(token)
	require.NoError(t, err)
}

func TestInvalidToken(t *testing.T) {
	manager := NewManager("secret", time.Minute)

	invalidManager := NewManager("different-secret", time.Minute)

	userID := uint(100)
	token, err := manager.GenerateToken(userID)
	require.NoError(t, err)

	t.Run("wrong signature", func(t *testing.T) {
		_, err := invalidManager.ValidateToken(token)
		assert.ErrorContains(t, err, "signature")
	})

}
