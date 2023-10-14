package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// StudentEnrollmentStatusHistoryRepo stores
type StudentEnrollmentStatusHistoryRepo struct{}

func (r *StudentEnrollmentStatusHistoryRepo) Upsert(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToUpsert *entity.StudentEnrollmentStatusHistory) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentEnrollmentStatusHistoryRepo.Upsert")
	defer span.End()

	now := time.Now()

	if err := multierr.Combine(
		enrollmentStatusHistoryToUpsert.CreatedAt.Set(now),
		enrollmentStatusHistoryToUpsert.UpdatedAt.Set(now),
		enrollmentStatusHistoryToUpsert.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("err set StudentEnrollmentStatusHistory: %w", err)
	}

	fields, values := enrollmentStatusHistoryToUpsert.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	stmt := "INSERT INTO " + enrollmentStatusHistoryToUpsert.TableName() + " (" + strings.Join(fields, ",") + ") VALUES (" + placeHolders + ") ON CONFLICT ON CONSTRAINT pk__student_enrollment_status_history DO UPDATE SET deleted_at = NULL;"

	cmd, err := db.Exec(ctx, stmt, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("cannot upsert StudentEnrollmentStatusHistory")
	}

	return nil
}

func (r *StudentEnrollmentStatusHistoryRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentEnrollmentStatusHistoryRepo.SoftDelete")
	defer span.End()

	sql := `UPDATE student_enrollment_status_history SET deleted_at = NOW(), updated_at = NOW() WHERE student_id = ANY($1)`
	_, err := db.Exec(ctx, sql, &studentIDs)
	if err != nil {
		return err
	}

	return nil
}
