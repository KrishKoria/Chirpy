package auth

import (
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
)

func TestMakeAndValidateJWT(t *testing.T) {
    // Setup
    userID := uuid.New()
    secret := "test-secret-key"
    expiresIn := time.Hour * 24

    // Create token
    token, err := MakeJWT(userID, secret, expiresIn)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    // Validate token
    extractedID, err := ValidateJWT(token, secret)
    assert.NoError(t, err)
    assert.Equal(t, userID, extractedID)
}

func TestExpiredJWT(t *testing.T) {
    // Setup with very short expiration
    userID := uuid.New()
    secret := "test-secret-key"
    expiresIn := time.Millisecond * 1 // Expire immediately

    // Create token
    token, err := MakeJWT(userID, secret, expiresIn)
    assert.NoError(t, err)

    // Wait for token to expire
    time.Sleep(time.Millisecond * 10)

    // Validate token - should fail
    _, err = ValidateJWT(token, secret)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "token is expired")
}

func TestInvalidSecretJWT(t *testing.T) {
    // Setup
    userID := uuid.New()
    correctSecret := "correct-secret"
    wrongSecret := "wrong-secret"
    expiresIn := time.Hour

    // Create token with correct secret
    token, err := MakeJWT(userID, correctSecret, expiresIn)
    assert.NoError(t, err)

    // Validate with wrong secret - should fail
    _, err = ValidateJWT(token, wrongSecret)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "signature is invalid")
}

func TestInvalidTokenFormat(t *testing.T) {
    // Test with malformed token
    _, err := ValidateJWT("not-a-valid-jwt-token", "any-secret")
    assert.Error(t, err)
}

func TestHashAndCheckPassword(t *testing.T) {
    // Test password hashing and verification
    password := "secure-password-123"
    
    // Hash the password
    hash, err := HashPassword(password)
    assert.NoError(t, err)
    assert.NotEmpty(t, hash)
    
    // Verify correct password
    err = CheckPasswordHash(password, hash)
    assert.NoError(t, err)
    
    // Verify incorrect password
    err = CheckPasswordHash("wrong-password", hash)
    assert.Error(t, err)
}