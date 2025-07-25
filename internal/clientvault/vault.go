package clientvault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/patric-chuzhbe/gophkeeper/internal/config"

	"net/http"
)

type Vault struct {
	config    *config.Config
	serverURL string
	crypto    cryptographer
}

type VaultRecord struct {
	Type  string            `json:"type"`
	Meta  map[string]string `json:"meta"`
	Data  []byte            `json:"data"`
	Nonce []byte            `json:"nonce"`
}

type cryptographer interface {
	DeriveKey(password, salt []byte) []byte
	Encrypt(plaintext []byte, key []byte) ([]byte, []byte, error)
	Decrypt(ciphertext []byte, nonce []byte, key []byte) ([]byte, error)
}

func New(serverURL string, crypto cryptographer) *Vault {
	return &Vault{serverURL: serverURL, crypto: crypto}
}

func (v *Vault) AddRecord(password string, record VaultRecord) error {
	salt := []byte("static-salt") // TODO: Для демонстрации — заменить на динамический!
	key := v.crypto.DeriveKey([]byte(password), salt)

	nonce, encrypted, err := v.crypto.Encrypt(record.Data, key)
	if err != nil {
		return err
	}
	record.Data = encrypted
	record.Nonce = nonce

	body, _ := json.Marshal(record)
	resp, err := http.Post(v.serverURL+"/vault", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(b))
	}

	fmt.Println("Record stored successfully.")
	return nil
}

func (v *Vault) FetchRecords(password string) ([]VaultRecord, error) {
	resp, err := http.Get(v.serverURL + "/vault")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var encryptedRecords []VaultRecord
	if err := json.NewDecoder(resp.Body).Decode(&encryptedRecords); err != nil {
		return nil, err
	}

	salt := []byte("static-salt") // В будущем — из БД
	key := v.crypto.DeriveKey([]byte(password), salt)

	var decrypted []VaultRecord
	for _, rec := range encryptedRecords {
		plaintext, err := v.crypto.Decrypt(rec.Data, rec.Nonce, key)
		if err != nil {
			continue // TODO: можно логировать
		}
		rec.Data = plaintext
		decrypted = append(decrypted, rec)
	}

	return decrypted, nil
}
