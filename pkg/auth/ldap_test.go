package auth

import (
	"errors"
	"testing"

	"github.com/zxh326/kite/pkg/model"
)

func TestNewLDAPConfig(t *testing.T) {
	t.Run("disabled setting", func(t *testing.T) {
		_, err := newLDAPConfig(&model.LDAPSetting{Enabled: false})
		if !errors.Is(err, ErrLDAPDisabled) {
			t.Fatalf("newLDAPConfig() error = %v, want ErrLDAPDisabled", err)
		}
	})

	t.Run("invalid setting", func(t *testing.T) {
		_, err := newLDAPConfig(&model.LDAPSetting{Enabled: true})
		if !errors.Is(err, ErrLDAPNotConfigured) {
			t.Fatalf("newLDAPConfig() error = %v, want ErrLDAPNotConfigured", err)
		}
	})

	t.Run("valid setting", func(t *testing.T) {
		setting := &model.LDAPSetting{
			Enabled:              true,
			ServerURL:            "ldap://ldap.example.com",
			BindDN:               "cn=admin,dc=example,dc=com",
			BindPassword:         "secret",
			UserBaseDN:           "ou=users,dc=example,dc=com",
			UserFilter:           "(uid=%s)",
			UsernameAttribute:    "uid",
			DisplayNameAttribute: "cn",
			GroupBaseDN:          "ou=groups,dc=example,dc=com",
			GroupFilter:          "(member=%s)",
			GroupNameAttribute:   "cn",
		}

		got, err := newLDAPConfig(setting)
		if err != nil {
			t.Fatalf("newLDAPConfig() error = %v", err)
		}
		if got.ServerURL != setting.ServerURL || got.BindDN != setting.BindDN || got.BindPassword != string(setting.BindPassword) {
			t.Fatalf("newLDAPConfig() returned unexpected config: %#v", got)
		}
	})
}

func TestFormatLDAPFilter(t *testing.T) {
	got, err := formatLDAPFilter("(uid=%s)", "alice")
	if err != nil {
		t.Fatalf("formatLDAPFilter() error = %v", err)
	}
	if got != "(uid=alice)" {
		t.Fatalf("formatLDAPFilter() = %q, want %q", got, "(uid=alice)")
	}

	if _, err := formatLDAPFilter("(uid=%s)(mail=%s)", "alice"); !errors.Is(err, ErrLDAPNotConfigured) {
		t.Fatalf("formatLDAPFilter() error = %v, want ErrLDAPNotConfigured", err)
	}
}
