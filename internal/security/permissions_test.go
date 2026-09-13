package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecureDirectory(t *testing.T) {
	dir := t.TempDir()
	securePath := filepath.Join(dir, "secure_dir")

	err := SecureDirectory(securePath)
	if err != nil {
		t.Fatalf("Failed to create secure directory: %v", err)
	}

	info, err := os.Stat(securePath)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	if info.Mode().Perm() != 0700 {
		t.Errorf("Expected permissions 0700, got %v", info.Mode().Perm())
	}
}

func TestSecureFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "secure_file.txt")

	// Create with open permissions
	err := os.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	err = SecureFile(filePath)
	if err != nil {
		t.Fatalf("Failed to secure file: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %v", info.Mode().Perm())
	}
}

func TestValidateFilePermissions(t *testing.T) {
	dir := t.TempDir()
	
	// Create secure file
	secureFile := filepath.Join(dir, "secure.txt")
	err := os.WriteFile(secureFile, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Failed to create secure file: %v", err)
	}

	err = ValidateFilePermissions(secureFile)
	if err != nil {
		t.Errorf("Secure file failed validation: %v", err)
	}

	// Create insecure file
	insecureFile := filepath.Join(dir, "insecure.txt")
	err = os.WriteFile(insecureFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create insecure file: %v", err)
	}

	err = ValidateFilePermissions(insecureFile)
	if err == nil {
		t.Error("Insecure file passed validation")
	}
}

func TestSecureTempFile(t *testing.T) {
	dir := t.TempDir()
	
	f, err := SecureTempFile(dir, "test-*")
	if err != nil {
		t.Fatalf("Failed to create secure temp file: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	info, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("Failed to stat temp file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %v", info.Mode().Perm())
	}
}
