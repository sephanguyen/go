package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/pkg/errors"
)

func (s *suite) systemRunJobToGenerateAPIKeyWithOrganization(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	zLogger := logger.NewZapLogger("warn", s.Cfg.Common.Environment == "local")
	err := usermgmt.RunGenerateAPIKeypair(ctx, &configurations.Config{
		OpenAPI: configurations.OpenAPI{
			AESIV:  AESIV,
			AESKey: AESKey,
		},
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, s.BobPostgresDB, zLogger, stepState.CurrentUserID, stepState.OrganizationID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "systemRunJobToGenerateAPIKeyWithOrganization failed")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) apiKeyIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count int
	query := `SELECT count(*) FROM api_keypair WHERE user_id = $1 AND resource_path = $2 AND LENGTH(private_key) > 0`
	err := database.Select(ctx, s.BobPostgresDBTrace, query, &stepState.CurrentUserID, &stepState.OrganizationID).ScanFields(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("apiKeyIsCreatedSuccessfully failed: expected 1, actual %v", count)
	}

	return StepStateToContext(ctx, stepState), nil
}
