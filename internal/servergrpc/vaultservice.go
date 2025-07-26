package servergrpc

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/patric-chuzhbe/gophkeeper/internal/auth"
	pb "github.com/patric-chuzhbe/gophkeeper/internal/servergrpc/proto"
	"github.com/patric-chuzhbe/gophkeeper/internal/serverstorage/postgresdb"
)

type VaultServer struct {
	pb.UnimplementedVaultServiceServer
	db *postgresdb.PostgresDB
}

/*< new funcs>*/
func (s *VaultServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	userID, err := s.db.Authenticate(ctx, req.Login, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "auth fault")
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{Token: token}, nil
}

func NewVaultServer(db *postgresdb.PostgresDB) *VaultServer {
	return &VaultServer{db: db}
}

func (s *VaultServer) AddRecord(ctx context.Context, req *pb.AddRecordRequest) (*pb.AddRecordResponse, error) {
	userID, err := auth.ParseToken(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	tx, err := s.db.BeginTransaction()
	if err != nil {
		return nil, err
	}
	defer s.db.RollbackTransaction(tx)

	err = s.db.InsertRecord(ctx, userID, req.Record.Data, req.Record.Nonce, req.Record.Type, tx)
	if err != nil {
		log.Printf("error inserting record: %v", err)
		return nil, status.Error(codes.Internal, "failed to add record")
	}

	err = s.db.CommitTransaction(tx)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB fault")
	}

	return &pb.AddRecordResponse{Success: true, Message: "Record added"}, nil
}

func (s *VaultServer) GetRecords(ctx context.Context, req *pb.GetRecordsRequest) (*pb.GetRecordsResponse, error) {
	userID, err := auth.ParseToken(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	records, err := s.db.GetUserRecords(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve records")
	}

	result := &pb.GetRecordsResponse{Records: []*pb.VaultRecord{}}
	for _, record := range records {
		result.Records = append(result.Records, &pb.VaultRecord{
			Type:  record.Type,
			Data:  record.Data,
			Nonce: record.Nonce,
		})
	}

	return result, nil
}

/*</new funcs>*/
