package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"go.uber.org/multierr"
)

type MasterMgmtClassStudentRepo struct {
}

const masterMgmtClassStudentRepoUpsertStmt = `INSERT INTO class_students AS cs (%s) VALUES (%s)
ON CONFLICT (student_id, class_id)
DO UPDATE SET
updated_at = NOW(), deleted_at = NULL
`

func (r *MasterMgmtClassStudentRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.ClassStudent) error {
	now := timeutil.Now()
	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	fieldNames, values := e.FieldMap()
	placeHolders := "$1, $2, $3, $4, $5"

	query := fmt.Sprintf(masterMgmtClassStudentRepoUpsertStmt, strings.Join(fieldNames, ","), placeHolders)

	ct, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return errors.New("cannot upsert class student")
	}
	return nil
}

func (r *MasterMgmtClassStudentRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, classIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.SoftDelete")
	defer span.End()

	entity := &entities.ClassStudent{}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE student_id = ANY($1) AND class_id = ANY($2) AND deleted_at IS NULL`, entity.TableName())

	_, err := db.Exec(ctx, query, studentIDs, classIDs)
	if err != nil {
		return err
	}
	return nil
}
