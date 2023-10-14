package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type FeedbackSessionRepo struct{}

func (a *FeedbackSessionRepo) GetOneBySubmissionID(ctx context.Context, db database.Ext, subID string) (f *domain.FeedbackSession, err error) {
	ctx, span := interceptors.StartSpan(ctx, "FeedbackSessionRepo.GetOneBySubmissionID")
	defer span.End()

	var result dto.FeedbackSession

	stmt := fmt.Sprintf(`
        SELECT %s
          FROM %s
         WHERE deleted_at IS NULL
           AND submission_id = $1
         LIMIT 1;
	`, strings.Join(database.GetFieldNames(&result), ", "), result.TableName())

	if err := database.Select(ctx, db, stmt, database.Text(subID)).ScanOne(&result); err != nil {
		if errors.IsPgxNoRows(err) {
			return f, errors.NewNoRowsExistedError("FeedbackSessionRepo.GetOneBySubmissionID", err)
		}
		return f, errors.NewDBError("FeedbackSessionRepo.GetOneBySubmissionID", err)
	}
	res := result.ToEntity()
	return &res, nil
}

func (a *FeedbackSessionRepo) GetManyBySubmissionIDs(ctx context.Context, db database.Ext, subIDs []string) (subs []domain.FeedbackSession, err error) {
	ctx, span := interceptors.StartSpan(ctx, "FeedbackSessionRepo.GetManyBySubmissionIDs")
	defer span.End()

	var entity dto.FeedbackSession
	fields, _ := entity.FieldMap()

	count := 0
	sessionIDPlaceholders := sliceutils.Map(subIDs, func(t string) string {
		count++
		return fmt.Sprintf("$%d", count)
	})
	placeholderStr := strings.Join(sessionIDPlaceholders, ",")
	values := sliceutils.Map(subIDs, func(s string) any {
		return s
	})

	query := fmt.Sprintf(`SELECT %s from %s
		 WHERE deleted_at is NULL
		 AND submission_id in (%s);
	`, strings.Join(fields, ", "), entity.TableName(), placeholderStr)

	rows, err := db.Query(ctx, query, values...)
	if err != nil {
		if errors.IsPgxNoRows(err) {
			return []domain.FeedbackSession{}, nil
		}
		return nil, errors.NewDBError("FeedbackSessionRepo.GetManyBySubmissionIDs", err)
	}

	return scanFeedbackSessions(rows)
}

func scanFeedbackSessions(rows pgx.Rows) ([]domain.FeedbackSession, error) {
	var fs []domain.FeedbackSession
	dtoHolder := &dto.FeedbackSession{}
	fields, _ := dtoHolder.FieldMap()

	defer rows.Close()
	for rows.Next() {
		e := new(dto.FeedbackSession)
		if err := rows.Scan(database.GetScanFields(e, fields)...); err != nil {
			return nil, errors.NewConversionError("FeedbackSessionRepo.scanFeedbackSessions", err)
		}
		fs = append(fs, e.ToEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewConversionError("FeedbackSessionRepo.scanFeedbackSessions", err)
	}

	return fs, nil
}

func (a *FeedbackSessionRepo) Insert(ctx context.Context, db database.Ext, feedback domain.FeedbackSession) error {
	ctx, span := interceptors.StartSpan(ctx, "FeedbackSession.Insert")
	defer span.End()
	_, err := uuid.Parse(feedback.ID)
	if err != nil {
		return errors.NewValidationError("Feedback session id must be UUID", nil)
	}
	feedbackDto := dto.FeedbackSession{
		ID:           database.Text(feedback.ID),
		SubmissionID: database.Text(feedback.SubmissionID),
		CreatedBy:    database.Text(feedback.CreatedBy),
		BaseEntity: dto.BaseEntity{
			CreatedAt: database.Timestamptz(feedback.CreatedAt),
			UpdatedAt: database.Timestamptz(feedback.CreatedAt),
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		},
	}

	if _, err := database.Insert(ctx, &feedbackDto, db.Exec); err != nil {
		return errors.NewDBError("FeedbackSessionRepo.Insert", err)
	}

	return nil
}
