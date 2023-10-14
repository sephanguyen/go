package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"github.com/jackc/pgx/v4"
)

type SubmissionRepo struct{}

func (a *SubmissionRepo) GetOneBySessionID(ctx context.Context, db database.Ext, sessionID string) (sub *domain.Submission, err error) {
	ctx, span := interceptors.StartSpan(ctx, "SubmissionRepo.GetOneBySessionID")
	defer span.End()

	var result dto.Submission

	stmt := fmt.Sprintf(`
        SELECT %s
          FROM %s
         WHERE deleted_at IS NULL
           AND session_id = $1
         LIMIT 1;
	`, strings.Join(database.GetFieldNames(&result), ", "), result.TableName())

	if err := database.Select(ctx, db, stmt, database.Text(sessionID)).ScanOne(&result); err != nil {
		if errors.IsPgxNoRows(err) {
			return sub, errors.NewNoRowsExistedError("SubmissionRepo.GetOneBySessionID", err)
		}
		return sub, errors.NewDBError("SubmissionRepo.GetOneBySessionID", err)
	}
	res := result.ToEntity()
	return &res, nil
}

func (a *SubmissionRepo) GetOneBySubmissionID(ctx context.Context, db database.Ext, subID string) (sub *domain.Submission, err error) {
	ctx, span := interceptors.StartSpan(ctx, "SubmissionRepo.GetOneBySubmissionID")
	defer span.End()

	var result dto.Submission

	stmt := fmt.Sprintf(`
        SELECT %s
          FROM %s
         WHERE deleted_at IS NULL
           AND id = $1
         LIMIT 1;
	`, strings.Join(database.GetFieldNames(&result), ", "), result.TableName())

	if err := database.Select(ctx, db, stmt, database.Text(subID)).ScanOne(&result); err != nil {
		if errors.IsPgxNoRows(err) {
			return sub, errors.NewNoRowsExistedError("SubmissionRepo.GetOneBySubmissionID", err)
		}
		return sub, errors.NewDBError("SubmissionRepo.GetOneBySubmissionID", err)
	}
	res := result.ToEntity()
	return &res, nil
}

func (a *SubmissionRepo) GetManyBySessionIDs(ctx context.Context, db database.Ext, sessionIDs []string) (subs []domain.Submission, err error) {
	ctx, span := interceptors.StartSpan(ctx, "SubmissionRepo.GetManyBySessionIDs")
	defer span.End()

	var entity dto.Submission
	fields, _ := entity.FieldMap()

	count := 0
	sessionIDPlaceholders := sliceutils.Map(sessionIDs, func(t string) string {
		count++
		return fmt.Sprintf("$%d", count)
	})
	placeholderStr := strings.Join(sessionIDPlaceholders, ",")
	values := sliceutils.Map(sessionIDs, func(s string) any {
		return s
	})

	query := fmt.Sprintf(`SELECT %s from %s
		 WHERE deleted_at is NULL
		 AND session_id in (%s);
	`, strings.Join(fields, ", "), entity.TableName(), placeholderStr)

	rows, err := db.Query(ctx, query, values...)
	if err != nil {
		return nil, errors.NewDBError("SubmissionRepo.GetManyBySessionIDs", err)
	}

	return scanSubmissions(rows)
}

func (a *SubmissionRepo) GetManyByAssessments(ctx context.Context, db database.Ext, studentID, asmID string) (subs []domain.Submission, err error) {
	ctx, span := interceptors.StartSpan(ctx, "SubmissionRepo.GetManyByAssessments")
	defer span.End()

	var entity dto.Submission
	fields, _ := entity.FieldMap()

	query := fmt.Sprintf(`SELECT %s from %s
		 WHERE deleted_at is NULL
		 AND student_id = $1
		 AND assessment_id = $2;
	`, strings.Join(fields, ", "), entity.TableName())

	rows, err := db.Query(ctx, query, database.Text(studentID), database.Text(asmID))
	if err != nil {
		return nil, errors.NewDBError("SubmissionRepo.GetManyByAssessments", err)
	}

	return scanSubmissions(rows)
}

func scanSubmissions(rows pgx.Rows) ([]domain.Submission, error) {
	var subs []domain.Submission
	dtoHolder := &dto.Submission{}
	fields, _ := dtoHolder.FieldMap()

	defer rows.Close()
	for rows.Next() {
		e := new(dto.Submission)
		if err := rows.Scan(database.GetScanFields(e, fields)...); err != nil {
			return nil, errors.NewConversionError("SubmissionRepo.scanSubmissions", err)
		}
		subs = append(subs, e.ToEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewConversionError("SubmissionRepo.scanSubmission", err)
	}

	return subs, nil
}

func (a *SubmissionRepo) Insert(ctx context.Context, db database.Ext, now time.Time, submission domain.Submission) error {
	ctx, span := interceptors.StartSpan(ctx, "SubmissionRepo.Insert")
	defer span.End()

	submissionDto := dto.Submission{}
	if err := submissionDto.FromEntity(now, submission); err != nil {
		return errors.NewDBError("submissionDto.FromEntity", err)
	}

	if _, err := database.Insert(ctx, &submissionDto, db.Exec); err != nil {
		return errors.NewDBError("database.Insert", err)
	}

	return nil
}

func (a *SubmissionRepo) UpdateAllocateMarkerSubmissions(ctx context.Context, db database.Ext, submissions []domain.Submission) error {
	ctx, span := interceptors.StartSpan(ctx, "SubmissionRepo.UpdateAllocateMarkerSubmissions")
	defer span.End()

	submissionHolder := dto.Submission{}
	table := submissionHolder.TableName()

	stmt := fmt.Sprintf("UPDATE %s SET allocated_marker_id = $2, updated_at = now() WHERE id = $1::TEXT AND deleted_at IS NULL", table)

	batch := &pgx.Batch{}
	for _, submission := range submissions {
		batch.Queue(stmt, submission.ID, submission.AllocatedMarkerID)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		ct, err := batchResults.Exec()

		if err != nil {
			return errors.NewDBError("SubmissionRepo.UpdateAllocateMarkerSubmissions Exec", err)
		}

		if ct.RowsAffected() == 0 {
			return errors.NewNoRowsUpdatedError("SubmissionRepo.UpdateAllocateMarkerSubmissions", nil)
		}
	}

	return nil
}
