package models

import (
	// "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserRequest_ValidateSuccess(t *testing.T){
	req := CreateUserRequest{
		Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
	}

	valid, msg := req.Validate()

	assert.True(t, valid)
	assert.Empty(t, msg)
}

func TestCreateUserRequest_ValidateMissingEmail(t *testing.T){
	req := CreateUserRequest{
		Email:    "",
        Username: "testuser",
        Password: "password123",
	}

	valid, msg := req.Validate()

	assert.False(t, valid)
	assert.Equal(t, "email is required", msg)
}

func TestCreateUserRequest_ValidateMissingUsername(t *testing.T){
	req := CreateUserRequest{
		Email:    "test@example.com",
        Username: "",
        Password: "password123",
	}

	valid, msg := req.Validate()

	assert.False(t, valid)
	assert.Equal(t, "username is required", msg)
}

func TestCreateUserRequest_ValidateMissingPassword(t *testing.T){
	req := CreateUserRequest{
		Email:    "test@example.com",
        Username: "testuser",
        Password: "",
	}

	valid, msg := req.Validate()

	assert.False(t, valid)
	assert.Equal(t, "password is required", msg)
}

func TestCreateUserRequest_ValidateShortPassword(t *testing.T){
	req := CreateUserRequest{
		Email:    "test@example.com",
        Username: "testuser",
        Password: "12345",
	}

	valid, msg := req.Validate()

	assert.False(t, valid)
	assert.Equal(t, "password must be at least 6 characters", msg)
}

func TestCreateUserRequest_ValidateExactPasswordLength(t *testing.T){
	req := CreateUserRequest{
		Email:    "test@example.com",
        Username: "testuser",
        Password: "123456",
	}

	valid, msg := req.Validate()

    assert.True(t, valid)
    assert.Empty(t, msg)
}


func TestLoginRequest_ValidateSuccess(t *testing.T){
	req := LoginRequest{
		Email:    "test@example.com",
        Password: "password123",	
	}

	valid, msg := req.Validate()

	assert.True(t, valid)
	assert.Empty(t, msg)
}

func TestLoginRequest_ValidateMissingEmail(t *testing.T){
	req := LoginRequest{
		Email:    "",
        Password: "password123",
	}

	valid, msg := req.Validate()

	assert.False(t, valid)
	assert.Equal(t, "email is required", msg)
}

func TestLoginRequest_ValidateMissingPassword(t *testing.T){
	req := LoginRequest{
		Email:    "test@example.com",
        Password: "",	
	}

	valid, msg := req.Validate()

	assert.False(t, valid)
	assert.Equal(t, "password is required",msg)
}

func TestNewUser(t *testing.T){
	email := "test@example.com"
    username := "testuser"
    passwordHash := "hashed_password"

	user := NewUser(email,username,passwordHash)
	// fmt.Println(user)
	assert.NotEmpty(t, user.ID)
    assert.Equal(t, email, user.Email)
    assert.Equal(t, username, user.Username)
    assert.Equal(t, passwordHash, user.PasswordHash)
    assert.NotZero(t, user.CreatedAt)
    assert.NotZero(t, user.UpdatedAt)
    assert.Nil(t, user.DeletedAt)
}