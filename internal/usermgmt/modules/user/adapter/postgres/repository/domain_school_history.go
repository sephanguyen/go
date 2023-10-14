package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DomainSchoolHistoryRepo struct{}

type SchoolHistoryAttribute struct {
	UserID         field.String
	SchoolID       field.String
	SchoolCourseID field.String
	IsCurrent      field.Boolean
	StartDate      field.Time
	EndDate        field.Time
	OrganizationID field.String
}

type SchoolHistory struct {
	SchoolHistoryAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewDomainSchoolHistory(sh entity.DomainSchoolHistory) *SchoolHistory {
	now := field.NewTime(time.Now())
	return &SchoolHistory{
		SchoolHistoryAttribute: SchoolHistoryAttribute{
			UserID:         sh.UserID(),
			SchoolID:       sh.SchoolID(),
			SchoolCourseID: sh.SchoolCourseID(),
			IsCurrent:      sh.IsCurrent(),
			StartDate:      sh.StartDate(),
			EndDate:        sh.EndDate(),
			OrganizationID: sh.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (sh *SchoolHistory) UserID() field.String {
	return sh.SchoolHistoryAttribute.UserID
}
func (sh *SchoolHistory) SchoolID() field.String {
	return sh.SchoolHistoryAttribute.SchoolID
}
func (sh *SchoolHistory) SchoolCourseID() field.String {
	return sh.SchoolHistoryAttribute.SchoolCourseID
}
func (sh *SchoolHistory) IsCurrent() field.Boolean {
	return sh.SchoolHistoryAttribute.IsCurrent
}
func (sh *SchoolHistory) StartDate() field.Time {
	return sh.SchoolHistoryAttribute.StartDate
}
func (sh *SchoolHistory) EndDate() field.Time {
	return sh.SchoolHistoryAttribute.EndDate
}
func (sh *SchoolHistory) OrganizationID() field.String {
	return sh.SchoolHistoryAttribute.OrganizationID
}

func (*SchoolHistory) TableName() string {
	return "school_history"
}

func (sh *SchoolHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"school_id",
			"school_course_id",
			"is_current",
			"start_date",
			"end_date",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&sh.SchoolHistoryAttribute.UserID,
			&sh.SchoolHistoryAttribute.SchoolID,
			&sh.SchoolHistoryAttribute.SchoolCourseID,
			&sh.SchoolHistoryAttribute.IsCurrent,
			&sh.SchoolHistoryAttribute.StartDate,
			&sh.SchoolHistoryAttribute.EndDate,
			&sh.CreatedAt,
			&sh.UpdatedAt,
			&sh.DeletedAt,
			&sh.SchoolHistoryAttribute.OrganizationID,
		}
}

func (repo *DomainSchoolHistoryRepo) SetCurrentSchoolByStudentIDsAndSchoolIDs(ctx context.Context, db database.QueryExecer, studentIDs, schoolIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolHistoryRepo.SetCurrentSchoolByStudentIDsAndSchoolIDs")
	defer span.End()

	sql := `UPDATE school_history SET is_current = true WHERE school_id = ANY($1) AND student_id = ANY($2) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, database.TextArray(schoolIDs), database.TextArray(studentIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (repo *DomainSchoolHistoryRepo) SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolHistoryRepo.SoftDeleteByStudentIDs")
	defer span.End()

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE student_id = ANY($1) AND deleted_at IS NULL`, (&SchoolHistory{}).TableName())
	_, err := db.Exec(ctx, sql, database.TextArray(studentIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (repo *DomainSchoolHistoryRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, schoolHistories ...entity.DomainSchoolHistory) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolHistoryRepo.UpsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}

	queueFn := func(b *pgx.Batch, schoolHistory *SchoolHistory) {
		fields, values := schoolHistory.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT school_history__pk 
			DO UPDATE SET school_course_id = EXCLUDED.school_course_id, start_date = EXCLUDED.start_date, is_current = EXCLUDED.is_current, end_date = EXCLUDED.end_date, updated_at = now(), deleted_at = NULL`,
			schoolHistory.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	for _, schoolHistory := range schoolHistories {
		repoSchoolHistory := NewDomainSchoolHistory(schoolHistory)

		queueFn(batch, repoSchoolHistory)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(schoolHistories); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("school_history was not upserted")
		}
	}

	return nil
}
