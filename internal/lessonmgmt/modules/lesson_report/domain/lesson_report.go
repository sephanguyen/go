package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"

	"github.com/jackc/pgx/v4"
)

type LessonReportStatus string

const (
	ReportStatusDraft     LessonReportStatus = "LESSON_REPORT_STATUS_DRAFT"
	ReportStatusSubmitted LessonReportStatus = "LESSON_REPORT_STATUS_SUBMITTED"
	ReportStatusNone      LessonReportStatus = "LESSON_REPORT_STATUS_NONE"
)

// Lesson Report reflect lesson_reports table
type LessonReport struct {
	LessonReportID         string
	ReportSubmittingStatus lesson_report_consts.ReportSubmittingStatus
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              *time.Time
	FormConfigID           string
	LessonID               string
	Details                LessonReportDetails
	FormConfig             *FormConfig
	UnleashToggles         map[string]bool

	Lesson              *lesson_domain.Lesson
	FeatureName         string
	IsUpdateMembersInfo bool
	IsSavePerStudent    bool
}
type LessonReports []*LessonReport

type LessonReportBuilder struct {
	lessonReport *LessonReport
}

var NotFoundDBErr = fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error()

func NewLessonReport() *LessonReportBuilder {
	return &LessonReportBuilder{
		lessonReport: &LessonReport{},
	}
}

func (l *LessonReportBuilder) Build() (*LessonReport, error) {
	return l.lessonReport, nil
}

func (l *LessonReportBuilder) WithLessonReportID(lessonReportID string) *LessonReportBuilder {
	l.lessonReport.LessonReportID = lessonReportID
	return l
}
func (l *LessonReportBuilder) WithReportSubmittingStatus(submittingStatus lesson_report_consts.ReportSubmittingStatus) *LessonReportBuilder {
	l.lessonReport.ReportSubmittingStatus = submittingStatus
	return l
}

func (l *LessonReportBuilder) WithFormConfigID(formConfigID string) *LessonReportBuilder {
	l.lessonReport.FormConfigID = formConfigID
	return l
}

func (l *LessonReportBuilder) WithLessonID(lessonID string) *LessonReportBuilder {
	l.lessonReport.LessonID = lessonID
	return l
}

func (l *LessonReportBuilder) WithModificationTime(createdAt, updatedAt time.Time) *LessonReportBuilder {
	l.lessonReport.CreatedAt = createdAt
	l.lessonReport.UpdatedAt = updatedAt
	return l
}

func (l *LessonReportBuilder) WithDeletedTime(deletedAt *time.Time) *LessonReportBuilder {
	l.lessonReport.DeletedAt = deletedAt
	return l
}

func (l *LessonReport) IsValid(ctx context.Context, db database.Ext, isDraft bool) error {
	var err error
	if len(l.LessonID) == 0 {
		return fmt.Errorf("lesson_id could not be empty")
	}

	if len(l.ReportSubmittingStatus) == 0 {
		return fmt.Errorf("submitting_status could not be empty")
	}

	allowFields := make(map[string]*FormConfigField)
	if l.FormConfig != nil {
		if err := l.FormConfig.IsValid(); err != nil {
			return fmt.Errorf("invalid form_config: %v", err)
		}
		// get field ids from form config
		allowFields = l.FormConfig.GetFieldsMap()
	}
	var attendanceConfig FormConfigField
	if isToggled, ok := l.UnleashToggles["Lesson_LessonManagement_BackOffice_ValidationLessonBeforeCompleted"]; ok && isToggled && !isDraft {
		attendanceConfig.FieldID = "attendance_status"
		attendanceConfig.ValueType = lesson_report_consts.FieldValueTypeString
		attendanceConfig.IsRequired = true
		allowFields[attendanceConfig.FieldID] = &attendanceConfig
	}
	if err = l.Details.OnlyHaveAllowFields(allowFields); err != nil {
		return err
	}

	if err = l.Details.IsValid(); err != nil {
		return fmt.Errorf("invalid details: %v", err)
	}

	if !isDraft {
		// check required field's values of each details
		requiredFields := make(map[string]*FormConfigField)
		if l.FormConfig != nil {
			requiredFields = l.FormConfig.GetRequiredFieldsMap()
		}
		if err = l.Details.ValidateRequiredFieldsValue(requiredFields); err != nil {
			return err
		}
	}

	return nil
}
