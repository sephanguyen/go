package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type LessonReportDetailRepo struct{}

func (l *LessonReportDetailRepo) GetByLessonReportID(ctx context.Context, db database.Ext, lessonReportID string) (domain.LessonReportDetails, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.GetByLessonReportID")
	defer span.End()

	fields, _ := (&LessonReportDetailDTO{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_report_details
		WHERE lesson_report_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	details := LessonReportDetailDTOs{}
	err := database.Select(ctx, db, query, &lessonReportID).ScanAll(&details)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	var detailDomains domain.LessonReportDetails
	for i := 0; i < len(details); i++ {
		detail, err := domain.NewLessonReportDetailBuilder().
			WithLessonReportDetailID(details[i].LessonReportDetailID.String).
			WithLessonReportID(details[i].LessonReportID.String).
			WithStudentID(details[i].StudentID.String).
			WithModificationTime(details[i].CreatedAt.Time, details[i].UpdatedAt.Time).
			Build()
		if err != nil {
			return nil, fmt.Errorf("Error parsing DTO to domain %w", err)
		}
		detailDomains = append(detailDomains, detail)
	}

	return detailDomains, nil
}

func (l *LessonReportDetailRepo) GetDetailByLessonReportID(ctx context.Context, db database.Ext, lessonReportID string) (domain.LessonReportDetails, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.GetByLessonReportID")
	defer span.End()

	query := fmt.Sprintf(`SELECT lrd.lesson_report_id, lrd.student_id, lrd.created_at, lrd.updated_at, lrd.deleted_at, 
			lrd.lesson_report_detail_id, lm.attendance_status, lm.attendance_remark, lrd.report_version
		FROM lesson_report_details lrd 
		INNER JOIN lesson_reports lr on lrd.lesson_report_id  = lr.lesson_report_id and lrd.lesson_report_id = $1 AND lrd.deleted_at IS NULL
		INNER JOIN lesson_members lm  on lr.lesson_id  = lm.lesson_id and lrd.student_id = lm.user_id  
		`,
	)
	details := LessonReportDetailWithAttendanceStatusDTOs{}

	err := database.Select(ctx, db, query, &lessonReportID).ScanAll(&details)
	detailDomains := make(domain.LessonReportDetails, 0, len(details))
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	for i := 0; i < len(details); i++ {
		detail, err := domain.NewLessonReportDetailBuilder().
			WithLessonReportDetailID(details[i].LessonReportDetailID.String).
			WithLessonReportID(details[i].LessonReportID.String).
			WithStudentID(details[i].StudentID.String).
			WithModificationTime(details[i].CreatedAt.Time, details[i].UpdatedAt.Time).
			WithAttendanceStatus(constant.StudentAttendStatus(details[i].AttendanceStatus.String)).
			WithAttendanceRemark(details[i].AttendanceRemark.String).
			WithReportVersion(int(details[i].ReportVersion.Int)).
			Build()
		if err != nil {
			return detailDomains, fmt.Errorf("Error parsing DTO to domain %w", err)
		}
		detailDomains = append(detailDomains, detail)
	}

	return detailDomains, nil
}

func (l *LessonReportDetailRepo) GetReportVersionByLessonID(ctx context.Context, db database.Ext, lessonID string) (domain.LessonReportDetails, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.GetReportVersionByLessonID")
	defer span.End()

	query := `SELECT lrd.lesson_report_id, lrd.student_id, lrd.created_at, lrd.updated_at, lrd.deleted_at, 
		lrd.lesson_report_detail_id, lrd.report_version
		FROM lesson_report_details lrd
		INNER JOIN lesson_reports lr on lrd.lesson_report_id = lr.lesson_report_id
		WHERE lr.lesson_id = $1`

	details := LessonReportDetailWithAttendanceStatusDTOs{}
	err := database.Select(ctx, db, query, &lessonID).ScanAll(&details)

	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	detailDomains := make(domain.LessonReportDetails, 0, len(details))
	for i := 0; i < len(details); i++ {
		detail, err := domain.NewLessonReportDetailBuilder().
			WithLessonReportDetailID(details[i].LessonReportDetailID.String).
			WithLessonReportID(details[i].LessonReportID.String).
			WithStudentID(details[i].StudentID.String).
			WithModificationTime(details[i].CreatedAt.Time, details[i].UpdatedAt.Time).
			WithAttendanceStatus(constant.StudentAttendStatus(details[i].AttendanceStatus.String)).
			WithAttendanceRemark(details[i].AttendanceRemark.String).
			WithReportVersion(int(details[i].ReportVersion.Int)).
			Build()
		if err != nil {
			return detailDomains, fmt.Errorf("error parsing DTO to domain %w", err)
		}
		detailDomains = append(detailDomains, detail)
	}

	return detailDomains, nil
}

// Upsert will update or insert lesson report details in a lesson report and remove details not in details args
func (l *LessonReportDetailRepo) Upsert(ctx context.Context, db database.Ext, lessonReportID string, details domain.LessonReportDetails) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(fmt.Sprintf(`UPDATE %s SET updated_at = now(), deleted_at = now() WHERE lesson_report_id = $1`, (&LessonReportDetailDTO{}).TableName()), lessonReportID)

	for i := range details {
		details[i].LessonReportID = lessonReportID
		detailDTO := &LessonReportDetailDTO{}
		database.AllNullEntity(detailDTO)
		if err := multierr.Combine(
			detailDTO.LessonReportDetailID.Set(details[i].LessonReportDetailID),
			detailDTO.LessonReportID.Set(details[i].LessonReportID),
			detailDTO.StudentID.Set(details[i].StudentID),
			detailDTO.CreatedAt.Set(details[i].CreatedAt),
			detailDTO.UpdatedAt.Set(details[i].UpdatedAt),
			detailDTO.ReportVersion.Set(details[i].ReportVersion),
		); err != nil {
			return fmt.Errorf("could not mapping from lesson report detail entity to lesson report dto: %w", err)
		}
		l.UpsertQueue(b, detailDTO)
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

// UpsertWithVersion : with version will update or insert lesson report details in a lesson report and remove details not in details args
func (l *LessonReportDetailRepo) UpsertWithVersion(ctx context.Context, db database.Ext, lessonReportID string, details domain.LessonReportDetails) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.Upsert")
	defer span.End()

	lessonReportStudentIds := make([]string, 0, len(details))
	for i := range details {
		lessonReportStudentIds = append(lessonReportStudentIds, details[i].StudentID)
	}
	concatenatedStudentIDs := strings.Join(lessonReportStudentIds, "','")

	b := &pgx.Batch{}
	b.Queue(fmt.Sprintf(`UPDATE %s SET updated_at = now(), deleted_at = now(), report_version = 0 WHERE lesson_report_id = $1 AND student_id NOT IN ('%s')`, (&LessonReportDetailDTO{}).TableName(), concatenatedStudentIDs), lessonReportID)

	for i := range details {
		details[i].LessonReportID = lessonReportID
		detailDTO := &LessonReportDetailDTO{}
		database.AllNullEntity(detailDTO)
		if err := multierr.Combine(
			detailDTO.LessonReportDetailID.Set(details[i].LessonReportDetailID),
			detailDTO.LessonReportID.Set(details[i].LessonReportID),
			detailDTO.StudentID.Set(details[i].StudentID),
			detailDTO.CreatedAt.Set(details[i].CreatedAt),
			detailDTO.UpdatedAt.Set(details[i].UpdatedAt),
			detailDTO.ReportVersion.Set(details[i].ReportVersion),
		); err != nil {
			return fmt.Errorf("could not mapping from lesson report detail entity to lesson report dto: %w", err)
		}
		l.UpsertQueueWithVersion(b, detailDTO)
	}

	result := db.SendBatch(ctx, b)

	defer result.Close()
	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		data, err := result.Exec()
		if i > 0 && data.RowsAffected() == 0 {
			return domain.ErrReportVersionIsOutOfDate
		}
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *LessonReportDetailRepo) UpsertOne(ctx context.Context, db database.Ext, lessonReportID string, details domain.LessonReportDetail) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}

	details.LessonReportID = lessonReportID
	detailDTO := &LessonReportDetailDTO{}
	database.AllNullEntity(detailDTO)
	if err := multierr.Combine(
		detailDTO.LessonReportDetailID.Set(details.LessonReportDetailID),
		detailDTO.LessonReportID.Set(details.LessonReportID),
		detailDTO.StudentID.Set(details.StudentID),
		detailDTO.CreatedAt.Set(details.CreatedAt),
		detailDTO.UpdatedAt.Set(details.UpdatedAt),
		detailDTO.ReportVersion.Set(details.ReportVersion),
	); err != nil {
		return fmt.Errorf("could not mapping from lesson report detail entity to lesson report dto: %w", err)
	}
	l.UpsertQueueWithVersion(b, detailDTO)

	result := db.SendBatch(ctx, b)

	defer result.Close()
	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		data, err := result.Exec()
		if data.RowsAffected() == 0 {
			return domain.ErrReportVersionIsOutOfDate
		}
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *LessonReportDetailRepo) UpsertQueue(b *pgx.Batch, e *LessonReportDetailDTO) {
	fields, values := e.FieldMap()

	placeHolders := database.GeneratePlaceholders(len(fields))
	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT unique__lesson_report_id__student_id DO "+
		"UPDATE SET updated_at = now(), deleted_at = NULL", e.TableName(), strings.Join(fields, ", "), placeHolders)

	b.Queue(sql, values...)
}

