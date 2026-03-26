package auth

import (
	"testing"

	"github.com/zxh326/kite/pkg/model"
)

func TestLDAPSettingResponse(t *testing.T) {
	setting := &model.LDAPSetting{
		Enabled:              true,
		ServerURL:            "ldap://ldap.example.com",
		UseStartTLS:          true,
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

	got := ldapSettingResponse(setting)

	if got["bindPassword"] != "" {
		t.Fatalf("bindPassword = %v, want empty string", got["bindPassword"])
	}
	if got["bindPasswordConfigured"] != true {
		t.Fatalf("bindPasswordConfigured = %v, want true", got["bindPasswordConfigured"])
	}
	if got["serverUrl"] != setting.ServerURL || got["bindDn"] != setting.BindDN {
		t.Fatalf("ldapSettingResponse() returned unexpected fields: %#v", got)
	}
}

func TestMergeLDAPSetting(t *testing.T) {
	current := &model.LDAPSetting{
		Enabled:              false,
		ServerURL:            "ldap://old.example.com",
		UseStartTLS:          false,
		BindDN:               "cn=old,dc=example,dc=com",
		BindPassword:         "old-secret",
		UserBaseDN:           "ou=old-users,dc=example,dc=com",
		UserFilter:           "(uid=%s)",
		UsernameAttribute:    "uid",
		DisplayNameAttribute: "cn",
		GroupBaseDN:          "ou=old-groups,dc=example,dc=com",
		GroupFilter:          "(member=%s)",
		GroupNameAttribute:   "cn",
	}
	enabled := true
	serverURL := "  ldaps://ldap.example.com  "
	bindDN := "  cn=admin,dc=example,dc=com  "
	bindPassword := "new-secret"
	userBaseDN := " ou=users,dc=example,dc=com "
	groupBaseDN := " ou=groups,dc=example,dc=com "

	got := mergeLDAPSetting(current, UpdateLDAPSettingRequest{
		Enabled:      &enabled,
		ServerURL:    &serverURL,
		UseStartTLS:  &enabled,
		BindDN:       &bindDN,
		BindPassword: &bindPassword,
		UserBaseDN:   &userBaseDN,
		GroupBaseDN:  &groupBaseDN,
	})

	if !got.Enabled {
		t.Fatalf("Enabled = false, want true")
	}
	if got.ServerURL != "ldaps://ldap.example.com" {
		t.Fatalf("ServerURL = %q, want trimmed URL", got.ServerURL)
	}
	if got.BindDN != "cn=admin,dc=example,dc=com" {
		t.Fatalf("BindDN = %q, want trimmed DN", got.BindDN)
	}
	if got.BindPassword != model.SecretString("new-secret") {
		t.Fatalf("BindPassword = %q, want updated secret", got.BindPassword)
	}
	if got.UserBaseDN != "ou=users,dc=example,dc=com" {
		t.Fatalf("UserBaseDN = %q, want trimmed base DN", got.UserBaseDN)
	}
	if got.GroupBaseDN != "ou=groups,dc=example,dc=com" {
		t.Fatalf("GroupBaseDN = %q, want trimmed base DN", got.GroupBaseDN)
	}
	if got.UserFilter != "(uid=%s)" || got.GroupFilter != "(member=%s)" {
		t.Fatalf("defaults should be preserved, got %#v", got)
	}
}
