package security

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/crypto/curve25519"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

var keyFormatRegex = regexp.MustCompile(`^[A-Za-z0-9+/]{43}=$`)

// IsWgAvailable checks if the wg command is available in PATH.
func IsWgAvailable() bool {
	_, err := exec.LookPath("wg")
	return err == nil
}

// GenerateKeyPair generates a new private and public key pair.
func GenerateKeyPair() (*KeyPair, error) {
	if IsWgAvailable() {
		// Use wg command
		privOut, err := exec.Command("wg", "genkey").Output()
		if err != nil {
			return nil, fmt.Errorf("wg genkey failed: %w", err)
		}
		privKey := strings.TrimSpace(string(privOut))

		cmd := exec.Command("wg", "pubkey")
		cmd.Stdin = strings.NewReader(privKey)
		pubOut, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("wg pubkey failed: %w", err)
		}
		pubKey := strings.TrimSpace(string(pubOut))

		return &KeyPair{
			PrivateKey: privKey,
			PublicKey:  pubKey,
		}, nil
	}

	// Fallback to crypto/rand and curve25519
	var privateKey [32]byte
	if _, err := rand.Read(privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// WireGuard standard key tweaking (curve25519)
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey[:]),
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey[:]),
	}, nil
}

// GeneratePresharedKey generates a new preshared key.
func GeneratePresharedKey() (string, error) {
	if IsWgAvailable() {
		out, err := exec.Command("wg", "genpsk").Output()
		if err != nil {
			return "", fmt.Errorf("wg genpsk failed: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	var psk [32]byte
	if _, err := rand.Read(psk[:]); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.StdEncoding.EncodeToString(psk[:]), nil
}

// SaveKeyToFile writes a key to a file with 0600 permissions.
func SaveKeyToFile(key, path string) error {
	if err := ValidateKeyFormat(key); err != nil {
		return fmt.Errorf("invalid key format: %w", err)
	}
	return os.WriteFile(path, []byte(key+"\n"), 0600)
}

// LoadKeyFromFile reads a key from a file and validates its format.
func LoadKeyFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	key := string(bytes.TrimSpace(data))
	if err := ValidateKeyFormat(key); err != nil {
		return "", fmt.Errorf("invalid key format in file: %w", err)
	}

	return key, nil
}

// ValidateKeyFormat checks if a key is a valid base64 encoded 44-character string.
func ValidateKeyFormat(key string) error {
	if len(key) != 44 {
		return fmt.Errorf("invalid key length: expected 44, got %d", len(key))
	}
	if !keyFormatRegex.MatchString(key) {
		return errors.New("invalid key format: must be base64 encoded")
	}
	return nil
}
