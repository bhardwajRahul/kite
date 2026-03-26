package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = rt.target.Scheme
	clone.URL.Host = rt.target.Host
	clone.Host = rt.target.Host
	return rt.base.RoundTrip(clone)
}

func TestGenericProviderExchangeCodeForToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("Content-Type = %q, want application/x-www-form-urlencoded", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("Accept = %q, want application/json", got)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if r.Form.Get("client_id") != "client-id" || r.Form.Get("client_secret") != "client-secret" || r.Form.Get("code") != "auth-code" || r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("redirect_uri") != "https://kite.example.com/api/auth/callback" {
			t.Fatalf("unexpected form values: %v", r.Form)
		}

		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			Scope:        "openid profile",
		})
	}))
	t.Cleanup(server.Close)

	provider := &GenericProvider{
		Config: OAuthConfig{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURL:  "https://kite.example.com/api/auth/callback",
		},
		TokenURL: server.URL + "/token",
	}

	got, err := provider.ExchangeCodeForToken("auth-code")
	if err != nil {
		t.Fatalf("ExchangeCodeForToken() error = %v", err)
	}
	if got.AccessToken != "access-token" || got.RefreshToken != "refresh-token" || got.TokenType != "Bearer" || got.ExpiresIn != 3600 || got.Scope != "openid profile" {
		t.Fatalf("ExchangeCodeForToken() = %#v", got)
	}
}

func TestGenericProviderRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if r.Form.Get("client_id") != "client-id" || r.Form.Get("client_secret") != "client-secret" || r.Form.Get("refresh_token") != "refresh-token" || r.Form.Get("grant_type") != "refresh_token" {
			t.Fatalf("unexpected form values: %v", r.Form)
		}

		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "",
			TokenType:    "Bearer",
		})
	}))
	t.Cleanup(server.Close)

	provider := &GenericProvider{
		Config: OAuthConfig{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		},
		TokenURL: server.URL + "/token",
	}

	got, err := provider.RefreshToken("refresh-token")
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if got.AccessToken != "new-access-token" || got.RefreshToken != "" || got.TokenType != "Bearer" {
		t.Fatalf("RefreshToken() = %#v", got)
	}
}

func TestGenericProviderGetUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("Authorization = %q, want Bearer access-token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("Accept = %q, want application/json", got)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"oid":          "oid-123",
			"login":        "alice",
			"displayName":  "Alice Display",
			"nickname":     "Alice Nick",
			"picture":      "https://example.com/avatar.png",
			"groups":       []any{"dev", "ops"},
			"unused_field": "ignored",
		})
	}))
	t.Cleanup(server.Close)

	provider := &GenericProvider{
		Name:        "github",
		UserInfoURL: server.URL + "/userinfo",
	}

	got, err := provider.GetUserInfo("access-token")
	if err != nil {
		t.Fatalf("GetUserInfo() error = %v", err)
	}
	if got.Provider != "github" {
		t.Fatalf("Provider = %q, want %q", got.Provider, "github")
	}
	if got.Sub != "oid-123" || got.Username != "alice" || got.Name != "Alice Nick" || got.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("GetUserInfo() user = %#v", got)
	}
	if want := []string{"dev", "ops"}; len(got.OIDCGroups) != len(want) || got.OIDCGroups[0] != want[0] || got.OIDCGroups[1] != want[1] {
		t.Fatalf("OIDCGroups = %#v, want %#v", got.OIDCGroups, want)
	}
}

func TestFetchAzureADGroups(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	requestCount := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("Authorization = %q, want Bearer access-token", got)
		}

		switch r.URL.Query().Get("page") {
		case "":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"value": []map[string]any{
					{"@odata.type": "#microsoft.graph.group", "id": "group-1", "displayName": "Group 1"},
					{"@odata.type": "#microsoft.graph.directoryRole", "id": "role-1", "displayName": "Role 1"},
					{"@odata.type": "#microsoft.graph.group", "displayName": "Missing ID"},
				},
				"@odata.nextLink": server.URL + "/v1.0/me/memberOf?page=2",
			})
		case "2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"value": []map[string]any{
					{"@odata.type": "#microsoft.graph.group", "id": "group-2", "displayName": "Group 2"},
				},
			})
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	t.Cleanup(server.Close)

	targetURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	http.DefaultTransport = rewriteTransport{target: targetURL, base: originalTransport}

	provider := &GenericProvider{}
	got, err := provider.fetchAzureADGroups("access-token")
	if err != nil {
		t.Fatalf("fetchAzureADGroups() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}
	want := []interface{}{"group-1", "group-2"}
	if len(got) != len(want) {
		t.Fatalf("fetchAzureADGroups() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("fetchAzureADGroups()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
