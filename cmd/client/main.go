package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

	dialContext, cancelDial := context.WithTimeout(context.Background(), cfg.GRPCDialTimeout)
	defer cancelDial()

	conn, err := grpc.DialContext(
		dialContext,
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	crypto := clientcrypto.New()

	cli := clientcli.New(clientvault.New(crypto, conn))

	cli.Execute()
}
