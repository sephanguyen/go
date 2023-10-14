package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type StudentEntryExitRecordsRepo struct {
}

// Create student_entryexit_records entity
func (r *StudentEntryExitRecordsRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentEntryExitRecords) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentEntryExitRecordsRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"entryexit_id", "deleted_at", "resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert StudentEntryExitRecordsRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert StudentEntryExitRecordsRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *StudentEntryExitRecordsRepo) LockAdvisoryByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEntryExitRecordsRepo.LockAdvisoryByStudentID")
	defer span.End()

	e := &entities.StudentEntryExitRecords{}
	lockID := fmt.Sprintf("%s-%s", e.TableName(), studentID)

	query := "SELECT pg_try_advisory_lock(hashtext($1))"
	var lockAcquired bool

	err := db.QueryRow(ctx, query, lockID).Scan(&lockAcquired)
	if err != nil {
		return false, fmt.Errorf("err LockAdvisoryByStudentID StudentEntryExitRecordsRepo: %w - studentID: %s", err, studentID)
	}
	return lockAcquired, nil
}

func (r *StudentEntryExitRecordsRepo) UnLockAdvisoryByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentEntryExitRecordsRepo.UnLockAdvisoryByStudentID")
	defer span.End()

	e := &entities.StudentEntryExitRecords{}
	lockID := fmt.Sprintf("%s-%s", e.TableName(), studentID)

	query := "SELECT pg_advisory_unlock(hashtext($1))"

	_, err := db.Exec(ctx, query, lockID)
	if err != nil {
		return fmt.Errorf("err UnLockAdvisoryByStudentID StudentEntryExitRecordsRepo: %w - studentID: %s", err, studentID)
	}

	return nil
}

// GetLatestRecords retrieves the latest entryexit record of student
func (r *StudentEntryExitRecordsRepo) GetLatestRecordByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentEntryExitRecords, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEntryExitRecordsRepo.GetLatestRecordByID")
	defer span.End()

	e := &entities.StudentEntryExitRecords{}
	fieldNames, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 And deleted_at IS NULL ORDER BY entry_at DESC LIMIT 1", strings.Join(fieldNames, ", "), e.TableName())

	if err := database.Select(ctx, db, query, studentID).ScanOne(e); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("err GetLatestRecordByID StudentEntryExitRecordsRepo: %w", err)
	}

	return e, nil
}

// Update updates StudentEntryExitRecords entity
func (r *StudentEntryExitRecordsRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.StudentEntryExitRecords) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentEntryExitRecordsRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "entryexit_id", []string{"entry_at", "exit_at", "updated_at"})

	if err != nil {
		return fmt.Errorf("err update StudentEntryExitRecordsRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update StudentEntryExitRecordsRepo: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

// SoftDeleteByID soft deletes StudentEntryExitRecords entity
func (r *StudentEntryExitRecordsRepo) SoftDeleteByID(ctx context.Context, db database.QueryExecer, id pgtype.Int4) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentEntryExitRecordsRepo.SoftDeleteByID")
	defer span.End()

	e := &entities.StudentEntryExitRecords{}

	now := time.Now()
	if err := multierr.Combine(
		e.ID.Set(id),
		e.UpdatedAt.Set(now),
		e.DeletedAt.Set(now),
	); err != nil {
		return fmt.Errorf("err delete multierr: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "entryexit_id", []string{"deleted_at", "updated_at"})
	if err != nil {
		return fmt.Errorf("err delete StudentEntryExitRecordsRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err delete StudentEntryExitRecordsRepo: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

type RetrieveEntryExitRecordFilter struct {
	StudentID    pgtype.Text
	RecordFilter eepb.RecordFilter
	Limit        pgtype.Int8
	Offset       pgtype.Int8
}

func (r *StudentEntryExitRecordsRepo) RetrieveRecordsByStudentID(ctx context.Context, db database.QueryExecer, filter RetrieveEntryExitRecordFilter) ([]*entities.StudentEntryExitRecords, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEntryExitRecordsRepo.RetrieveRecordsByStudentID")
	defer span.End()

	e := &entities.StudentEntryExitRecords{}
	filterType := generateRetrieveRecordsWhereClause(filter.RecordFilter)
	query := fmt.Sprintf("SELECT entryexit_id, entry_at, exit_at FROM %s WHERE %s ORDER BY entry_at DESC LIMIT $2 OFFSET $3", e.TableName(), filterType)
	rows, err := db.Query(ctx, query, &filter.StudentID, &filter.Limit, &filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("err retrieve records StudentEntryExitRecordsRepo: %w", err)
	}
	defer rows.Close()
	var result []*entities.StudentEntryExitRecords
	for rows.Next() {
		var entryExitRecord entities.StudentEntryExitRecords
		if err := rows.Scan(&entryExitRecord.ID, &entryExitRecord.EntryAt, &entryExitRecord.ExitAt); err != nil {
			return nil, err
		}
		result = append(result, &entryExitRecord)
	}
	return result, nil
}

func generateRetrieveRecordsWhereClause(filterRecord eepb.RecordFilter) string {
	var filterType string
	switch filterRecord {
	case eepb.RecordFilter_ALL:
		filterType = "student_id = $1 And deleted_at IS NULL"
	case eepb.RecordFilter_THIS_MONTH:
		filterType = "student_id = $1 And EXTRACT(year FROM entry_at) = EXTRACT(year FROM CURRENT_DATE) And EXTRACT(month FROM entry_at) = extract (month FROM CURRENT_DATE) And deleted_at IS NULL"
	case eepb.RecordFilter_LAST_MONTH:
		filterType = "student_id = $1 And EXTRACT(year FROM entry_at) = EXTRACT(year FROM (CURRENT_DATE - INTERVAL '1 month')) And EXTRACT(month FROM entry_at) = EXTRACT(month FROM (CURRENT_DATE - INTERVAL '1 month')) And deleted_at IS NULL"
	case eepb.RecordFilter_THIS_YEAR:
		filterType = "student_id = $1 And EXTRACT(year FROM entry_at) = EXTRACT(year FROM CURRENT_DATE) And deleted_at IS NULL"
	}
	return filterType
}
