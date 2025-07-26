package clientvault

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const sessionDir = ".vault-cli"    // TODO: move to config
const sessionFile = "session.json" // TODO: move to config

type Session struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func getSessionPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(homeDir, sessionDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, sessionFile), nil
}

func SaveSession(token string, password string) error {
	path, err := getSessionPath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(Session{Token: token, Password: password})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func LoadSession() (string, string, error) {
	path, err := getSessionPath()
	if err != nil {
		return "", "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return "", "", err
	}

	if sess.Token == "" {
		return "", "", errors.New("token is empty")
	}

	return sess.Token, sess.Password, nil
}
