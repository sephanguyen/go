package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lrd "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
)

type (
	LessonStatus             string
	LessonType               string
	LessonTeachingModel      string
	LessonReportStatusFilter string
)

const (
	LessonStatusDraft LessonStatus = "LESSON_STATUS_DRAFT"

	LessonTypeOnline  LessonType = "LESSON_TYPE_ONLINE"
	LessonTypeOffline LessonType = "LESSON_TYPE_OFFLINE"

	LessonTeachingModelIndividual LessonTeachingModel = "LESSON_TEACHING_MODEL_INDIVIDUAL"
	LessonTeachingModelGroup      LessonTeachingModel = "LESSON_TEACHING_MODEL_GROUP"

	ReportStatusSubmitted LessonReportStatusFilter = "LESSON_REPORT_SUBMITTING_STATUS_SUBMITTED"
	ReportStatusDraft     LessonReportStatusFilter = "LESSON_REPORT_SUBMITTING_STATUS_SAVED"
	ReportStatusNone      LessonReportStatusFilter = "LESSON_REPORT_SUBMITTING_STATUS_NONE"
)

var (
	migrationLessonStatus = map[domain.LessonSchedulingStatus]LessonStatus{
		domain.LessonSchedulingStatusDraft:     LessonStatusDraft,
		domain.LessonSchedulingStatusPublished: LessonStatusDraft,
	}
	migrationLessonType = map[domain.LessonTeachingMedium]LessonType{
		domain.LessonTeachingMediumOffline: LessonTypeOffline,
		domain.LessonTeachingMediumOnline:  LessonTypeOnline,
		domain.LessonTeachingMediumZoom:    LessonTypeOnline,
		domain.LessonTeachingMediumClassDo: LessonTypeOnline,
	}
	migrationLessonTeachingModel = map[domain.LessonTeachingMethod]LessonTeachingModel{
		domain.LessonTeachingMethodIndividual: LessonTeachingModelIndividual,
		domain.LessonTeachingMethodGroup:      LessonTeachingModelGroup,
	}
	ReportStatusMapping = map[lrd.LessonReportStatus]LessonReportStatusFilter{
		lrd.ReportStatusDraft:     ReportStatusDraft,
		lrd.ReportStatusSubmitted: ReportStatusSubmitted,
		lrd.ReportStatusNone:      ReportStatusNone,
	}
)

func NewLessonFromEntity(l *domain.Lesson) (*Lesson, error) {
	dto := &Lesson{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.LessonID.Set(l.LessonID),
		dto.Name.Set(l.Name),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
		dto.StartTime.Set(l.StartTime),
		dto.EndTime.Set(l.EndTime),
		dto.CenterID.Set(l.LocationID),
		dto.SchedulingStatus.Set(l.SchedulingStatus),
		dto.TeachingMedium.Set(l.TeachingMedium),
		dto.TeachingMethod.Set(l.TeachingMethod),
		dto.IsLocked.Set(l.IsLocked),
		dto.ZoomLink.Set(l.ZoomLink),
		dto.ZoomOwnerID.Set(l.ZoomOwnerID),
		dto.ZoomID.Set(l.ZoomID),
		dto.ZoomOccurrenceID.Set(l.ZoomOccurrenceID),
		dto.ClassDoOwnerID.Set(l.ClassDoOwnerID),
		dto.ClassDoLink.Set(l.ClassDoLink),
		dto.ClassDoRoomID.Set(l.ClassDoRoomID),
		dto.LessonCapacity.Set(l.LessonCapacity),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
	}
	// if teaching_method == "group", don't set course_id
	// since if teaching_method == "group", the request must already contains class_id and course_id
	if dto.TeachingMethod.String == string(domain.LessonTeachingMethodIndividual) && len(l.Learners) > 0 {
		if err := dto.CourseID.Set(l.Learners[0].CourseID); err != nil {
			return nil, fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
		}
	}
	if dto.TeachingMethod.String == string(domain.LessonTeachingMethodGroup) {
		var err error
		if len(l.CourseID) > 0 {
			multierr.Append(err, dto.CourseID.Set(l.CourseID))
		}
		if len(l.ClassID) > 0 {
			multierr.Append(err, dto.ClassID.Set(l.ClassID))
		}
		if err != nil {
			return nil, fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
		}
	}
	if len(l.Teachers) > 0 {
		if err := dto.TeacherID.Set(l.Teachers[0].TeacherID); err != nil {
			return nil, fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
		}
	}
	if len(l.SchedulerID) > 0 {
		if err := dto.SchedulerID.Set(l.SchedulerID); err != nil {
			return nil, fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
		}
	}
	if l.PreparationTime != -1 {
		if err := dto.PreparationTime.Set(l.PreparationTime); err != nil {
			return nil, fmt.Errorf("could not mapping preparation_time from lesson entity to lesson dto: %w", err)
		}
	}
	if l.BreakTime != -1 {
		if err := dto.BreakTime.Set(l.BreakTime); err != nil {
			return nil, fmt.Errorf("could not mapping break_time from lesson entity to lesson dto: %w", err)
		}
	}
	return dto, nil
}

