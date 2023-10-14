package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LessonReportDetailRepo struct{}

func (l *LessonReportDetailRepo) GetByLessonReportID(ctx context.Context, db database.Ext, lessonReportID pgtype.Text) (entities.LessonReportDetails, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.GetByLessonReportID")
	defer span.End()

	fields, _ := (&entities.LessonReportDetail{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_report_details
		WHERE lesson_report_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	details := entities.LessonReportDetails{}
	err := database.Select(ctx, db, query, &lessonReportID).ScanAll(&details)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return details, nil
}

func (l *LessonReportDetailRepo) UpsertQueue(b *pgx.Batch, e *entities.LessonReportDetail) {
	fields, values := e.FieldMap()

	placeHolders := database.GeneratePlaceholders(len(fields))
	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT unique__lesson_report_id__student_id DO "+
		"UPDATE SET updated_at = now(), deleted_at = NULL", e.TableName(), strings.Join(fields, ", "), placeHolders)

	b.Queue(sql, values...)
}

// Upsert will update or insert lesson report details in a lesson report and remove details not in details args
func (l *LessonReportDetailRepo) Upsert(ctx context.Context, db database.Ext, lessonReportID pgtype.Text, details entities.LessonReportDetails) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(fmt.Sprintf(`UPDATE %s SET updated_at = now(), deleted_at = now() WHERE lesson_report_id = $1`, (&entities.LessonReportDetail{}).TableName()), lessonReportID)

	for i := range details {
		details[i].LessonReportID = lessonReportID
		l.UpsertQueue(b, details[i])
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *LessonReportDetailRepo) UpsertFieldValuesQueue(b *pgx.Batch, e *entities.PartnerDynamicFormFieldValue) {
	fields, values := e.FieldMap()

	placeHolders := database.GeneratePlaceholders(len(fields))
	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT unique__lesson_report_detail_id__field_id DO "+
		"UPDATE SET updated_at = now(), deleted_at = NULL, int_value = $8, string_value = $9, bool_value = $10, string_array_value = $11, int_array_value = $12, string_set_value = $13, int_set_value = $14, field_render_guide = $15",
		e.TableName(), strings.Join(fields, ", "), placeHolders)

	b.Queue(sql, values...)
}

func (l *LessonReportDetailRepo) UpsertFieldValues(ctx context.Context, db database.Ext, values []*entities.PartnerDynamicFormFieldValue) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.UpsertFieldValues")
	defer span.End()

	b := &pgx.Batch{}
	for i := range values {
		l.UpsertFieldValuesQueue(b, values[i])
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *LessonReportDetailRepo) DeleteByLessonReportID(ctx context.Context, db database.Ext, lessonReportID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.DeleteByLessonReport")
	defer span.End()

	query := "UPDATE lesson_report_details SET deleted_at = now(), updated_at = now() WHERE lesson_report_id = $1 AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &lessonReportID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonReportDetailRepo) DeleteFieldValuesByDetails(ctx context.Context, db database.Ext, detailIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.DeleteFieldValuesByDetails")
	defer span.End()

	query := "UPDATE partner_dynamic_form_field_values SET deleted_at = now(), updated_at = now() WHERE lesson_report_detail_id = $1 AND deleted_at IS NULL"
	b := &pgx.Batch{}
	for i := range detailIDs.Elements {
		b.Queue(query, &detailIDs.Elements[i])
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *LessonReportDetailRepo) GetFieldValuesByDetailIDs(ctx context.Context, db database.Ext, detailIDs pgtype.TextArray) (entities.PartnerDynamicFormFieldValues, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.GetFieldValuesByDetailIDs")
	defer span.End()

	fields, _ := (&entities.PartnerDynamicFormFieldValue{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM partner_dynamic_form_field_values
		WHERE lesson_report_detail_id = ANY($1) AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	values := entities.PartnerDynamicFormFieldValues{}
	err := database.Select(ctx, db, query, &detailIDs).ScanAll(&values)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return values, nil
}
