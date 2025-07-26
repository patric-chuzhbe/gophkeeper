package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/patric-chuzhbe/gophkeeper/internal/config"
	"github.com/patric-chuzhbe/gophkeeper/internal/servergrpc"
	pb "github.com/patric-chuzhbe/gophkeeper/internal/servergrpc/proto"
	"github.com/patric-chuzhbe/gophkeeper/internal/serverstorage/postgresdb"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgresdb.New(cfg.DatabaseDSN, cfg.MigrationsDir)
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterVaultServiceServer(grpcServer, servergrpc.NewVaultServer(db))

	log.Println("gRPC server listening on", cfg.GRPCAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
