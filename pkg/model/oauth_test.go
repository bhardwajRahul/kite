package model

import "testing"

func TestNormalizeOAuthProviderName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trim and lower", "  GitHub  ", "github"},
		{"empty", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeOAuthProviderName(tt.input); got != tt.expected {
				t.Fatalf("NormalizeOAuthProviderName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsReservedOAuthProviderName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"password", "password", true},
		{"ldap", " LDAP ", true},
		{"custom", "github", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReservedOAuthProviderName(tt.input); got != tt.want {
				t.Fatalf("IsReservedOAuthProviderName() = %v, want %v", got, tt.want)
			}
		})
	}
}
