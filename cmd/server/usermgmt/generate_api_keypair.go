package usermgmt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

var (
	userID string
)

func init() {
	bootstrap.RegisterJob("usermgmt_generate_api_keypair", runGenerateAPIKeypair).
		Desc("Cmd to generate api key").
		StringVar(&organizationID, "organizationID", "", "organization id").
		StringVar(&userID, "userID", "", "user id")
}

type GenerateAPIKeypairRequest struct {
	userID field.String
}

func (r *GenerateAPIKeypairRequest) UserID() field.String {
	return r.userID
}

func runGenerateAPIKeypair(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("bob")
	zLogger := rsc.Logger()
	return RunGenerateAPIKeypair(ctx, &c, db.DB.(*pgxpool.Pool), zLogger, userID, organizationID)
}

func RunGenerateAPIKeypair(ctx context.Context, c *configurations.Config, dbPool *pgxpool.Pool, zLogger *zap.Logger, userID, organizationID string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger.Sugar().Info("-----START: Generating API Keypair-----")
	defer func() {
		_ = zLogger.Sugar().Sync()
	}()

	ctx = auth.InjectFakeJwtToken(ctx, organizationID)
	domainAPIKeypairRepo := &repository.DomainAPIKeypairRepo{
		EncryptedKey:  c.OpenAPI.AESKey,
		InitialVector: c.OpenAPI.AESIV,
	}
	userGroupRepo := &repository.DomainUserGroupRepo{}
	userGroupMemberRepo := &repository.DomainUserGroupMemberRepo{}

	service := service.APIKeyPairService{
		DB:                   dbPool,
		DomainAPIKeypairRepo: domainAPIKeypairRepo,
		UserGroupRepo:        userGroupRepo,
		UserGroupMemberRepo:  userGroupMemberRepo,
	}

	req := GenerateAPIKeypairRequest{
		userID: field.NewString(userID),
	}

	err := service.GenerateKey(ctx, &req)
	if err != nil {
		zLogger.Sugar().Fatalf("service.GenerateKey err: %v", err)
		return fmt.Errorf("service.GenerateKey err: %v", err)
	}

	zLogger.Sugar().Info("-----DONE: Generating API Keypair. Please check the database-----")
	return nil
}
