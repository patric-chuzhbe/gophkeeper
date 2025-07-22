package clientvault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/patric-chuzhbe/gophkeeper/internal/config"

	//"github.com/patric-chuzhbe/gophkeeper/cmd/client"

	//"github.com/patric-chuzhbe/gophkeeper/internal/config"
	"github.com/patric-chuzhbe/gophkeeper/internal/clientcrypto"
	"io"
	"net/http"
)

type Vault struct {
	config *config.Config
}

type VaultRecord struct {
	Type  string            `json:"type"`
	Meta  map[string]string `json:"meta"`
	Data  []byte            `json:"data"`
	Nonce []byte            `json:"nonce"`
}

func New(config *config.Config) *Vault {
	return &Vault{config: config}
}

func (v *Vault) AddRecord(password string, record VaultRecord) error {
	salt := []byte("static-salt") // Для демонстрации — заменить на динамический!
	key := clientcrypto.DeriveKey([]byte(password), salt)

	nonce, encrypted, err := clientcrypto.Encrypt(record.Data, key)
	if err != nil {
		return err
	}
	record.Data = encrypted
	record.Nonce = nonce

	body, _ := json.Marshal(record)
	resp, err := http.Post(v.config.ServerURL+"/vault", "application/json", bytes.NewReader(body))
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
