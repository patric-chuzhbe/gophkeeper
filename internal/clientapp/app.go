package clientapp

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
	cli   *clientcli.CLI
	conn  *grpc.ClientConn
	vault *clientvault.Vault
	cfg   *config.Config
}

func New() (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	dialContext, cancelDial := context.WithTimeout(context.Background(), cfg.GRPCDialTimeout)
	defer cancelDial()
	conn, err := grpc.DialContext(
		dialContext,
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	crypto := clientcrypto.New()

	vault := clientvault.New(crypto, conn)

	cli := clientcli.New(vault)

	return &App{
		cli:   cli,
		conn:  conn,
		vault: vault,
		cfg:   cfg,
	}, nil
}

func (a *App) Run() {
	a.cli.Execute()
}

func (a *App) Close() {
	if a.conn == nil {
		return
	}
	err := a.conn.Close()
	if err != nil {
		log.Fatalf("failed to close gRPC connection: %v", err)
	}
}
