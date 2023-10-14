package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func NewVirtualLessonFromEntity(l *domain.VirtualLesson) (*VirtualLesson, error) {
	dto := &VirtualLesson{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.LessonID.Set(l.LessonID),
		dto.Name.Set(l.Name),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
		dto.StartTime.Set(l.StartTime),
		dto.EndTime.Set(l.EndTime),
		dto.CenterID.Set(l.CenterID),
		dto.SchedulingStatus.Set(l.SchedulingStatus),
		dto.TeachingMedium.Set(l.TeachingMedium),
		dto.TeachingMethod.Set(l.TeachingMethod),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
	}

	return dto, nil
}

type CourseIDs struct {
	CourseIDs pgtype.TextArray
}

type LearnerIDs struct {
	LearnerIDs pgtype.TextArray
}

type TeacherIDs struct {
	TeacherIDs pgtype.TextArray
}
type VirtualLesson struct {
	LessonID             pgtype.Text
	Name                 pgtype.Text
	TeacherID            pgtype.Text // Deprecated
	CourseID             pgtype.Text // Deprecated
	ControlSettings      pgtype.JSONB
	EndAt                pgtype.Timestamptz // Deprecated
	StartTime            pgtype.Timestamptz
	EndTime              pgtype.Timestamptz
	LessonGroupID        pgtype.Text
	RoomID               pgtype.Text
	StreamLearnerCounter pgtype.Int4
	RoomState            pgtype.JSONB
	ClassID              pgtype.Text
	CenterID             pgtype.Text
	TeachingMedium       pgtype.Text // old field is LessonType
	TeachingMethod       pgtype.Text // old field is TeachingModel
	SchedulingStatus     pgtype.Text // old field is Status
	SchedulerID          pgtype.Text
	ZoomLink             pgtype.Text
	LessonCapacity       pgtype.Int4
	ClassDoOwnerID       pgtype.Text
	ClassDoLink          pgtype.Text

	// default timestamps
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz

	LearnerIDs
	TeacherIDs
	CourseIDs
}

func (l *VirtualLesson) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_id",
			"teacher_id",
			"course_id",
			"control_settings",
			"created_at",
			"updated_at",
			"deleted_at",
			"end_at",
			"lesson_group_id",
			"room_id",
			"stream_learner_counter",
			"name",
			"start_time",
			"end_time",
			"room_state",
			"class_id",
			"center_id",
			"teaching_medium",
			"teaching_method",
			"scheduling_status",
			"scheduler_id",
			"zoom_link",
			"lesson_capacity",
			"classdo_owner_id",
			"classdo_link",
		}, []interface{}{
			&l.LessonID,
			&l.TeacherID,
			&l.CourseID,
			&l.ControlSettings,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
			&l.EndAt,
			&l.LessonGroupID,
			&l.RoomID,
			&l.StreamLearnerCounter,
			&l.Name,
			&l.StartTime,
			&l.EndTime,
			&l.RoomState,
			&l.ClassID,
			&l.CenterID,
			&l.TeachingMedium,
			&l.TeachingMethod,
			&l.SchedulingStatus,
			&l.SchedulerID,
			&l.ZoomLink,
			&l.LessonCapacity,
			&l.ClassDoOwnerID,
			&l.ClassDoLink,
		}
}

func (l *VirtualLesson) TableName() string {
	return "lessons"
}

func (l *VirtualLesson) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}

