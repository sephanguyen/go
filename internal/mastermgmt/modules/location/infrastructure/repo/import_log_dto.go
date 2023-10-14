package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

const (
	ImportTypeLocation     = "location"
	ImportTypeLocationType = "location_type"
	ImportTypeUnknown      = "unknown"
)

type ImportLog struct {
	ID         pgtype.Text `sql:"mastermgmt_import_log_id,pk"`
	UserID     pgtype.Text `sql:"user_id"`
	ImportType pgtype.Text `sql:"import_type"`
	Payload    pgtype.JSONB
	CreatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (i *ImportLog) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"mastermgmt_import_log_id", "user_id", "import_type", "payload", "created_at", "deleted_at"}
	values = []interface{}{&i.ID, &i.UserID, &i.ImportType, &i.Payload, &i.CreatedAt, &i.DeletedAt}
	return
}

func (i *ImportLog) TableName() string {
	return "mastermgmt_import_log"
}

func ToImportLog(e *domain.ImportLog) (*ImportLog, error) {
	importLogDto := &ImportLog{}
	database.AllNullEntity(importLogDto)
	err := multierr.Combine(
		importLogDto.ID.Set(e.ID),
		importLogDto.UserID.Set(e.UserID),
		importLogDto.ImportType.Set(e.ImportType),
		importLogDto.Payload.Set(e.Payload),
		importLogDto.CreatedAt.Set(e.CreatedAt),
	)
	if e.DeletedAt != nil {
		if err := importLogDto.DeletedAt.Set(e.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("could not mapping from importLog entity to importLog dto: %w", err)
	}
	return importLogDto, err
}
