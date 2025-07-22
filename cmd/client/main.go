package main

import (
	"github.com/patric-chuzhbe/gophkeeper/client/commands"
	"github.com/patric-chuzhbe/gophkeeper/internal/config"
	"log"
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
	commands.Execute()
}
