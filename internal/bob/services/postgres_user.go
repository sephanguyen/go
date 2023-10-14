package services

import (
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type PostgresUserService struct {
	pb.UnimplementedPostgresUserServiceServer
	DB               database.Ext
	OriginalKey      string
	PrivateKey       *rsa.PrivateKey
	PostgresUserRepo interface {
		Get(ctx context.Context, db database.QueryExecer) ([]*entities.PostgresUser, error)
	}
}

func (p *PostgresUserService) GetPostgresUserPermission(ctx context.Context, req *pb.GetPostgresUserPermissionRequest) (*pb.GetPostgresUserPermissionResponse, error) {
	md, valid := metadata.FromIncomingContext(ctx)
	if !valid {
		return nil, status.Error(codes.Unknown, "can't get MD info from incoming context")
	}

	key := md["key"]
	if len(key) == 0 {
		return nil, status.Error(codes.PermissionDenied, "key is invalid")
	}

	bytes, err := base64.StdEncoding.DecodeString(key[0])
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to decode key from base64: %v", err)
	}

	decryptedBytes, err := p.PrivateKey.Decrypt(nil, bytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to decrypt data: %v", err)
	}

	if string(decryptedBytes) != p.OriginalKey {
		return nil, status.Error(codes.PermissionDenied, "key after decrypted invalid")
	}

	if len(key) == 0 {
		return nil, status.Error(codes.PermissionDenied, "key is invalid")
	}

	postgresUsers, err := p.PostgresUserRepo.Get(ctx, p.DB)
	if err != nil {
		return nil, fmt.Errorf("error when call PostgresUserRepo.Get: %v", err)
	}
	var response = new(pb.GetPostgresUserPermissionResponse)
	for i := range postgresUsers {
		postgresUser := &pb.PostgresUser{
			UserName:     postgresUsers[i].UserName.String,
			UseRepl:      postgresUsers[i].UseRepl.Bool,
			UseSuper:     postgresUsers[i].UseSuper.Bool,
			UseCreateDb:  postgresUsers[i].UseCreateDB.Bool,
			UseByPassRls: postgresUsers[i].UseByPassRLS.Bool,
		}
		response.PostgresUsers = append(response.PostgresUsers, postgresUser)
	}
	return response, nil
}

type PostgresNamespaceService struct {
	DB                    database.Ext
	OriginalKey           string
	PrivateKey            *rsa.PrivateKey
	PostgresNamespaceRepo interface {
		Get(ctx context.Context, db database.QueryExecer) ([]*entities.PostgresNamespace, error)
	}
}

func (p *PostgresNamespaceService) GetPostgresNamespace(ctx context.Context, req *pb.GetPostgresNamespaceRequest) (*pb.GetPostgresNamespaceResponse, error) {
	md, valid := metadata.FromIncomingContext(ctx)
	if !valid {
		return nil, status.Error(codes.Unknown, "can't get MD info from incoming context")
	}

	key := md["key"]
	fmt.Println("key: ", key[0])
	if len(key) == 0 {
		return nil, status.Error(codes.PermissionDenied, "key is invalid")
	}

	bytes, err := base64.StdEncoding.DecodeString(key[0])
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to decode key from base64: %v", err)
	}

	decryptedBytes, err := p.PrivateKey.Decrypt(nil, bytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to decrypt data: %v", err)
	}

	if string(decryptedBytes) != p.OriginalKey {
		return nil, status.Error(codes.PermissionDenied, "key after decrypted invalid")
	}

	postgresNamespaces, err := p.PostgresNamespaceRepo.Get(ctx, p.DB)
	if err != nil {
		return nil, fmt.Errorf("error when call PostgresNamespaceRepo.Get: %v", err)
	}

	var response = new(pb.GetPostgresNamespaceResponse)
	for i := range postgresNamespaces {
		postgresNamespace := &pb.PostgresNamespace{
			Namespace:        postgresNamespaces[i].Namespace.String,
			AccessPrivileges: database.FromTextArray(postgresNamespaces[i].AccessPrivileges),
		}
		response.PostgresNamespaces = append(response.PostgresNamespaces, postgresNamespace)
	}
	return response, nil
}
