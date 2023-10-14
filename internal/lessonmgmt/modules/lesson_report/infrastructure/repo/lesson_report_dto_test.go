package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocationFromEntity(t *testing.T) {
	t.Parallel()
	now := time.Time{}
	tcs := []struct {
		name         string
		lessonReport *domain.LessonReport
		dto          *LessonReportDTO
	}{
		{
			name: "full fields",
			lessonReport: &domain.LessonReport{
				LessonReportID:         "lesson-report-id-1",
				ReportSubmittingStatus: lesson_report_consts.ReportSubmittingStatusSaved,
				FormConfigID:           "form-config-id",
				LessonID:               "lesson-id-2",
				CreatedAt:              now,
				UpdatedAt:              now,
				FormConfig: &domain.FormConfig{
					FormConfigID: "form-config-id",
				},
			},
			dto: &LessonReportDTO{
				LessonReportID:         database.Text("lesson-report-id-1"),
				ReportSubmittingStatus: database.Text(string(lesson_report_consts.ReportSubmittingStatusSaved)),
				FormConfigID:           database.Text("form-config-id"),
				LessonID:               database.Text("lesson-id-2"),
				CreatedAt:              database.Timestamptz(now),
				UpdatedAt:              database.Timestamptz(now),
				DeletedAt:              pgtype.Timestamptz{Status: pgtype.Null},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewLessonReportDTOFromDomain(tc.lessonReport)
			require.NoError(t, err)
			assert.EqualValues(t, tc.dto, actual)
		})
	}
}
