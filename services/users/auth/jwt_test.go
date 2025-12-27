package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T){
	jm := NewJWTManager("test-secret-key")

	//Act
	token, expiresAt, err := jm.GenerateToken("user123", "test@example.com", "testuser", 1*time.Hour)

	// Assert
	assert.NoError(t,err)
	assert.NotEmpty(t,token)
	assert.NotZero(t,expiresAt)
}

func TestValidateToken(t *testing.T){
	jm := NewJWTManager("test-secret-key")

	token, _, _ := jm.GenerateToken("user123", "test@example.com", "testuser", 1*time.Hour)

	claims, err := jm.ValidateToken(token)

	assert.NoError(t,err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "test@example.com",claims.Email)

}