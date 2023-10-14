package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type InternalAdminUserRepo struct{}

func (repo *InternalAdminUserRepo) GetOne(ctx context.Context, db database.QueryExecer) (*domain.InternalAdminUser, error) {
	var dto = &dto.InternalAdminUserPgDTO{}
	fields := strings.Join(database.GetFieldNames(dto), ",")
	_, values := dto.FieldMap()
	err := db.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s
		FROM internal_admin_user iau
		WHERE iau.deleted_at IS NULL
		LIMIT 1;
	`, fields)).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: [%v]", err)
	}
	return dto.ToInternalAdminUserDomain(), nil
}
