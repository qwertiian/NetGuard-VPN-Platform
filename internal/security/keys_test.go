package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateKeyFormat(t *testing.T) {
	validKey := "cI2v/abcdefghijklmnopqrstuvwxyz0123456789+A="
	if err := ValidateKeyFormat(validKey); err != nil {
		t.Errorf("Valid key failed validation: %v", err)
	}

	invalidLength := "cI2v/abcdefghijklmnopqrstuvwxyz0123456789+"
	if err := ValidateKeyFormat(invalidLength); err == nil {
		t.Error("Key with invalid length passed validation")
	}

	invalidChars := "cI2v/abcdefghijklmnopqrstuvwxyz0123456789+!"
	if err := ValidateKeyFormat(invalidChars); err == nil {
		t.Error("Key with invalid characters passed validation")
	}
}

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if err := ValidateKeyFormat(kp.PrivateKey); err != nil {
		t.Errorf("Generated private key has invalid format: %v", err)
	}

	if err := ValidateKeyFormat(kp.PublicKey); err != nil {
		t.Errorf("Generated public key has invalid format: %v", err)
	}
}

func TestGeneratePresharedKey(t *testing.T) {
	psk, err := GeneratePresharedKey()
	if err != nil {
		t.Fatalf("Failed to generate preshared key: %v", err)
	}

	if err := ValidateKeyFormat(psk); err != nil {
		t.Errorf("Generated preshared key has invalid format: %v", err)
	}
}

func TestSaveAndLoadKey(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "test.key")
	
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	err = SaveKeyToFile(kp.PrivateKey, keyPath)
	if err != nil {
		t.Fatalf("Failed to save key: %v", err)
	}

	loadedKey, err := LoadKeyFromFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to load key: %v", err)
	}

	if loadedKey != kp.PrivateKey {
		t.Errorf("Loaded key does not match saved key. Expected %s, got %s", kp.PrivateKey, loadedKey)
	}

	// Verify permissions
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %v", info.Mode().Perm())
	}
}
