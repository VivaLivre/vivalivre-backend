package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	jwtSecret = []byte("unit-test-secret-with-at-least-32-chars")
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("Password123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "Password123" {
		t.Fatal("password was stored in plain text")
	}
	if !CheckPassword("Password123", hash) {
		t.Error("CheckPassword rejected the correct password")
	}
	if CheckPassword("wrong-password", hash) {
		t.Error("CheckPassword accepted a wrong password")
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	token, err := GenerateToken(42, "admin")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	userID, role, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if userID != 42 || role != "admin" {
		t.Errorf("claims mismatch: got userID=%d role=%q", userID, role)
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	claims := jwt.MapClaims{"user_id": 1, "role": "user", "exp": time.Now().Add(time.Hour).Unix()}
	forged, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("another-secret"))

	if _, _, err := ValidateToken(forged); err == nil {
		t.Error("token signed with another secret was accepted")
	}
}

func TestValidateTokenRejectsExpired(t *testing.T) {
	claims := jwt.MapClaims{"user_id": 1, "role": "user", "exp": time.Now().Add(-time.Minute).Unix()}
	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)

	if _, _, err := ValidateToken(expired); err == nil {
		t.Error("expired token was accepted")
	}
}

func TestValidateTokenRejectsNoneAlgorithm(t *testing.T) {
	claims := jwt.MapClaims{"user_id": 1, "role": "admin", "exp": time.Now().Add(time.Hour).Unix()}
	unsigned, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)

	if _, _, err := ValidateToken(unsigned); err == nil {
		t.Error("token with alg=none was accepted")
	}
}
