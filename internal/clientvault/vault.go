package clientvault

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	pb "github.com/patric-chuzhbe/gophkeeper/internal/servergrpc/proto"
)

type Vault struct {
	crypto   cryptographer
	client   pb.VaultServiceClient
	token    string
	password string
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

/*< new funcs>*/

func (v *Vault) Register(login, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := v.client.Register(ctx, &pb.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	fmt.Println("Registration successful.")

	return nil
}

func (v *Vault) Login(login string, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := v.client.Login(ctx, &pb.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	v.token = resp.Token

	if err := SaveSession(resp.Token, password); err != nil {
		fmt.Println("Warning: token not saved to disk:", err)
	} else {
		fmt.Println("Token saved")
	}

	fmt.Println("Login successful.")
	return nil
}

func (v *Vault) SetToken(token string) {
	v.token = token
}

func New(crypto cryptographer, grpcConn *grpc.ClientConn) *Vault {
	v := &Vault{
		crypto: crypto,
		client: pb.NewVaultServiceClient(grpcConn),
	}

	if token, password, err := LoadSession(); err == nil {
		v.token = token
		v.password = password
	}

	return v
}

func (v *Vault) AddRecord(record VaultRecord) error {
	salt := []byte("static-salt") // TODO: получать от сервера
	key := v.crypto.DeriveKey([]byte(v.password), salt)

	nonce, encrypted, err := v.crypto.Encrypt(record.Data, key)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = v.client.AddRecord(ctx, &pb.AddRecordRequest{
		Token: v.token,
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

func (v *Vault) FetchRecords() ([]VaultRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := v.client.GetRecords(ctx, &pb.GetRecordsRequest{Token: v.token})
	if err != nil {
		return nil, fmt.Errorf("gRPC GetRecords failed: %w", err)
	}

	salt := []byte("static-salt") // TODO: получать от сервера
	key := v.crypto.DeriveKey([]byte(v.password), salt)

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

/*</new funcs>*/