func (l *VirtualLesson) PreUpdate() error {
	now := time.Now()
	if err := multierr.Combine(
		l.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}

func (l *VirtualLesson) Normalize() error {
	// default CourseID by first element of CourseIDs
	if l.CourseID.Status != pgtype.Present &&
		l.CourseIDs.CourseIDs.Status == pgtype.Present &&
		len(l.CourseIDs.CourseIDs.Elements) != 0 {
		l.CourseID = l.CourseIDs.CourseIDs.Elements[0]
	} else if len(l.CourseIDs.CourseIDs.Elements) == 0 {
		err := l.CourseIDs.CourseIDs.Set([]pgtype.Text{l.CourseID})
		if err != nil {
			return err
		}
	}

	// default TeacherID by first element of TeacherIDs
	if l.TeacherID.Status != pgtype.Present &&
		l.TeacherIDs.TeacherIDs.Status == pgtype.Present &&
		len(l.TeacherIDs.TeacherIDs.Elements) != 0 {
		l.TeacherID = l.TeacherIDs.TeacherIDs.Elements[0]
	} else if len(l.TeacherIDs.TeacherIDs.Elements) == 0 {
		err := l.TeacherIDs.TeacherIDs.Set([]pgtype.Text{l.TeacherID})
		if err != nil {
			return err
		}
	}

	if l.StreamLearnerCounter.Status != pgtype.Present {
		err := l.StreamLearnerCounter.Set(0)
		if err != nil {
			return err
		}
	}

	teacherIDs := database.FromTextArray(l.TeacherIDs.TeacherIDs)
	teacherIDs = golibs.GetUniqueElementStringArray(teacherIDs)
	if err := l.TeacherIDs.TeacherIDs.Set(teacherIDs); err != nil {
		return err
	}

	courseIDs := database.FromTextArray(l.CourseIDs.CourseIDs)
	courseIDs = golibs.GetUniqueElementStringArray(courseIDs)
	if err := l.CourseIDs.CourseIDs.Set(courseIDs); err != nil {
		return err
	}

	learnerIDs := database.FromTextArray(l.LearnerIDs.LearnerIDs)
	learnerIDs = golibs.GetUniqueElementStringArray(learnerIDs)
	if err := l.LearnerIDs.LearnerIDs.Set(learnerIDs); err != nil {
		return err
	}

	return nil
}

type ListVirtualLessonParamsDTO struct {
	StudentIDs             pgtype.TextArray
	CourseIDs              pgtype.TextArray
	LocationIDs            pgtype.TextArray
	StartDate              pgtype.Timestamptz
	EndDate                pgtype.Timestamptz
	LessonSchedulingStatus pgtype.TextArray

	Limit int32
	Page  int32
}

func ToListVirtualLessonParamsDTO(p *payloads.GetVirtualLessonsArgs) *ListVirtualLessonParamsDTO {
	args := &ListVirtualLessonParamsDTO{
		StudentIDs:             pgtype.TextArray{Status: pgtype.Null},
		CourseIDs:              pgtype.TextArray{Status: pgtype.Null},
		LocationIDs:            pgtype.TextArray{Status: pgtype.Null},
		StartDate:              pgtype.Timestamptz{Status: pgtype.Null},
		EndDate:                pgtype.Timestamptz{Status: pgtype.Null},
		LessonSchedulingStatus: pgtype.TextArray{Status: pgtype.Null},
		Limit:                  p.Limit,
		Page:                   p.Page,
	}

	if len(p.StudentIDs) > 0 {
		args.StudentIDs = database.TextArray(p.StudentIDs)
	}

	if len(p.CourseIDs) > 0 {
		args.CourseIDs = database.TextArray(p.CourseIDs)
	}

	if len(p.LocationIDs) > 0 {
		args.LocationIDs = database.TextArray(p.LocationIDs)
	}

	if !p.StartDate.IsZero() {
		args.StartDate = database.Timestamptz(p.StartDate)
	}

	if !p.EndDate.IsZero() {
		args.EndDate = database.Timestamptz(p.EndDate)
	}

	if len(p.LessonSchedulingStatuses) > 0 {
		status := []string{}
		for _, v := range p.LessonSchedulingStatuses {
			status = append(status, string(v))
		}
		args.LessonSchedulingStatus = database.TextArray(status)
	}

	return args
}

type GetLessonsQueryBuild struct {
	Distinct        string
	JoinQuery       string
	WhereQuery      string
	MainSelectQuery string
	PreviousQuery   string

	QueryArgs   []interface{}
	ParamsCount int

	OffsetLessonID pgtype.Text
}

func NewGetLessonsQueryBuild(p payloads.GetLessonsArgs) (build GetLessonsQueryBuild) {
	build.WhereQuery = ` WHERE l.deleted_at IS NULL 
		AND l.resource_path = $1 `

	build.QueryArgs = append(build.QueryArgs,
		p.SchoolID,
		database.Timestamptz(p.CurrentTime),
	)
	build.ParamsCount = len(build.QueryArgs)

	switch p.TimeLookup {
	case payloads.TimeLookupStartTime:
		build.WhereQuery += fmt.Sprintf(` AND l.start_time %s $2::timestamptz `, string(p.LessonTimeCompare))
	case payloads.TimeLookupEndTime:
		build.WhereQuery += fmt.Sprintf(` AND l.end_time %s $2::timestamptz `, string(p.LessonTimeCompare))
	case payloads.TimeLookupEndTimeIncludeWithoutEndAt:
		build.WhereQuery += fmt.Sprintf(` AND ( l.end_time %s $2::timestamptz OR end_at IS NULL ) `, string(p.LessonTimeCompare))
	case payloads.TimeLookupEndTimeIncludeWithEndAt:
		build.WhereQuery += fmt.Sprintf(` AND ( l.end_time %s $2::timestamptz OR end_at IS NOT NULL ) `, string(p.LessonTimeCompare))
	}

	switch p.LiveLessonStatus {
	case payloads.LiveLessonStatusNone:
		build.WhereQuery += " "
	case payloads.LiveLessonStatusNotEnded:
		build.WhereQuery += " AND end_at IS NULL "
	case payloads.LiveLessonStatusEnded:
		build.WhereQuery += " AND end_at IS NOT NULL "
	}

	if len(p.LocationIDs) > 0 {
		build.ParamsCount++
		build.WhereQuery += fmt.Sprintf(` AND l.center_id = ANY($%d) `, build.ParamsCount)
		build.QueryArgs = append(build.QueryArgs, database.TextArray(p.LocationIDs))
	}

	if len(p.TeacherIDs) > 0 {
		build.ParamsCount++
		build.JoinQuery += ` JOIN lessons_teachers lt ON l.lesson_id = lt.lesson_id `
		build.WhereQuery += fmt.Sprintf(` AND lt.teacher_id = ANY($%d) and lt.deleted_at IS NULL `, build.ParamsCount)
		build.QueryArgs = append(build.QueryArgs, database.TextArray(p.TeacherIDs))
	}

	if len(p.StudentIDs) > 0 {
		build.ParamsCount++
		build.Distinct = DistinctKeyword
		build.JoinQuery += ` JOIN lesson_members lm ON l.lesson_id = lm.lesson_id `
		build.WhereQuery += fmt.Sprintf(`AND lm.user_id = ANY($%d) AND lm.deleted_at IS NULL `, build.ParamsCount)
		build.QueryArgs = append(build.QueryArgs, database.TextArray(p.StudentIDs))
	}

	if len(p.CourseIDs) > 0 {
		build.ParamsCount++
		build.Distinct = DistinctKeyword
		if len(p.StudentIDs) == 0 {
			build.JoinQuery += ` left join lesson_members lm on l.lesson_id = lm.lesson_id `
			build.WhereQuery += " AND lm.deleted_at IS NULL "
		}
		build.WhereQuery += fmt.Sprintf(` AND ( lm.course_id = ANY($%d) OR l.course_id = ANY($%d) ) `, build.ParamsCount, build.ParamsCount)
		build.QueryArgs = append(build.QueryArgs, database.TextArray(p.CourseIDs))
	}

	if len(p.LessonSchedulingStatuses) > 0 {
		status := make([]string, 0, len(p.LessonSchedulingStatuses))
		for _, v := range p.LessonSchedulingStatuses {
			status = append(status, string(v))
		}
		build.ParamsCount++
		build.WhereQuery += fmt.Sprintf(` AND l.scheduling_status = ANY($%d) `, build.ParamsCount)
		build.QueryArgs = append(build.QueryArgs, database.TextArray(status))
	}

	if !p.FromDate.IsZero() {
		build.ParamsCount++
		build.WhereQuery += fmt.Sprintf(` AND l.end_time >= $%d::timestamptz `, build.ParamsCount)
		build.QueryArgs = append(build.QueryArgs, database.Timestamptz(p.FromDate))
	}

	if !p.ToDate.IsZero() {
		build.ParamsCount++
		build.WhereQuery += fmt.Sprintf(` AND l.start_time <= $%d::timestamptz `, build.ParamsCount)
		build.QueryArgs = append(build.QueryArgs, database.Timestamptz(p.ToDate))
	}

	if p.SortAscending {
		build.MainSelectQuery = getLessonQueryAscending
		build.PreviousQuery = previousLessonQueryAscending
	} else {
		build.MainSelectQuery = getLessonQueryDescending
		build.PreviousQuery = previousLessonQueryDescending
	}

	build.OffsetLessonID = pgtype.Text{Status: pgtype.Null}
	if len(p.OffsetLessonID) > 0 {
		build.OffsetLessonID = database.Text(p.OffsetLessonID)
	}

	return build
}
