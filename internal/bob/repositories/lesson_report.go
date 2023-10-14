package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LessonReportRepo struct{}

type ListIndividualLessonReportArgs struct {
	Limit          uint32
	LessonReportID pgtype.Text
	SchoolID       pgtype.Int4
}

func (l *LessonReportRepo) FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.LessonReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.FindByID")
	defer span.End()

	lessonReport := &entities.LessonReport{}
	fields, values := lessonReport.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM lesson_reports
		WHERE lesson_report_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return lessonReport, nil
}

func (l *LessonReportRepo) Create(ctx context.Context, db database.Ext, report *entities.LessonReport) (*entities.LessonReport, error) {
	fieldNames, args := report.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		report.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return nil, err
	}

	return report, nil
}

func (l *LessonReportRepo) Update(ctx context.Context, db database.Ext, report *entities.LessonReport) (*entities.LessonReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.Update")
	defer span.End()

	err := report.PreUpdate()
	if err != nil {
		return nil, fmt.Errorf("lesson_report.PreUpdate: %s", err)
	}

	updatedFields := []string{
		"report_submitting_status",
		"updated_at",
		"form_config_id",
		"lesson_id",
	}

	cmd, err := database.UpdateFields(ctx, report, db.Exec, "lesson_report_id", updatedFields)
	if err != nil {
		return nil, fmt.Errorf("database.Update: %s", err)
	}
	if cmd.RowsAffected() != 1 {
		return nil, fmt.Errorf("expect 1 row affected, got %d", cmd.RowsAffected())
	}
	return report, nil
}

func (l *LessonReportRepo) UpdateLessonReportSubmittingStatusByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, status entities.ReportSubmittingStatus) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.UpdateLessonReportSubmittingStatusByID")
	defer span.End()

	query := "UPDATE lesson_reports SET report_submitting_status =$1, updated_at = NOW() " +
		"WHERE lesson_report_id = $2 " +
		"AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &status, &id)
	if err != nil {
		return fmt.Errorf("db.QueryRow: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("LessonReportRepo.UpdateLessonReportSubmittingStatusByID: Can't update lesson report; is the report submitted and not yet deleted?")
	}

	return nil
}

func (l *LessonReportRepo) Delete(ctx context.Context, db database.Ext, id pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.Delete")
	defer span.End()

	query := "UPDATE lesson_reports SET deleted_at = now(), updated_at = now() WHERE lesson_report_id = $1 AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &id)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonReportRepo) FindByLessonID(ctx context.Context, db database.Ext, lessonID pgtype.Text) (*entities.LessonReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.FindByLessonID")
	defer span.End()

	lessonReport := &entities.LessonReport{}
	fields, values := lessonReport.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM lesson_reports
		WHERE lesson_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &lessonID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return lessonReport, nil
}

func (l *LessonReportRepo) DeleteReportsBelongToLesson(ctx context.Context, db database.Ext, lessonID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.DeleteReportsBelongToLesson")
	defer span.End()

	query := "UPDATE lesson_reports SET deleted_at = now(), updated_at = now() WHERE lesson_id = $1 AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}