func (l *LessonReportDetailRepo) UpsertQueueWithVersion(b *pgx.Batch, e *LessonReportDetailDTO) {
	fields, values := e.FieldMap()

	placeHolders := database.GeneratePlaceholders(len(fields))

	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT unique__lesson_report_id__student_id DO "+
		"UPDATE SET updated_at = now(), deleted_at = NULL, report_version = lesson_report_details.report_version + 1 WHERE lesson_report_details.deleted_at is not null OR (lesson_report_details.deleted_at is null AND lesson_report_details.report_version = excluded.report_version)", e.TableName(), strings.Join(fields, ", "), placeHolders)

	b.Queue(sql, values...)
}

func (l *LessonReportDetailRepo) UpsertFieldValuesQueue(b *pgx.Batch, detailDTO *PartnerDynamicFormFieldValueDTO) {
	fields, values := detailDTO.FieldMap()

	placeHolders := database.GeneratePlaceholders(len(fields))
	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT unique__lesson_report_detail_id__field_id DO "+
		"UPDATE SET updated_at = now(), deleted_at = NULL, int_value = $8, string_value = $9, bool_value = $10, string_array_value = $11, int_array_value = $12, string_set_value = $13, int_set_value = $14, field_render_guide = $15",
		detailDTO.TableName(), strings.Join(fields, ", "), placeHolders)

	b.Queue(sql, values...)
}

