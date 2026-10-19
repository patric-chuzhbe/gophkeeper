package main

import (
	"log"

	"github.com/patric-chuzhbe/gophkeeper/internal/clientapp"
)

func main() {
	a, err := clientapp.New()
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	a.Run()
}
