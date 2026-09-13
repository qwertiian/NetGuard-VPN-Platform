package security

import (
	"errors"
	"fmt"
	"os"
)

// SecureDirectory creates a directory with 0700 permissions.
func SecureDirectory(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	if info.Mode().Perm() != 0700 {
		if err := os.Chmod(path, 0700); err != nil {
			return fmt.Errorf("failed to set directory permissions: %w", err)
		}
	}

	return nil
}

// SecureFile sets file permissions to 0600.
func SecureFile(path string) error {
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	return nil
}

// ValidateFilePermissions checks if a file has 0600 permissions.
func ValidateFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	perm := info.Mode().Perm()
	if perm&0077 != 0 {
		return errors.New("file is accessible by group or others")
	}

	return nil
}

// SecureTempFile creates a secure temporary file with 0600 permissions.
func SecureTempFile(dir, pattern string) (*os.File, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if err := SecureFile(file.Name()); err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	return file, nil
}
