package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgx/v4"
)

type AssessmentSessionRepo struct{}

var _ repository.AssessmentSessionRepo = (*AssessmentSessionRepo)(nil)

func (a *AssessmentSessionRepo) GetLatestByIdentity(ctx context.Context, db database.Ext, assessmentID, userID string) (domain.Session, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentSessionRepo.GetLatestByIdentity")
	defer span.End()

	var result dto.AssessmentSession

	stmt := fmt.Sprintf(`
        SELECT %s
          FROM %s
         WHERE deleted_at IS NULL
           AND assessment_id = $1
           AND user_id = $2
         ORDER BY created_at DESC
         LIMIT 1;
	`, strings.Join(database.GetFieldNames(&result), ", "), result.TableName())

	if err := database.Select(ctx, db, stmt, database.Text(assessmentID), database.Text(userID)).ScanOne(&result); err != nil {
		if errors.IsPgxNoRows(err) {
			return domain.Session{}, errors.NewNoRowsExistedError("database.Select", err)
		}
		return domain.Session{}, errors.NewDBError("database.Select", err)
	}

	session, err := result.ToEntity()
	if err != nil {
		return domain.Session{}, errors.NewDBError("result.ToEntity", err)
	}

	return session, nil
}

func (a *AssessmentSessionRepo) GetManyByAssessments(ctx context.Context, db database.Ext, assessmentID, userID string) ([]domain.Session, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentSessionRepo.GetManyByAssessments")
	defer span.End()

	var holder dto.AssessmentSession

	query := fmt.Sprintf(`
        SELECT %s
          FROM %s
         WHERE deleted_at IS NULL
           AND assessment_id = $1
           AND user_id = $2
         ORDER BY created_at DESC;
	`, strings.Join(database.GetFieldNames(&holder), ", "), holder.TableName())

	rows, err := db.Query(ctx, query, database.Text(assessmentID), database.Text(userID))
	if err != nil {
		if errors.IsPgxNoRows(err) {
			return []domain.Session{}, nil
		}
		return nil, errors.NewDBError("AssessmentSessionRepo.GetManyByAssessments", err)
	}

	return scanAssessmentSessions(rows)
}

func (a *AssessmentSessionRepo) Insert(ctx context.Context, db database.Ext, now time.Time, session domain.Session) error {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentSessionRepo.Insert")
	defer span.End()

	assessmentSessionDto := dto.AssessmentSession{}
	if err := assessmentSessionDto.FromEntity(now, session); err != nil {
		return errors.NewDBError("assessmentSessionDto.FromEntity", err)
	}

	if _, err := database.Insert(ctx, &assessmentSessionDto, db.Exec); err != nil {
		return errors.NewDBError("database.Insert", err)
	}

	return nil
}

func (a *AssessmentSessionRepo) UpdateStatus(ctx context.Context, db database.Ext, now time.Time, session domain.Session) error {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentSessionRepo.UpdateStatus")
	defer span.End()

	assessmentSessionDto := dto.AssessmentSession{}
	if err := assessmentSessionDto.FromEntity(now, session); err != nil {
		return errors.NewDBError("assessmentSessionDto.FromEntity", err)
	}

	if _, err := database.UpdateFields(ctx, &assessmentSessionDto, db.Exec, "session_id", []string{
		"status",
		"updated_at",
	}); err != nil {
		return errors.NewDBError("database.UpdateFields", err)
	}

	return nil
}

func (a *AssessmentSessionRepo) GetByID(ctx context.Context, db database.Ext, id string) (domain.Session, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentSessionRepo.GetByID")
	defer span.End()

	var result dto.AssessmentSession

	stmt := fmt.Sprintf(`
        SELECT %s
          FROM %s
         WHERE session_id = $1;
	`, strings.Join(database.GetFieldNames(&result), ", "), result.TableName())

	if err := database.Select(ctx, db, stmt, database.Text(id)).ScanOne(&result); err != nil {
		if errors.IsPgxNoRows(err) {
			return domain.Session{}, errors.NewNoRowsExistedError("database.Select", err)
		}
		return domain.Session{}, errors.NewDBError("database.Select", err)
	}

	session, err := result.ToEntity()
	if err != nil {
		return domain.Session{}, errors.NewDBError("result.ToEntity", err)
	}

	return session, nil
}

func scanAssessmentSessions(rows pgx.Rows) ([]domain.Session, error) {
	var sessions []domain.Session
	asm := &dto.AssessmentSession{}
	fields, _ := asm.FieldMap()

	defer rows.Close()
	for rows.Next() {
		ds := new(dto.AssessmentSession)
		if err := rows.Scan(database.GetScanFields(ds, fields)...); err != nil {
			return nil, errors.NewConversionError("AssessmentSessionRepo.scanAssessmentSessions", err)
		}
		s := domain.Session{
			ID:           ds.SessionID.String,
			AssessmentID: ds.AssessmentID.String,
			UserID:       ds.UserID.String,
			Status:       domain.SessionStatus(ds.Status.String),
			CreatedAt:    ds.CreatedAt.Time,
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewConversionError("AssessmentSessionRepo.scanAssessmentSessions", err)
	}
	return sessions, nil
}
