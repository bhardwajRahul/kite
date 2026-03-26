package model

import "testing"

func TestLDAPSetting_Normalized(t *testing.T) {
	setting := LDAPSetting{
		Model:                Model{ID: 7},
		Enabled:              true,
		ServerURL:            " ldap://ldap.example.com ",
		UseStartTLS:          true,
		BindDN:               " cn=admin,dc=example,dc=com ",
		BindPassword:         SecretString("secret"),
		UserBaseDN:           " ou=users,dc=example,dc=com ",
		UserFilter:           " ",
		UsernameAttribute:    " ",
		DisplayNameAttribute: " displayName ",
		GroupBaseDN:          " ou=groups,dc=example,dc=com ",
		GroupFilter:          " ",
		GroupNameAttribute:   " ",
	}

	normalized := setting.Normalized()
	if normalized.ID != 7 {
		t.Fatalf("Normalized().ID = %d, want 7", normalized.ID)
	}
	if normalized.ServerURL != "ldap://ldap.example.com" {
		t.Fatalf("Normalized().ServerURL = %q", normalized.ServerURL)
	}
	if normalized.BindDN != "cn=admin,dc=example,dc=com" {
		t.Fatalf("Normalized().BindDN = %q", normalized.BindDN)
	}
	if normalized.UserFilter != DefaultLDAPUserFilter {
		t.Fatalf("Normalized().UserFilter = %q, want %q", normalized.UserFilter, DefaultLDAPUserFilter)
	}
	if normalized.UsernameAttribute != DefaultLDAPUsernameAttribute {
		t.Fatalf("Normalized().UsernameAttribute = %q, want %q", normalized.UsernameAttribute, DefaultLDAPUsernameAttribute)
	}
	if normalized.DisplayNameAttribute != "displayName" {
		t.Fatalf("Normalized().DisplayNameAttribute = %q", normalized.DisplayNameAttribute)
	}
	if normalized.GroupFilter != DefaultLDAPGroupFilter {
		t.Fatalf("Normalized().GroupFilter = %q, want %q", normalized.GroupFilter, DefaultLDAPGroupFilter)
	}
	if normalized.GroupNameAttribute != DefaultLDAPGroupNameAttribute {
		t.Fatalf("Normalized().GroupNameAttribute = %q, want %q", normalized.GroupNameAttribute, DefaultLDAPGroupNameAttribute)
	}
}

func TestLDAPSetting_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setting LDAPSetting
		wantErr bool
	}{
		{
			name: "disabled",
			setting: LDAPSetting{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "valid",
			setting: LDAPSetting{
				Enabled:              true,
				ServerURL:            "ldap://ldap.example.com",
				BindDN:               "cn=admin,dc=example,dc=com",
				BindPassword:         SecretString("secret"),
				UserBaseDN:           "ou=users,dc=example,dc=com",
				UserFilter:           "(uid=%s)",
				GroupBaseDN:          "ou=groups,dc=example,dc=com",
				GroupFilter:          "(member=%s)",
				UsernameAttribute:    "uid",
				DisplayNameAttribute: "cn",
				GroupNameAttribute:   "cn",
			},
			wantErr: false,
		},
		{
			name: "invalid url",
			setting: LDAPSetting{
				Enabled:      true,
				ServerURL:    "https://ldap.example.com",
				BindDN:       "cn=admin,dc=example,dc=com",
				BindPassword: SecretString("secret"),
				UserBaseDN:   "ou=users,dc=example,dc=com",
				UserFilter:   "(uid=%s)",
				GroupBaseDN:  "ou=groups,dc=example,dc=com",
				GroupFilter:  "(member=%s)",
			},
			wantErr: true,
		},
		{
			name: "invalid filter",
			setting: LDAPSetting{
				Enabled:      true,
				ServerURL:    "ldap://ldap.example.com",
				BindDN:       "cn=admin,dc=example,dc=com",
				BindPassword: SecretString("secret"),
				UserBaseDN:   "ou=users,dc=example,dc=com",
				UserFilter:   "(uid=%s)(extra=%s)",
				GroupBaseDN:  "ou=groups,dc=example,dc=com",
				GroupFilter:  "(member=%s)",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setting.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasExactlyOneLDAPPlaceholder(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"exactly one", "(uid=%s)", true},
		{"escaped percent", "(uid=%%s)", false},
		{"two placeholders", "(uid=%s)(member=%s)", false},
		{"no placeholder", "(uid=admin)", false},
		{"invalid verb", "(uid=%d)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasExactlyOneLDAPPlaceholder(tt.in); got != tt.want {
				t.Fatalf("HasExactlyOneLDAPPlaceholder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLDAPSetting_BindPasswordConfigured(t *testing.T) {
	var nilSetting *LDAPSetting
	if nilSetting.BindPasswordConfigured() {
		t.Fatal("nil setting reported configured bind password")
	}

	if got := (&LDAPSetting{}).BindPasswordConfigured(); got {
		t.Fatal("empty bind password reported configured")
	}

	if got := (&LDAPSetting{BindPassword: SecretString("secret")}).BindPasswordConfigured(); !got {
		t.Fatal("configured bind password reported missing")
	}
}

func TestNormalizeLDAPTextWithDefault(t *testing.T) {
	if got := normalizeLDAPTextWithDefault("  value  ", "fallback"); got != "value" {
		t.Fatalf("normalizeLDAPTextWithDefault() = %q, want %q", got, "value")
	}
	if got := normalizeLDAPTextWithDefault("   ", "fallback"); got != "fallback" {
		t.Fatalf("normalizeLDAPTextWithDefault() = %q, want %q", got, "fallback")
	}
}

func TestValidateLDAPServerURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ldap", "ldap://ldap.example.com", false},
		{"ldaps", "ldaps://ldap.example.com", false},
		{"invalid scheme", "https://ldap.example.com", true},
		{"missing host", "ldap://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLDAPServerURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateLDAPServerURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
