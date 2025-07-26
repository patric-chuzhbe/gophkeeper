package clientvault

import (
	"context"
	"fmt"
	"time"

	"github.com/patric-chuzhbe/gophkeeper/internal/config"

	"google.golang.org/grpc"

	pb "github.com/patric-chuzhbe/gophkeeper/internal/servergrpc/proto"
)

type Vault struct {
	config *config.Config
	crypto cryptographer
	client pb.VaultServiceClient
}

type VaultRecord struct {
	Type  string            `json:"type"`
	Meta  map[string]string `json:"meta"`
	Data  []byte            `json:"data"`
	Nonce []byte            `json:"nonce"`
}

type cryptographer interface {
	DeriveKey(password, salt []byte) []byte
	Encrypt(plaintext []byte, key []byte) ([]byte, []byte, error)
	Decrypt(ciphertext []byte, nonce []byte, key []byte) ([]byte, error)
}

func New(crypto cryptographer, grpcConn *grpc.ClientConn) *Vault {
	return &Vault{
		crypto: crypto,
		client: pb.NewVaultServiceClient(grpcConn),
	}
}

func (v *Vault) AddRecord(login string, password string, record VaultRecord) error {
	salt := []byte("static-salt") // TODO: получать от сервера
	key := v.crypto.DeriveKey([]byte(password), salt)

	nonce, encrypted, err := v.crypto.Encrypt(record.Data, key)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = v.client.AddRecord(ctx, &pb.AddRecordRequest{
		Login:    login,
		Password: password,
		Record: &pb.VaultRecord{
			Type:  record.Type,
			Meta:  record.Meta,
			Data:  encrypted,
			Nonce: nonce,
		},
	})
	if err != nil {
		return fmt.Errorf("gRPC AddRecord failed: %w", err)
	}

	fmt.Println("Record stored successfully.")
	return nil
}

func (v *Vault) FetchRecords(login string, password string) ([]VaultRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := v.client.GetRecords(ctx, &pb.GetRecordsRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC GetRecords failed: %w", err)
	}

	salt := []byte("static-salt") // TODO: получать от сервера
	key := v.crypto.DeriveKey([]byte(password), salt)

	var records []VaultRecord
	for _, r := range res.Records {
		plain, err := v.crypto.Decrypt(r.Data, r.Nonce, key)
		if err != nil {
			continue // TODO: логировать
		}
		records = append(records, VaultRecord{
			Type:  r.Type,
			Meta:  r.Meta,
			Data:  plain,
			Nonce: r.Nonce,
		})
	}
	return records, nil
}
