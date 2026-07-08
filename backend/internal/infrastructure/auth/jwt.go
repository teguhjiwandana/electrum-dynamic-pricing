package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tokenDuration = 24 * time.Hour

// JWTService handles token generation, validation, and user authentication.
type JWTService struct {
	userRepo UserLookup
	secret   []byte
}

// TokenClaims contains the JWT payload data.
type TokenClaims struct {
	Username string
	Role     string
}

// UserLookup is the port for user authentication data access.
type UserLookup interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
}

// User represents an authenticated user.
type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
}

// NewJWTService creates a JWT service with the given user repository.
func NewJWTService(userRepo UserLookup) *JWTService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "electrum-jwt-secret"
	}
	return &JWTService{userRepo: userRepo, secret: []byte(secret)}
}

// GenerateToken creates a signed JWT with username and role claims.
func (s *JWTService) GenerateToken(username, role string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(tokenDuration)
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"iat":      now.Unix(),
		"exp":      exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encoded, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}
	return encoded, exp.Unix(), nil
}

// ValidateToken parses and validates a JWT, returning the claims.
func (s *JWTService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return &claims, nil
}

// Authenticate verifies username/password and returns a JWT if valid.
func (s *JWTService) Authenticate(ctx context.Context, username, password string) (string, int64, string, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("auth error")
	}
	if user == nil {
		return "", 0, "", "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", 0, "", "", fmt.Errorf("invalid credentials")
	}

	token, exp, err := s.GenerateToken(user.Username, user.Role)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("generate token: %w", err)
	}

	return token, exp, user.Username, user.Role, nil
}
