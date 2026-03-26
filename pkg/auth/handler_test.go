package auth

import (
	"errors"
	"fmt"
	"testing"
)

func TestUniqueStrings(t *testing.T) {
	got := uniqueStrings([]string{"oauth", "ldap", "oauth", "password", "ldap"})
	want := []string{"oauth", "ldap", "password"}

	if len(got) != len(want) {
		t.Fatalf("uniqueStrings() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("uniqueStrings() got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestIsCredentialFailure(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "invalid credentials", err: errInvalidCredentials, want: true},
		{name: "ldap invalid credentials", err: ErrLDAPInvalidCredentials, want: true},
		{name: "ldap disabled", err: ErrLDAPDisabled, want: true},
		{name: "ldap not configured", err: ErrLDAPNotConfigured, want: true},
		{name: "wrapped invalid credentials", err: fmt.Errorf("wrap: %w", errInvalidCredentials), want: true},
		{name: "other error", err: errors.New("other"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCredentialFailure(tt.err); got != tt.want {
				t.Fatalf("isCredentialFailure() = %v, want %v", got, tt.want)
			}
		})
	}
}
