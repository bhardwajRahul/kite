package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/zxh326/kite/pkg/model"
)

func TestDiscoverOAuthEndpoints(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"userinfo_endpoint":      server.URL + "/userinfo",
		})
	}))
	t.Cleanup(server.Close)

	meta, err := discoverOAuthEndpoints(server.URL, "test-provider")
	if err != nil {
		t.Fatalf("discoverOAuthEndpoints() error = %v", err)
	}
	if meta.AuthURL != server.URL+"/authorize" || meta.TokenURL != server.URL+"/token" || meta.UserInfoURL != server.URL+"/userinfo" {
		t.Fatalf("discoverOAuthEndpoints() = %#v", meta)
	}
}

func TestNewGenericProviderAndAuthURL(t *testing.T) {
	provider, err := NewGenericProvider(model.OAuthProvider{
		Name:         "github",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthURL:      "https://example.com/oauth/authorize",
		TokenURL:     "https://example.com/oauth/token",
		UserInfoURL:  "https://example.com/oauth/userinfo",
		Scopes:       "openid,profile,email",
	})
	if err != nil {
		t.Fatalf("NewGenericProvider() error = %v", err)
	}
	if provider.Config.Scopes != "openid profile email" {
		t.Fatalf("provider.Config.Scopes = %q, want %q", provider.Config.Scopes, "openid profile email")
	}

	authURL := provider.GetAuthURL("state-value")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "example.com" || parsed.Path != "/oauth/authorize" {
		t.Fatalf("GetAuthURL() parsed URL = %s", parsed.String())
	}

	query := parsed.Query()
	if query.Get("client_id") != "client-id" || query.Get("redirect_uri") != "" || query.Get("scope") != "openid profile email" || query.Get("state") != "state-value" || query.Get("response_type") != "code" {
		t.Fatalf("GetAuthURL() query = %v", query)
	}
}
