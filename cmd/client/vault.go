package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"github.com/patric-chuzhbe/gophkeeper/internal/config"
	"io"
	"net/http"
)

type VaultRecord struct {
	Type  string            `json:"type"`
	Meta  map[string]string `json:"meta"`
	Data  []byte            `json:"data"`
	Nonce []byte            `json:"nonce"`
}

func AddRecord(password string, record VaultRecord) error {
	salt := []byte("static-salt") // Для демонстрации — заменить на динамический!
	key := DeriveKey([]byte(password), salt)

	nonce, encrypted, err := Encrypt(record.Data, key)
	if err != nil {
		return err
	}
	record.Data = encrypted
	record.Nonce = nonce

	body, _ := json.Marshal(record)
	resp, err := http.Post(app.config.ServerURL+"/vault", "application/json", bytes.NewReader(body))
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
