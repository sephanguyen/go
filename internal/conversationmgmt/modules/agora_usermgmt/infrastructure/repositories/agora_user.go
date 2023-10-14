package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	agorausermgmt "github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type AgoraUserRepo struct{}

func (repo *AgoraUserRepo) Create(ctx context.Context, db database.QueryExecer, agoraUser *models.AgoraUser) error {
	ctx, span := interceptors.StartSpan(ctx, "AgoraUserRepo.Create")
	defer span.End()

	if agoraUser.UserID.String == "" {
		return fmt.Errorf("missing manabie_user_id when creating agora user")
	}

	now := time.Now()
	err := multierr.Combine(
		agoraUser.CreatedAt.Set(now),
		agoraUser.UpdatedAt.Set(now),
		agoraUser.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if agoraUser.AgoraUserID.String == "" {
		_ = agoraUser.AgoraUserID.Set(agorausermgmt.GetAgoraUserID(agoraUser.UserID.String))
	}

	fields := database.GetFieldNames(agoraUser)
	values := database.GetScanFields(agoraUser, fields)
	placeHolders := database.GeneratePlaceholders(len(fields))
	tableName := agoraUser.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as au (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT agora_user_pk 
		DO NOTHING;
	`, tableName, strings.Join(fields, ", "), placeHolders)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not upsert questionnaire")
	}

	return nil
}

// Support store user that can't create on Agora system
func (repo *AgoraUserRepo) CreateAgoraUserFailure(ctx context.Context, db database.QueryExecer, agoraUserFailure *models.AgoraUserFailure) error {
	ctx, span := interceptors.StartSpan(ctx, "AgoraUserRepo.CreateAgoraUserFailure")
	defer span.End()

	if agoraUserFailure.UserID.String == "" {
		return fmt.Errorf("missing manabie_user_id when creating agora user")
	}

	now := time.Now()
	err := multierr.Combine(
		agoraUserFailure.CreatedAt.Set(now),
		agoraUserFailure.UpdatedAt.Set(now),
		agoraUserFailure.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if agoraUserFailure.AgoraUserID.String == "" {
		_ = agoraUserFailure.AgoraUserID.Set(agorausermgmt.GetAgoraUserID(agoraUserFailure.UserID.String))
	}

	fields := database.GetFieldNames(agoraUserFailure)
	values := database.GetScanFields(agoraUserFailure, fields)
	placeHolders := database.GeneratePlaceholders(len(fields))
	tableName := agoraUserFailure.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as au (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT agora_user_failure_pk 
		DO NOTHING;
	`, tableName, strings.Join(fields, ", "), placeHolders)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not upsert questionnaire")
	}

	return nil
}