func (l *LessonReportDetailRepo) UpsertFieldValues(ctx context.Context, db database.Ext, values []*domain.PartnerDynamicFormFieldValue) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.UpsertFieldValues")
	defer span.End()

	b := &pgx.Batch{}
	for i := range values {
		domain := values[i]
		detailDTO := &PartnerDynamicFormFieldValueDTO{}
		database.AllNullEntity(detailDTO)
		if err := multierr.Combine(
			detailDTO.LessonReportDetailID.Set(domain.LessonReportDetailID),
			detailDTO.DynamicFormFieldValueID.Set(domain.DynamicFormFieldValueID),
			detailDTO.FieldID.Set(domain.FieldID),
			detailDTO.ValueType.Set(domain.ValueType),
			detailDTO.IntValue.Set(int32(domain.IntValue)),
			detailDTO.StringValue.Set(domain.StringValue),
			detailDTO.IntArrayValue.Set(domain.IntArrayValue),
			detailDTO.BoolValue.Set(domain.BoolValue),
			detailDTO.StringArrayValue.Set(domain.StringArrayValue),
			detailDTO.StringSetValue.Set(domain.StringSetValue),
			detailDTO.IntSetValue.Set(domain.IntSetValue),
			detailDTO.FieldRenderGuide.Set(domain.FieldRenderGuide),
			detailDTO.CreatedAt.Set(domain.CreatedAt),
			detailDTO.UpdatedAt.Set(domain.UpdatedAt),
		); err != nil {
			return fmt.Errorf("could not mapping from PartnerDynamicFormFieldValueDTO: %w", err)
		}
		l.UpsertFieldValuesQueue(b, detailDTO)
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