type Lesson struct {
	LessonID             pgtype.Text
	Name                 pgtype.Text
	TeacherID            pgtype.Text // Deprecated
	CourseID             pgtype.Text // Deprecated
	ControlSettings      pgtype.JSONB
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
	EndAt                pgtype.Timestamptz
	StartTime            pgtype.Timestamptz
	EndTime              pgtype.Timestamptz
	LessonGroupID        pgtype.Text
	RoomID               pgtype.Text
	LessonType           pgtype.Text // Deprecated
	Status               pgtype.Text // Deprecated
	StreamLearnerCounter pgtype.Int4
	LearnerIds           pgtype.TextArray // Deprecated
	RoomState            pgtype.JSONB
	TeachingModel        pgtype.Text // Deprecated
	ClassID              pgtype.Text
	CenterID             pgtype.Text
	TeachingMedium       pgtype.Text // old field is LessonType
	TeachingMethod       pgtype.Text // old field is TeachingModel
	SchedulingStatus     pgtype.Text // old field is Status
	SchedulerID          pgtype.Text
	IsLocked             pgtype.Bool
	ZoomLink             pgtype.Text
	ZoomOwnerID          pgtype.Text
	ZoomID               pgtype.Text
	ZoomOccurrenceID     pgtype.Text
	ClassDoOwnerID       pgtype.Text
	ClassDoLink          pgtype.Text
	ClassDoRoomID        pgtype.Text
	LessonCapacity       pgtype.Int4
	PreparationTime      pgtype.Int4
	BreakTime            pgtype.Int4
}

func (l *Lesson) FieldMap() ([]string, []interface{}) {
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
			"lesson_type",
			"status",
			"stream_learner_counter",
			"learner_ids",
			"name",
			"start_time",
			"end_time",
			"room_state",
			"teaching_model",
			"class_id",
			"center_id",
			"teaching_medium",
			"teaching_method",
			"scheduling_status",
			"scheduler_id",
			"is_locked",
			"zoom_link",
			"zoom_owner_id",
			"zoom_id",
			"zoom_occurrence_id",
			"lesson_capacity",
			"preparation_time",
			"break_time",
			"classdo_owner_id",
			"classdo_link",
			"classdo_room_id",
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
			&l.LessonType,
			&l.Status,
			&l.StreamLearnerCounter,
			&l.LearnerIds,
			&l.Name,
			&l.StartTime,
			&l.EndTime,
			&l.RoomState,
			&l.TeachingModel,
			&l.ClassID,
			&l.CenterID,
			&l.TeachingMedium,
			&l.TeachingMethod,
			&l.SchedulingStatus,
			&l.SchedulerID,
			&l.IsLocked,
			&l.ZoomLink,
			&l.ZoomOwnerID,
			&l.ZoomID,
			&l.ZoomOccurrenceID,
			&l.LessonCapacity,
			&l.PreparationTime,
			&l.BreakTime,
			&l.ClassDoOwnerID,
			&l.ClassDoLink,
			&l.ClassDoRoomID,
		}
}

func (l *Lesson) TableName() string {
	return "lessons"
}

func (l *Lesson) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}

