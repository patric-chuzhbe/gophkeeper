package main

import (
	"encoding/json"
	"net/http"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var u User
	json.NewDecoder(r.Body).Decode(&u)
	// TODO: hash password and save
	w.WriteHeader(http.StatusCreated)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var u User
	json.NewDecoder(r.Body).Decode(&u)
	// TODO: verify password and return token
	w.Write([]byte(`{"token":"example-token"}`))
}

func handleVault(w http.ResponseWriter, r *http.Request) {
	// TODO: auth + encrypted data handling
	w.Write([]byte(`vault endpoint`))
}
