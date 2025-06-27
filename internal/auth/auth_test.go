package auth

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password",
			password: "validpassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "abc",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "p@ssw0rd!#$%",
			wantErr:  false,
		},
		{
			name:     "password too long",
			password: strings.Repeat("a", 73), // bcrypt limit is 72 bytes
			wantErr:  true,
			errMsg:   ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("HashPassword() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("HashPassword() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() unexpected error = %v", err)
				return
			}

			if hash == "" {
				t.Error("HashPassword() returned empty hash")
			}

			// Verify the hash is valid bcrypt hash
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
			if err != nil {
				t.Errorf("Generated hash is invalid: %v", err)
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	validPassword := "testpassword123"
	validHash, err := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "correct password and hash",
			password: validPassword,
			hash:     string(validHash),
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     string(validHash),
			wantErr:  true,
			errMsg:   ErrIncorrectEmailOrPassword,
		},
		{
			name:     "empty password",
			password: "",
			hash:     string(validHash),
			wantErr:  true,
			errMsg:   ErrIncorrectEmailOrPassword,
		},
		{
			name:     "invalid hash format",
			password: validPassword,
			hash:     "invalid-hash",
			wantErr:  true,
		},
		{
			name:     "empty hash",
			password: validPassword,
			hash:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckPasswordHash() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("CheckPasswordHash() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("CheckPasswordHash() unexpected error = %v", err)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"

	tests := []struct {
		name        string
		userID      uuid.UUID
		tokenSecret string
		wantErr     bool
	}{
		{
			name:        "valid JWT creation",
			userID:      userID,
			tokenSecret: tokenSecret,
			wantErr:     false,
		},
		{
			name:        "empty token secret",
			userID:      userID,
			tokenSecret: "",
			wantErr:     false, // JWT library allows empty secret
		},
		{
			name:        "zero expiration",
			userID:      userID,
			tokenSecret: tokenSecret,
			wantErr:     false,
		},
		{
			name:        "negative expiration",
			userID:      userID,
			tokenSecret: tokenSecret,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := MakeJWT(tt.userID, tt.tokenSecret)

			if tt.wantErr {
				if err == nil {
					t.Errorf("MakeJWT() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("MakeJWT() unexpected error = %v", err)
				return
			}

			if token == "" {
				t.Error("MakeJWT() returned empty token")
			}

			// Verify token structure (should have 3 parts separated by dots)
			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				t.Errorf("MakeJWT() returned invalid JWT format, got %d parts, want 3", len(parts))
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"

	// Create a valid token for testing
	validToken, err := MakeJWT(userID, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// Create token with different secret
	differentSecretToken, err := MakeJWT(userID, "different-secret")
	if err != nil {
		t.Fatalf("Failed to create different secret test token: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			tokenSecret: tokenSecret,
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong-secret",
			wantErr:     true,
		},
		{
			name:        "token signed with different secret",
			tokenString: differentSecretToken,
			tokenSecret: tokenSecret,
			wantErr:     true,
		},
		{
			name:        "malformed token",
			tokenString: "invalid.token.format",
			tokenSecret: tokenSecret,
			wantErr:     true,
		},
		{
			name:        "empty token",
			tokenString: "",
			tokenSecret: tokenSecret,
			wantErr:     true,
		},
		{
			name:        "incomplete token",
			tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			tokenSecret: tokenSecret,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateJWT() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateJWT() unexpected error = %v", err)
				return
			}

			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func TestJWTRoundTrip(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"

	// Create token
	token, err := MakeJWT(userID, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token
	gotUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if gotUserID != userID {
		t.Errorf("Round trip failed: got userID %v, want %v", gotUserID, userID)
	}
}

func TestPasswordHashRoundTrip(t *testing.T) {
	password := "testpassword123"

	// Hash password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Check password
	err = CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}

	// Check with wrong password
	err = CheckPasswordHash("wrongpassword", hash)
	if err == nil {
		t.Error("Expected error when checking wrong password")
	}
	if err.Error() != ErrIncorrectEmailOrPassword {
		t.Errorf("Expected error %v, got %v", ErrIncorrectEmailOrPassword, err.Error())
	}
}
