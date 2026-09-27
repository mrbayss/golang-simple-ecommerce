package jwt_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/jwt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTKey() *jwt.Key {
	v := viper.New()
	v.Set("jwt.secret_key", "test-secret-key-1234567890-very-long")
	return jwt.NewJWTToken(v)
}

func TestJWT_GenerateAndVerifyToken(t *testing.T) {
	key := newTestJWTKey()
	userID := "user-uuid-123"
	email := "user@example.com"
	role := "ADMIN"

	t.Run("Successfully generate and verify token", func(t *testing.T) {
		tokenString, claims, err := key.GenerateToken(userID, email, role, time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)
		require.NotNil(t, claims)
		assert.Equal(t, userID, claims.ID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)

		verifiedClaims, err := key.VerifyToken(tokenString)
		require.NoError(t, err)
		require.NotNil(t, verifiedClaims)
		assert.Equal(t, userID, verifiedClaims.ID)
		assert.Equal(t, email, verifiedClaims.Email)
		assert.Equal(t, role, verifiedClaims.Role)
	})

	t.Run("Fail when token is expired", func(t *testing.T) {
		// generate token that is already expired
		tokenString, _, err := key.GenerateToken(userID, email, role, -time.Hour)
		require.NoError(t, err)

		verifiedClaims, err := key.VerifyToken(tokenString)
		require.Error(t, err)
		assert.Nil(t, verifiedClaims)
	})

	t.Run("Fail when token is signed with a different key", func(t *testing.T) {
		otherViper := viper.New()
		otherViper.Set("jwt.secret_key", "different-secret-key")
		otherKey := jwt.NewJWTToken(otherViper)

		tokenString, _, err := otherKey.GenerateToken(userID, email, role, time.Hour)
		require.NoError(t, err)

		verifiedClaims, err := key.VerifyToken(tokenString)
		require.Error(t, err)
		assert.Nil(t, verifiedClaims)
	})

	t.Run("Fail when token string is malformed", func(t *testing.T) {
		verifiedClaims, err := key.VerifyToken("invalid.token.structure")
		require.Error(t, err)
		assert.Nil(t, verifiedClaims)
	})
}

func TestJWT_Concurrency(t *testing.T) {
	key := newTestJWTKey()
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("user-%d", id)
			email := fmt.Sprintf("user%d@example.com", id)
			role := "CUSTOMER"

			token, claims, err := key.GenerateToken(userID, email, role, time.Hour)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.Equal(t, userID, claims.ID)

			verified, err := key.VerifyToken(token)
			assert.NoError(t, err)
			assert.Equal(t, userID, verified.ID)
			assert.Equal(t, email, verified.Email)
		}(i)
	}

	wg.Wait()
}
