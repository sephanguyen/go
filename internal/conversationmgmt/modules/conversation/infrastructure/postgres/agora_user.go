package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type AgoraUserRepo struct{}

func (repo *AgoraUserRepo) GetByUserIDs(ctx context.Context, db database.Ext, userIDs []string) ([]*domain.ChatVendorUser, error) {
	fields := strings.Join(database.GetFieldNames(&dto.AgoraUserPgDTO{}), ",")
	dtoArr := dto.AgoraUserPgDTOs{}
	err := database.Select(ctx, db, fmt.Sprintf(`
			SELECT %s
			FROM agora_user au
			WHERE au.user_id = ANY($1)
			AND au.deleted_at IS NULL;
		`, fields), userIDs).ScanAll(&dtoArr)
	if err != nil {
		return nil, fmt.Errorf("database.Select %w", err)
	}

	return dtoArr.ToChatVendorUsersDomain(), nil
}

func (repo *AgoraUserRepo) GetByUserID(ctx context.Context, db database.Ext, userID string) (*domain.ChatVendorUser, error) {
	dto := &dto.AgoraUserPgDTO{}
	fields := strings.Join(database.GetFieldNames(dto), ",")
	_, values := dto.FieldMap()

	err := db.QueryRow(ctx, fmt.Sprintf(`
			SELECT %s
			FROM agora_user au
			WHERE au.user_id = $1
				AND au.deleted_at IS NULL
			LIMIT 1;
		`, fields), database.Text(userID)).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow %w", err)
	}

	return dto.ToChatVendorUserDomain(), nil
}

func (repo *AgoraUserRepo) GetByVendorUserIDs(ctx context.Context, db database.Ext, vendorUserIDs []string) ([]*domain.ChatVendorUser, error) {
	fields := strings.Join(database.GetFieldNames(&dto.AgoraUserPgDTO{}), ",")
	dtoArr := dto.AgoraUserPgDTOs{}
	err := database.Select(ctx, db, fmt.Sprintf(`
			SELECT %s
			FROM agora_user au
			WHERE au.agora_user_id = ANY($1)
				AND au.deleted_at IS NULL;
		`, fields), vendorUserIDs).ScanAll(&dtoArr)
	if err != nil {
		return nil, fmt.Errorf("database.Select %w", err)
	}

	return dtoArr.ToChatVendorUsersDomain(), nil
}

func (repo *AgoraUserRepo) GetByVendorUserID(ctx context.Context, db database.Ext, vendorUserID string) (*domain.ChatVendorUser, error) {
	dto := &dto.AgoraUserPgDTO{}
	fields := strings.Join(database.GetFieldNames(dto), ",")
	_, values := dto.FieldMap()

	err := db.QueryRow(ctx, fmt.Sprintf(`
			SELECT %s
			FROM agora_user au
			WHERE au.agora_user_id = $1
				AND au.deleted_at IS NULL
			LIMIT 1;
		`, fields), database.Text(vendorUserID)).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow %w", err)
	}

	return dto.ToChatVendorUserDomain(), nil
}
