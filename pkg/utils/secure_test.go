package utils

import "testing"

func TestEncryptDecryptString(t *testing.T) {
	original := "Hello, World!asdas"
	encrypted := EncryptString(original)
	decrypted, err := DecryptString(encrypted)
	t.Log("Encrypted:", encrypted)
	t.Log("Decrypted:", decrypted)
	if err != nil {
		t.Fatalf("DecryptString() failed: %v", err)
	}
	if decrypted != original {
		t.Errorf("DecryptString() = %q, want %q", decrypted, original)
	}
}

func TestDecryptStringInvalidInput(t *testing.T) {
	if _, err := DecryptString("not-base64"); err == nil {
		t.Fatal("DecryptString() error = nil, want error")
	}
}
