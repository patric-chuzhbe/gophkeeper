package main

import (
	"log"

	"github.com/patric-chuzhbe/gophkeeper/internal/clientcli"
	"github.com/patric-chuzhbe/gophkeeper/internal/clientcrypto"
	"github.com/patric-chuzhbe/gophkeeper/internal/clientvault"
	"github.com/patric-chuzhbe/gophkeeper/internal/config"
)

type App struct {
	config *config.Config
}

var app *App

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	app = &App{
		config: cfg,
	}

	crypto := clientcrypto.New()

	cli := clientcli.New(
		clientvault.New(
			app.config.ServerURL,
			crypto,
		),
	)

	cli.Execute()
}
