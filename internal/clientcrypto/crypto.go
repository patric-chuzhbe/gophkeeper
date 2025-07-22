package clientcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"golang.org/x/crypto/argon2"
	"io"
)

func DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}

func Encrypt(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}
