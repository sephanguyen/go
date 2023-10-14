package repo

import (
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/domain"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
)

type LessonAllocationByWeek struct {
	AcademicWeekID   pgtype.Text
	LessonID         pgtype.Text
	StartTime        pgtype.Timestamptz
	EndTime          pgtype.Timestamptz
	LocationID       pgtype.Text
	LessonStatus     pgtype.Text
	TeachingMethod   pgtype.Text
	AttendanceStatus pgtype.Text
	LessonReportID   pgtype.Text
	IsLocked         pgtype.Bool
}

func (l *LessonAllocationByWeek) FieldMap() ([]string, []interface{}) {
	return []string{
			"academic_week_id",
			"lesson_id",
			"start_time",
			"end_time",
			"attendance_status",
			"scheduling_status",
			"center_id",
			"teaching_method",
			"lesson_report_id",
			"is_locked",
		}, []interface{}{
			&l.AcademicWeekID,
			&l.LessonID,
			&l.StartTime,
			&l.EndTime,
			&l.AttendanceStatus,
			&l.LessonStatus,
			&l.LocationID,
			&l.TeachingMethod,
			&l.LessonReportID,
			&l.IsLocked,
		}
}

type LessonAllocationByWeeks []*LessonAllocationByWeek

func (law LessonAllocationByWeeks) ToLessonAllocation() map[string][]*domain.LessonAllocationInfo {
	res := make(map[string][]*domain.LessonAllocationInfo, 0)
	for _, l := range law {
		res[l.AcademicWeekID.String] = append(res[l.AcademicWeekID.String], &domain.LessonAllocationInfo{
			LessonID:         l.LessonID.String,
			StartTime:        l.StartTime.Time,
			EndTime:          l.EndTime.Time,
			LocationID:       l.LocationID.String,
			AttendanceStatus: lesson_domain.StudentAttendStatus(l.AttendanceStatus.String),
			Status:           lesson_domain.LessonSchedulingStatus(l.LessonStatus.String),
			TeachingMethod:   lesson_domain.LessonTeachingMethod(l.TeachingMethod.String),
			LessonReportID:   l.LessonReportID.String,
			IsLocked:         l.IsLocked.Bool,
		})
	}
	return res
}
