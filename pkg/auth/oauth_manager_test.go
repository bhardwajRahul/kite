package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zxh326/kite/pkg/common"
	"github.com/zxh326/kite/pkg/model"
)

func TestGetRequestHost(t *testing.T) {
	originalHost := common.Host
	t.Cleanup(func() {
		common.Host = originalHost
	})

	t.Run("configured host overrides request", func(t *testing.T) {
		common.Host = "https://kite.example.com"
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "http://ignored.local", nil)

		if got := getRequestHost(c); got != common.Host {
			t.Fatalf("getRequestHost() = %q, want %q", got, common.Host)
		}
	})

	t.Run("forwarded headers", func(t *testing.T) {
		common.Host = ""
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		req := httptest.NewRequest(http.MethodGet, "http://internal.local", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "kite.example.com")
		c.Request = req

		if got := getRequestHost(c); got != "https://kite.example.com" {
			t.Fatalf("getRequestHost() = %q, want %q", got, "https://kite.example.com")
		}
	})

	t.Run("request host and scheme", func(t *testing.T) {
		common.Host = ""
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		req := httptest.NewRequest(http.MethodGet, "http://kite.local", nil)
		c.Request = req

		if got := getRequestHost(c); got != "http://kite.local" {
			t.Fatalf("getRequestHost() = %q, want %q", got, "http://kite.local")
		}
	})

	t.Run("fallback", func(t *testing.T) {
		common.Host = ""
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "http://kite.local", nil)
		c.Request.Host = ""

		if got := getRequestHost(c); got != "http://localhost" {
			t.Fatalf("getRequestHost() = %q, want %q", got, "http://localhost")
		}
	})
}

func TestGenerateAndValidateJWT(t *testing.T) {
	originalSecret := common.JwtSecret
	common.JwtSecret = "test-secret"
	t.Cleanup(func() {
		common.JwtSecret = originalSecret
	})

	manager := NewOAuthManager()
	user := &model.User{
		Model:    model.Model{ID: 42},
		Username: "alice",
		Provider: model.AuthProviderLDAP,
	}

	tokenString, err := manager.GenerateJWT(user, "refresh-token")
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	claims, err := manager.ValidateJWT(tokenString)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if claims.UserID != user.ID || claims.Username != user.Username || claims.Provider != user.Provider {
		t.Fatalf("ValidateJWT() claims = %#v, want user fields from input", claims)
	}
	if claims.RefreshToken != "refresh-token" {
		t.Fatalf("ValidateJWT() refresh token = %q, want %q", claims.RefreshToken, "refresh-token")
	}
	if claims.Issuer != "Kite" {
		t.Fatalf("ValidateJWT() issuer = %q, want %q", claims.Issuer, "Kite")
	}
}

func TestGenerateJWTStripsLargeRefreshToken(t *testing.T) {
	originalSecret := common.JwtSecret
	common.JwtSecret = "test-secret"
	t.Cleanup(func() {
		common.JwtSecret = originalSecret
	})

	manager := NewOAuthManager()
	user := &model.User{
		Model:    model.Model{ID: 7},
		Username: "bob",
		Provider: "github",
	}
	largeRefreshToken := strings.Repeat("r", maxAuthTokenCookieValueLength+1)

	tokenString, err := manager.GenerateJWT(user, largeRefreshToken)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	claims, err := manager.ValidateJWT(tokenString)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if claims.RefreshToken != "" {
		t.Fatalf("ValidateJWT() refresh token = %q, want empty string", claims.RefreshToken)
	}
}
