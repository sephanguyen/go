package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

// Lesson Report reflect lesson_reports table
type LessonReportDTO struct {
	LessonReportID         pgtype.Text
	ReportSubmittingStatus pgtype.Text
	CreatedAt              pgtype.Timestamptz
	UpdatedAt              pgtype.Timestamptz
	DeletedAt              pgtype.Timestamptz
	FormConfigID           pgtype.Text
	LessonID               pgtype.Text
}

// FieldMap Lesson Report table data fields
func (lr *LessonReportDTO) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_report_id",
			"report_submitting_status",
			"created_at",
			"updated_at",
			"deleted_at",
			"form_config_id",
			"lesson_id",
		}, []interface{}{
			&lr.LessonReportID,
			&lr.ReportSubmittingStatus,
			&lr.CreatedAt,
			&lr.UpdatedAt,
			&lr.DeletedAt,
			&lr.FormConfigID,
			&lr.LessonID,
		}
}

// TableName returns "lesson_reports"
func (lr *LessonReportDTO) TableName() string {
	return "lesson_reports"
}

func (lr *LessonReportDTO) PreUpdate() error {
	return lr.UpdatedAt.Set(time.Now())
}

type LessonReportDTOs []*LessonReportDTO

func (lr *LessonReportDTOs) Add() database.Entity {
	e := &LessonReportDTO{}
	*lr = append(*lr, e)

	return e
}

func NewLessonReportDTOFromDomain(l *domain.LessonReport) (*LessonReportDTO, error) {
	dto := &LessonReportDTO{}
	database.AllNullEntity(dto)
	formConfigId := l.FormConfigID
	if formConfigId == "" {
		formConfigId = l.FormConfig.FormConfigID
	}
	if err := multierr.Combine(
		dto.LessonReportID.Set(l.LessonReportID),
		dto.LessonID.Set(l.LessonID),
		dto.FormConfigID.Set(formConfigId),
		dto.ReportSubmittingStatus.Set(l.ReportSubmittingStatus),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from lesson report entity to lesson report dto: %w", err)
	}
	if len(l.LessonReportID) == 0 {
		err := dto.LessonReportID.Set(idutil.ULIDNow())
		if err != nil {
			return nil, fmt.Errorf("could not set ID to lesson report dto: %w", err)
		}
	}
	if dto.ReportSubmittingStatus.String != string(lesson_report_consts.ReportSubmittingStatusApproved) &&
		dto.ReportSubmittingStatus.String != string(lesson_report_consts.ReportSubmittingStatusSaved) &&
		dto.ReportSubmittingStatus.String != string(lesson_report_consts.ReportSubmittingStatusSubmitted) {
		return nil, fmt.Errorf("NewLessonReportFromEntity: invalid submitting status")
	}
	return dto, nil
}

func (lr *LessonReportDTO) ToLessonReportDomain() (*domain.LessonReport, error) {
	lessonReportEntity, err := domain.NewLessonReport().
		WithLessonID(lr.LessonID.String).
		WithLessonReportID(lr.LessonReportID.String).
		WithFormConfigID(lr.FormConfigID.String).
		WithModificationTime(lr.CreatedAt.Time, lr.UpdatedAt.Time).
		WithReportSubmittingStatus(lesson_report_consts.ReportSubmittingStatus(lr.ReportSubmittingStatus.String)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("Error ToLessonReportDomain: %w", err)
	}
	return lessonReportEntity, nil
}
