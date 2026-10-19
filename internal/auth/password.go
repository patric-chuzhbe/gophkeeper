package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLen   = 16
	timeCost  = 1
	memory    = 64 * 1024
	threads   = 4
	keyLength = 32
	prefix    = "argon2id"
)

// HashPassword returns a base64-encoded hash using Argon2id
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, keyLength)
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s$%s$%s", prefix, saltB64, hashB64), nil
}

// VerifyPassword compares password with the hashed value
func VerifyPassword(password, encodedHash string) error {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 3 || parts[0] != prefix {
		return errors.New("invalid hash format")
	}

	saltB64 := parts[1]
	hashB64 := parts[2]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return err
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return err
	}

	actualHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, keyLength)

	if !equal(actualHash, expectedHash) {
		return errors.New("password does not match")
	}
	return nil
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte
	for i := range a {
		res |= a[i] ^ b[i]
	}
	return res == 0
}
