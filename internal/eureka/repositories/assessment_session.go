package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type AssessmentSessionRepo struct{}

func (r *AssessmentSessionRepo) GetAssessmentSessionByAssessmentIDs(ctx context.Context, db database.QueryExecer, assessmentIDs pgtype.TextArray) ([]*entities.AssessmentSession, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentSessionRepo.GetAssessmentSessionByAssessmentIDs")
	defer span.End()

	var assessmentSession entities.AssessmentSession
	var assessmentSessions entities.AssessmentSessions

	fields, _ := assessmentSession.FieldMap()

	stmt := fmt.Sprintf(`
    SELECT %s
    FROM %s
    WHERE assessment_id = ANY($1::_TEXT)
	AND deleted_at IS NULL
	AND status = 'COMPLETED'
	ORDER BY created_at DESC`, strings.Join(fields, ", "), assessmentSession.TableName())

	err := database.Select(ctx, db, stmt, &assessmentIDs).ScanAll(&assessmentSessions)

	if err != nil {
		return nil, err
	}

	return assessmentSessions, nil
}

func (r *AssessmentSessionRepo) CountByAssessment(ctx context.Context, db database.QueryExecer, assessmentIDs pgtype.TextArray) (int32, error) {
	assessmentSession := &entities.AssessmentSession{}
	var total int32

	stmt := `SELECT count(*) FROM %s WHERE assessment_id = ANY($1::_TEXT) AND deleted_at IS NULL`
	stmt = fmt.Sprintf(stmt, assessmentSession.TableName())

	if err := db.QueryRow(ctx, stmt, assessmentIDs).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}
