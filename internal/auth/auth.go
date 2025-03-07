package auth

import (
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "errors"
    "fmt"
    "time"
    "net/http"
    "strings"
    "crypto/rand"
    "encoding/hex"
)

func HashPassword(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    
    return string(hashedBytes), nil
}

func CheckPasswordHash(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}


func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
        Issuer: "chirpy",
        IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
        ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
        Subject: userID.String(),
    })
    
    return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &jwt.RegisteredClaims{},
        func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(tokenSecret), nil
        },
    )

    if err != nil {
        return uuid.Nil, err
    }

    if !token.Valid {
        return uuid.Nil, errors.New("invalid token")
    }
    
    claims, ok := token.Claims.(*jwt.RegisteredClaims)
    if !ok {
        return uuid.Nil, errors.New("invalid token claims")
    }
    
    userID, err := uuid.Parse(claims.Subject)
    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
    }
    
    return userID, nil
}


func GetBearerToken(headers http.Header) (string, error) {
    authHeader := headers.Get("Authorization")
    if authHeader == "" {
        return "", errors.New("authorization header is missing")
    }

    const prefix = "Bearer "
    if !strings.HasPrefix(authHeader, prefix) {
        return "", errors.New("authorization header must start with 'Bearer '")
    }

    token := strings.TrimPrefix(authHeader, prefix)
    token = strings.TrimSpace(token)
    
    if token == "" {
        return "", errors.New("token is empty")
    }

    return token, nil
}


func MakeRefreshToken() (string, error) {
    tokenBytes := make([]byte, 32)
    _, err := rand.Read(tokenBytes)
    if err != nil {
        return "", err
    }
    
    return hex.EncodeToString(tokenBytes), nil
}