func (l *Lesson) PreUpdate() error {
	now := time.Now()
	if err := multierr.Combine(
		l.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}

func (l *Lesson) PreUpdateSchedulingID(lessonID, schedulerID string) error {
	database.AllNullEntity(l)
	err := multierr.Combine(
		l.LessonID.Set(lessonID),
		l.SchedulerID.Set(schedulerID),
	)
	if err != nil {
		return fmt.Errorf("could not mapping from lessonId and schedulerId to lesson dto: %w", err)
	}
	return nil
}

func (l *Lesson) Normalize() error {
	if l.Status.Status != pgtype.Present {
		if err := l.Status.Set(
			migrationLessonStatus[domain.LessonSchedulingStatus(l.SchedulingStatus.String)],
		); err != nil {
			return err
		}
	}

	if l.LessonType.Status != pgtype.Present {
		err := l.LessonType.Set(
			migrationLessonType[domain.LessonTeachingMedium(l.TeachingMedium.String)],
		)
		if err != nil {
			return err
		}
	}

	if l.TeachingModel.Status != pgtype.Present {
		err := l.TeachingModel.Set(
			migrationLessonTeachingModel[domain.LessonTeachingMethod(l.TeachingMethod.String)],
		)
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

	if l.LearnerIds.Status != pgtype.Present {
		err := l.LearnerIds.Set([]string{})
		if err != nil {
			return err
		}
	}

	if l.IsLocked.Status != pgtype.Present {
		if err := l.IsLocked.Set(false); err != nil {
			return err
		}
	}

	return nil
}

func (u *Lesson) PreLock(lessonId string) error {
	database.AllNullEntity(u)

	err := multierr.Combine(
		u.LessonID.Set(lessonId),
		u.IsLocked.Set(true),
	)
	if err != nil {
		return fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
	}
	return nil
}

type Lessons []*Lesson

func (u *Lessons) Add() database.Entity {
	e := &Lesson{}
	*u = append(*u, e)

	return e
}

type ListLessonArgs struct {
	Limit                  uint32
	SchoolID               string
	LessonID               pgtype.Text
	CurrentTime            pgtype.Timestamptz
	LessonTime             string
	Compare                string
	Courses                pgtype.TextArray
	Teachers               pgtype.TextArray
	Students               pgtype.TextArray
	FromDate               pgtype.Timestamptz
	ToDate                 pgtype.Timestamptz
	KeyWord                pgtype.Text
	TimeZone               string
	Dow                    []int32
	LocationIDs            pgtype.TextArray
	ClassIDs               pgtype.TextArray
	Grades                 []int32
	FromTime               string
	ToTime                 string
	LessonSchedulingStatus pgtype.TextArray
	GradesV2               pgtype.TextArray
	CourseTypeIDs          pgtype.TextArray
	ReportStatus           pgtype.TextArray
}

func ToListLessonArgsDto(p *payloads.GetLessonListArg) *ListLessonArgs {
	args := &ListLessonArgs{
		Limit:                  p.Limit,
		SchoolID:               p.SchoolID,
		LessonTime:             p.LessonTime,
		Compare:                p.Compare,
		FromTime:               p.FromTime,
		ToTime:                 p.ToTime,
		TimeZone:               p.TimeZone,
		Grades:                 p.Grades,
		GradesV2:               pgtype.TextArray{Status: pgtype.Null},
		LessonID:               pgtype.Text{Status: pgtype.Null},
		CurrentTime:            pgtype.Timestamptz{Status: pgtype.Null},
		Courses:                pgtype.TextArray{Status: pgtype.Null},
		Teachers:               pgtype.TextArray{Status: pgtype.Null},
		Students:               pgtype.TextArray{Status: pgtype.Null},
		FromDate:               pgtype.Timestamptz{Status: pgtype.Null},
		ToDate:                 pgtype.Timestamptz{Status: pgtype.Null},
		KeyWord:                pgtype.Text{Status: pgtype.Null},
		LocationIDs:            pgtype.TextArray{Status: pgtype.Null},
		ClassIDs:               pgtype.TextArray{Status: pgtype.Null},
		LessonSchedulingStatus: pgtype.TextArray{Status: pgtype.Null},
		CourseTypeIDs:          pgtype.TextArray{Status: pgtype.Null},
		ReportStatus:           pgtype.TextArray{Status: pgtype.Null},
	}

	if len(p.CourseIDs) > 0 {
		args.Courses = database.TextArray(p.CourseIDs)
	}

	if len(p.TeacherIDs) > 0 {
		args.Teachers = database.TextArray(p.TeacherIDs)
	}

	if len(p.StudentIDs) > 0 {
		args.Students = database.TextArray(p.StudentIDs)
	}

	if !p.FromDate.IsZero() {
		args.FromDate = database.Timestamptz(p.FromDate)
	}

	if !p.ToDate.IsZero() {
		args.ToDate = database.Timestamptz(p.ToDate)
	}

	if !p.CurrentTime.IsZero() {
		args.CurrentTime = database.Timestamptz(p.CurrentTime)
	}

	if p.KeyWord != "" {
		args.KeyWord = database.Text(p.KeyWord)
	}

	if p.LessonID != "" {
		args.LessonID = database.Text(p.LessonID)
	}

	if len(p.LocationIDs) > 0 {
		args.LocationIDs = database.TextArray(p.LocationIDs)
	}

	if len(p.ClassIDs) > 0 {
		args.ClassIDs = database.TextArray(p.ClassIDs)
	}
	if len(p.GradesV2) > 0 {
		args.GradesV2 = database.TextArray(p.GradesV2)
	}

	if len(p.LessonSchedulingStatuses) > 0 {
		status := []string{}
		for _, v := range p.LessonSchedulingStatuses {
			status = append(status, string(v))
		}

		args.LessonSchedulingStatus = database.TextArray(status)
	}

	if len(p.Dow) > 0 {
		dow := make([]int32, 0, len(p.Dow))
		for _, v := range p.Dow {
			dow = append(dow, int32(v))
		}
		args.Dow = dow
	}
	if len(p.CourseTypesIDs) > 0 {
		args.CourseTypeIDs = database.TextArray(p.CourseTypesIDs)
	}
	if len(p.LessonReportStatus) > 0 {
		for _, rs := range p.LessonReportStatus {
			args.ReportStatus = database.AppendText(args.ReportStatus, database.Text(string(ReportStatusMapping[rs])))
		}
	}
	return args
}

func (l *ListLessonArgs) GetParamGrades() interface{} {
	if l.isExistGradesV2() {
		return &l.GradesV2
	}
	return &l.Grades
}

func (l *ListLessonArgs) isExistGradesV2() bool {
	return len(l.GradesV2.Elements) > 0
}

func (l *ListLessonArgs) ExistsConditionGrades() bool {
	isExistGradesV2 := l.isExistGradesV2()
	if isExistGradesV2 {
		return isExistGradesV2
	}
	return len(l.Grades) > 0
}

func (l *ListLessonArgs) IsPresentReportStatus() bool {
	return l.ReportStatus.Status == pgtype.Present
}

func (l *ListLessonArgs) IsHaveNoneReportStatus() bool {
	reportStatus := database.FromTextArray(l.ReportStatus)
	return slices.Contains(reportStatus, string(ReportStatusNone))
}

type ListLessonOnCalendarArgs struct {
	LocationID                          pgtype.Text
	FromDate                            pgtype.Timestamptz
	ToDate                              pgtype.Timestamptz
	Timezone                            string
	StudentIDs                          pgtype.TextArray
	CourseIDs                           pgtype.TextArray
	TeacherIDs                          pgtype.TextArray
	ClassIDs                            pgtype.TextArray
	IsIncludeNoneAssignedTeacherLessons pgtype.Bool
}

func ToListLessonOnCalendarArgsDto(p *payloads.GetLessonListOnCalendarArgs) (*ListLessonOnCalendarArgs, error) {
	argsDTO := &ListLessonOnCalendarArgs{
		LocationID:                          pgtype.Text{Status: pgtype.Null},
		FromDate:                            pgtype.Timestamptz{Status: pgtype.Null},
		ToDate:                              pgtype.Timestamptz{Status: pgtype.Null},
		Timezone:                            p.Timezone,
		StudentIDs:                          pgtype.TextArray{Status: pgtype.Null},
		CourseIDs:                           pgtype.TextArray{Status: pgtype.Null},
		TeacherIDs:                          pgtype.TextArray{Status: pgtype.Null},
		ClassIDs:                            pgtype.TextArray{Status: pgtype.Null},
		IsIncludeNoneAssignedTeacherLessons: pgtype.Bool{Status: pgtype.Null},
	}

	if err := multierr.Combine(
		argsDTO.LocationID.Set(p.LocationID),
		argsDTO.FromDate.Set(p.FromDate),
		argsDTO.ToDate.Set(p.ToDate),
		argsDTO.IsIncludeNoneAssignedTeacherLessons.Set(p.IsIncludeNoneAssignedTeacherLessons),
	); err != nil {
		return nil, fmt.Errorf("could not map list lesson on calendar args dto: %w", err)
	}

	if len(p.StudentIDs) > 0 {
		if err := argsDTO.StudentIDs.Set(p.StudentIDs); err != nil {
			return nil, fmt.Errorf("could not set StudentIDs on list lesson on calendar args dto: %w", err)
		}
	}

	if len(p.CourseIDs) > 0 {
		if err := argsDTO.CourseIDs.Set(p.CourseIDs); err != nil {
			return nil, fmt.Errorf("could not set CourseIDs on list lesson on calendar args dto: %w", err)
		}
	}

	if len(p.TeacherIDs) > 0 {
		if err := argsDTO.TeacherIDs.Set(p.TeacherIDs); err != nil {
			return nil, fmt.Errorf("could not set TeacherIDs on list lesson on calendar args dto: %w", err)
		}
	}

	if len(p.ClassIDs) > 0 {
		if err := argsDTO.ClassIDs.Set(p.ClassIDs); err != nil {
			return nil, fmt.Errorf("could not set ClassIDs on list lesson on calendar args dto: %w", err)
		}
	}

	return argsDTO, nil
}

type LessonToExport struct {
	StartTime         pgtype.Timestamptz
	EndTime           pgtype.Timestamptz
	TeachingMethod    pgtype.Text // 1 is individual - 2 is group
	PartnerInternalID pgtype.Text
	TeachingMedium    pgtype.Text // 1 is offline - 2 is group
	TeacherIDs        pgtype.Text
	StudentCourseIDs  pgtype.Text // studentID/CourseID_studentID/CourseID
}

func (l *LessonToExport) FieldMap() ([]string, []interface{}) {
	return []string{
			"start_time",
			"end_time",
			"teaching_method",
			"partner_internal_id",
			"teaching_medium",
			"teacher_ids",
			"student_course_ids",
		}, []interface{}{
			&l.StartTime,
			&l.EndTime,
			&l.TeachingMethod,
			&l.PartnerInternalID,
			&l.TeachingMedium,
			&l.TeacherIDs,
			&l.StudentCourseIDs,
		}
}

func (l *LessonToExport) TableName() string {
	return "lessons"
}

func InitLessonTemplate() *LessonToExport {
	l := LessonToExport{}
	now := time.Now()
	l.StartTime = database.Timestamptz(now)
	l.EndTime = database.Timestamptz(now.Add(2 * time.Hour))
	l.TeachingMethod = database.Text(domain.MapKeyLessonTeachingMethod[domain.LessonTeachingMethodIndividual])
	l.PartnerInternalID = database.Text("sample_center_id")
	return &l
}

func InitLessonTemplateV2() *LessonToExport {
	l := LessonToExport{}
	now := time.Now()
	l.StartTime = database.Timestamptz(now)
	l.EndTime = database.Timestamptz(now.Add(2 * time.Hour))
	l.TeachingMethod = database.Text(domain.MapKeyLessonTeachingMethod[domain.LessonTeachingMethodIndividual])
	l.PartnerInternalID = database.Text("sample_center_id")
	l.TeachingMedium = database.Text(domain.MapKeyLessonTeachingMedium[domain.LessonTeachingMediumOffline])
	l.TeacherIDs = database.Text("teacherID1_teacherID2_teacherID3")
	l.StudentCourseIDs = database.Text("studentID1/courseID1_studentID2/courseID2")
	return &l
}

type ListLessonByLocationStatusAndDateTimeRangeArgs struct {
	LocationID   pgtype.Text
	LessonStatus pgtype.Text
	StartDate    pgtype.Timestamptz
	EndDate      pgtype.Timestamptz
	StartTime    pgtype.Timestamptz
	EndTime      pgtype.Timestamptz
	Timezone     pgtype.Text
}

func ToListLessonByLocationStatusAndDateTimeRangeArgsDto(p *payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs) *ListLessonByLocationStatusAndDateTimeRangeArgs {
	argsDTO := &ListLessonByLocationStatusAndDateTimeRangeArgs{
		LocationID:   pgtype.Text{Status: pgtype.Null},
		LessonStatus: pgtype.Text{Status: pgtype.Null},
		StartDate:    pgtype.Timestamptz{Status: pgtype.Null},
		EndDate:      pgtype.Timestamptz{Status: pgtype.Null},
		StartTime:    pgtype.Timestamptz{Status: pgtype.Null},
		EndTime:      pgtype.Timestamptz{Status: pgtype.Null},
		Timezone:     pgtype.Text{Status: pgtype.Null},
	}

	if len(p.LocationID) > 0 {
		argsDTO.LocationID = database.Text(p.LocationID)
	}

	if len(p.LessonStatus) > 0 {
		argsDTO.LessonStatus = database.Text(string(p.LessonStatus))
	}

	if !p.StartDate.IsZero() {
		argsDTO.StartDate = database.Timestamptz(p.StartDate)
	}

	if !p.EndDate.IsZero() {
		argsDTO.EndDate = database.Timestamptz(p.EndDate)
	}

	if !p.StartTime.IsZero() {
		argsDTO.StartTime = database.Timestamptz(p.StartTime)
	}

	if !p.EndTime.IsZero() {
		argsDTO.EndTime = database.Timestamptz(p.EndTime)
	}

	if p.Timezone == "" {
		argsDTO.Timezone = database.Text("UTC")
	} else {
		argsDTO.Timezone = database.Text(p.Timezone)
	}

	return argsDTO
}
