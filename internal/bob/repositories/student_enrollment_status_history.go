package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type StudentEnrolledHistoryRepo struct{}

func (u *StudentEnrolledHistoryRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, locationID pgtype.Text) ([]*entities.StudentEnrollmentStatusHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEnrolledHistoryRepo.Retrieve")
	defer span.End()
	s := &entities.StudentEnrollmentStatusHistory{}
	fields := database.GetFieldNames(s)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = ANY ($1) AND location_id = $2 AND enrollment_status = 'STUDENT_ENROLLMENT_STATUS_ENROLLED' AND deleted_at is null", strings.Join(fields, ","), s.TableName())
	rows, err := db.Query(ctx, query, &ids, &locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*entities.StudentEnrollmentStatusHistory
	for rows.Next() {
		p := new(entities.StudentEnrollmentStatusHistory)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return res, nil
}
