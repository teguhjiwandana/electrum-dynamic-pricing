package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// mock UserLookup
// ---------------------------------------------------------------------------

type mockUserLookup struct {
	user *User
	err  error
}

func (m *mockUserLookup) GetByUsername(_ context.Context, _ string) (*User, error) {
	return m.user, m.err
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestService() *JWTService {
	return NewJWTService(&mockUserLookup{})
}

// ---------------------------------------------------------------------------
// 1. TestGenerateAndValidateToken
// ---------------------------------------------------------------------------

func TestGenerateAndValidateToken(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name     string
		username string
		role     string
	}{
		{"admin", "admin", "admin"},
		{"user", "john_doe", "user"},
		{"moderator", "mod", "moderator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, exp, err := svc.GenerateToken(tt.username, tt.role)
			if err != nil {
				t.Fatalf("GenerateToken failed: %v", err)
			}
			if tok == "" {
				t.Fatal("expected non-empty token")
			}
			if exp <= time.Now().Unix() {
				t.Fatal("exp should be in the future")
			}

			claims, err := svc.ValidateToken(tok)
			if err != nil {
				t.Fatalf("ValidateToken failed: %v", err)
			}

			gotUsername, _ := (*claims)["username"].(string)
			gotRole, _ := (*claims)["role"].(string)

			if gotUsername != tt.username {
				t.Errorf("username = %q, want %q", gotUsername, tt.username)
			}
			if gotRole != tt.role {
				t.Errorf("role = %q, want %q", gotRole, tt.role)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. TestValidateToken_Invalid
// ---------------------------------------------------------------------------

func TestValidateToken_Invalid(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"garbage", "garbage"},
		{"malformed JWT", "eyJ.abc.xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateToken(tt.token)
			if err == nil {
				t.Errorf("expected error for token %q", tt.token)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. TestValidateToken_Expired
// ---------------------------------------------------------------------------

func TestValidateToken_Expired(t *testing.T) {
	svc := newTestService()

	claims := jwt.MapClaims{
		"username": "test",
		"role":     "user",
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
		"exp":      time.Now().Add(-1 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokStr, err := tok.SignedString(svc.secret)
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	_, err = svc.ValidateToken(tokStr)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

// ---------------------------------------------------------------------------
// 4. TestBcryptHashAndVerify
// ---------------------------------------------------------------------------

func TestBcryptHashAndVerify(t *testing.T) {
	const password = "testpass"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	// correct password
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		t.Errorf("correct password should verify: %v", err)
	}

	// wrong password
	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrongpass")); err == nil {
		t.Error("wrong password should fail verification")
	}
}

// ---------------------------------------------------------------------------
// 5. TestAuthenticate_Success
// ---------------------------------------------------------------------------

func TestAuthenticate_Success(t *testing.T) {
	const password = "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	svc := &JWTService{
		userRepo: &mockUserLookup{
			user: &User{
				ID:           1,
				Username:     "admin",
				PasswordHash: string(hash),
				Role:         "admin",
			},
		},
		secret: []byte("electrum-jwt-secret"),
	}

	token, exp, username, role, err := svc.Authenticate(context.Background(), "admin", password)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if exp <= time.Now().Unix() {
		t.Error("exp should be in the future")
	}
	if username != "admin" {
		t.Errorf("username = %q, want %q", username, "admin")
	}
	if role != "admin" {
		t.Errorf("role = %q, want %q", role, "admin")
	}

	// validate the returned token
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken on generated token failed: %v", err)
	}
	if got, _ := (*claims)["username"].(string); got != "admin" {
		t.Errorf("token username = %q, want %q", got, "admin")
	}
	if got, _ := (*claims)["role"].(string); got != "admin" {
		t.Errorf("token role = %q, want %q", got, "admin")
	}
}

// ---------------------------------------------------------------------------
// 6. TestAuthenticate_UserNotFound
// ---------------------------------------------------------------------------

func TestAuthenticate_UserNotFound(t *testing.T) {
	svc := &JWTService{
		userRepo: &mockUserLookup{user: nil, err: nil},
		secret:   []byte("electrum-jwt-secret"),
	}

	_, _, _, _, err := svc.Authenticate(context.Background(), "ghost", "pass")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 7. TestAuthenticate_WrongPassword
// ---------------------------------------------------------------------------

func TestAuthenticate_WrongPassword(t *testing.T) {
	const correctPassword = "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	svc := &JWTService{
		userRepo: &mockUserLookup{
			user: &User{
				ID:           1,
				Username:     "admin",
				PasswordHash: string(hash),
				Role:         "admin",
			},
		},
		secret: []byte("electrum-jwt-secret"),
	}

	_, _, _, _, err = svc.Authenticate(context.Background(), "admin", "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' error, got: %v", err)
	}
}
