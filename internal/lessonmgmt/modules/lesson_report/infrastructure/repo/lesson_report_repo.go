package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
)

type LessonReportRepo struct{}

func (l *LessonReportRepo) DeleteReportsBelongToLesson(ctx context.Context, db database.Ext, lessonIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.DeleteReportsBelongToLesson")
	defer span.End()

	query := `with deleted_report as (
		UPDATE lesson_reports lr SET deleted_at = now(), updated_at = now() 
		WHERE lesson_id = ANY($1) AND deleted_at IS NULL
		RETURNING lr.lesson_report_id
	) update lesson_report_details lrd
		set deleted_at = now(), report_version = 0
		from deleted_report dr where dr.lesson_report_id = lrd.lesson_report_id 
	`

	_, err := db.Exec(ctx, query, &lessonIDs)

	return err
}

func (l *LessonReportRepo) Create(ctx context.Context, db database.Ext, report *domain.LessonReport) (*domain.LessonReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.Create")
	defer span.End()

	lessonReportDTO, err := NewLessonReportDTOFromDomain(report)
	if err != nil {
		return nil, err
	}
	fieldNames, args := lessonReportDTO.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		lessonReportDTO.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return nil, err
	}
	report.LessonReportID = lessonReportDTO.LessonReportID.String
	return report, nil
}

func (l *LessonReportRepo) FindByID(ctx context.Context, db database.Ext, id string) (*domain.LessonReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.FindByID")
	defer span.End()
	lessonReportDTO := &LessonReportDTO{}
	fields, values := lessonReportDTO.FieldMap()
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
	return lessonReportDTO.ToLessonReportDomain()
}

func (l *LessonReportRepo) FindByLessonID(ctx context.Context, db database.Ext, lessonID string) (*domain.LessonReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.FindByLessonID")
	defer span.End()

	lessonReportDTO := &LessonReportDTO{}
	fields, values := lessonReportDTO.FieldMap()
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

	return lessonReportDTO.ToLessonReportDomain()
}

func (l *LessonReportRepo) Update(ctx context.Context, db database.Ext, lessonReport *domain.LessonReport) (*domain.LessonReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.Update")
	defer span.End()
	lessonReportDTO, err := NewLessonReportDTOFromDomain(lessonReport)
	if err != nil {
		return nil, err
	}
	err = lessonReportDTO.PreUpdate()
	if err != nil {
		return nil, fmt.Errorf("lesson_report.PreUpdate: %s", err)
	}

	updatedFields := []string{
		"report_submitting_status",
		"updated_at",
		"form_config_id",
		"lesson_id",
	}

	cmd, err := database.UpdateFields(ctx, lessonReportDTO, db.Exec, "lesson_report_id", updatedFields)
	if err != nil {
		return nil, fmt.Errorf("database.Update: %s", err)
	}
	if cmd.RowsAffected() != 1 {
		return nil, fmt.Errorf("expect 1 row affected, got %d", cmd.RowsAffected())
	}
	return lessonReport, nil
}

func (l *LessonReportRepo) FindByResourcePath(ctx context.Context, db database.Ext, resourcePath string, limit int, offSet int) (domain.LessonReports, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.FindByResourcePath")
	defer span.End()
	values := LessonReportDTOs{}

	query := fmt.Sprintf(`
		SELECT l.lesson_report_id, 
			l.report_submitting_status, l.created_at, l.updated_at,
			l.deleted_at, l.form_config_id, l.lesson_id
		FROM lesson_reports l
			JOIN partner_form_configs pfc 
			ON l.form_config_id = pfc.form_config_id 
				AND pfc.feature_name = 'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT'
				AND pfc.deleted_at IS NULL
		WHERE l.deleted_at IS NULL
		AND l.resource_path = $1
		ORDER BY l.lesson_report_id 
		LIMIT $2 OFFSET $3`,
	)
	err := database.Select(ctx, db, query, &resourcePath, &limit, &offSet).ScanAll(&values)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	lessonReports := make(domain.LessonReports, 0, len(values))
	for _, v := range values {
		lessonReport, _ := v.ToLessonReportDomain()
		lessonReports = append(lessonReports, lessonReport)
	}
	return lessonReports, nil
}

func (l *LessonReportRepo) DeleteLessonReportWithoutStudent(ctx context.Context, db database.Ext, lessonID []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportRepo.DeleteLessonReportWithoutStudent")
	defer span.End()

	query := `with deleted_report AS ( WITH lessonstudent AS(
				SELECT lesson_id from lessons WHERE lesson_id = ANY($1)
				EXCEPT
				SELECT lm.lesson_id
				FROM lesson_members lm 
				WHERE lm.lesson_id = ANY($1) and lm.deleted_at is null) 
		update lesson_reports lr
		set deleted_at = now() 
		from  lessonstudent WHERE lr.lesson_id = lessonstudent.lesson_id and lr.deleted_at is null
		RETURNING lr.lesson_report_id
		) update lesson_report_details lrd
		set deleted_at = now(), report_version = 0
		from deleted_report dr where dr.lesson_report_id = lrd.lesson_report_id`

	_, err := db.Exec(ctx, query, lessonID)
	return err
}
