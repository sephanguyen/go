package domain

import (
	"context"
	"fmt"
	"time"

	calendar_constants "github.com/manabie-com/backend/internal/calendar/domain/constants"
	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	utils "github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type (
	LessonSchedulingStatus  string
	LessonTeachingMedium    string
	LessonTeachingMethod    string
	StudentAttendStatus     string
	DateOfWeek              int32
	StudentAttendanceNotice string
	StudentAttendanceReason string
	LessonBulkAction        string
)

const (
	LessonSchedulingStatusDraft     LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_DRAFT"
	LessonSchedulingStatusPublished LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_PUBLISHED"
	LessonSchedulingStatusCompleted LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_COMPLETED"
	LessonSchedulingStatusCanceled  LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_CANCELED"

	LessonTeachingMediumOffline LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_OFFLINE"
	LessonTeachingMediumOnline  LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_ONLINE"
	LessonTeachingMediumZoom    LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_ZOOM"
	LessonTeachingMediumClassDo LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_CLASS_DO"

	LessonTeachingMethodIndividual LessonTeachingMethod = "LESSON_TEACHING_METHOD_INDIVIDUAL"
	LessonTeachingMethodGroup      LessonTeachingMethod = "LESSON_TEACHING_METHOD_GROUP"

	StudentAttendStatusEmpty          StudentAttendStatus = "STUDENT_ATTEND_STATUS_EMPTY"
	StudentAttendStatusAttend         StudentAttendStatus = "STUDENT_ATTEND_STATUS_ATTEND"
	StudentAttendStatusAbsent         StudentAttendStatus = "STUDENT_ATTEND_STATUS_ABSENT"
	StudentAttendStatusLate           StudentAttendStatus = "STUDENT_ATTEND_STATUS_LATE"
	StudentAttendStatusInformedAbsent StudentAttendStatus = "STUDENT_ATTEND_STATUS_INFORMED_ABSENT"
	StudentAttendStatusInformedLate   StudentAttendStatus = "STUDENT_ATTEND_STATUS_INFORMED_LATE"
	StudentAttendStatusLeaveEarly     StudentAttendStatus = "STUDENT_ATTEND_STATUS_LEAVE_EARLY"
	StudentAttendStatusReallocate     StudentAttendStatus = "STUDENT_ATTEND_STATUS_REALLOCATE"

	NoticeEmpty StudentAttendanceNotice = "NOTICE_EMPTY"
	InAdvance   StudentAttendanceNotice = "IN_ADVANCE"
	OnTheDay    StudentAttendanceNotice = "ON_THE_DAY"
	NoContact   StudentAttendanceNotice = "NO_CONTACT"

	ReasonEmpty       StudentAttendanceReason = "REASON_EMPTY"
	PhysicalCondition StudentAttendanceReason = "PHYSICAL_CONDITION"
	SchoolEvent       StudentAttendanceReason = "SCHOOL_EVENT"
	FamilyReason      StudentAttendanceReason = "FAMILY_REASON"
	ReasonOther       StudentAttendanceReason = "REASON_OTHER"

	Monday    DateOfWeek = 0
	Tuesday   DateOfWeek = 1
	Wednesday DateOfWeek = 2
	Thurday   DateOfWeek = 3
	Friday    DateOfWeek = 4
	Saturday  DateOfWeek = 5
	Sunday    DateOfWeek = 6

	LessonBulkActionCancel  LessonBulkAction = "LESSON_BULK_ACTION_CANCEL"
	LessonBulkActionPublish LessonBulkAction = "LESSON_BULK_ACTION_PUBLISH"

	YYYMMDD string = "2006-01-02"
)

var MapKeyLessonTeachingMethod = map[LessonTeachingMethod]string{
	LessonTeachingMethodIndividual: "1",
	LessonTeachingMethodGroup:      "2",
}

var MapValueLessonTeachingMethod = map[string]LessonTeachingMethod{
	"1": LessonTeachingMethodIndividual,
	"2": LessonTeachingMethodGroup,
}

var MapKeyLessonTeachingMedium = map[LessonTeachingMedium]string{
	LessonTeachingMediumOffline: "1",
	LessonTeachingMediumOnline:  "2",
}

var MapValueLessonTeachingMedium = map[string]LessonTeachingMedium{
	"1": LessonTeachingMediumOffline,
	"2": LessonTeachingMediumOnline,
}

type LessonBuilder struct {
	lesson *Lesson
}

func NewLesson() *LessonBuilder {
	return &LessonBuilder{
		lesson: &Lesson{},
	}
}

// BuildDraft will return a lesson but not valid data
func (l *LessonBuilder) BuildDraft() *Lesson {
	return l.lesson
}

func (l *LessonBuilder) Build(ctx context.Context, db database.Ext) (*Lesson, error) {
	if err := l.lesson.IsValid(ctx, db); err != nil {
		return nil, fmt.Errorf("invalid lesson: %w", err)
	}
	return l.lesson, nil
}

func (l *LessonBuilder) WithLocationID(locationID string) *LessonBuilder {
	l.lesson.LocationID = locationID
	return l
}

func (l *LessonBuilder) WithCourseID(courseID string) *LessonBuilder {
	l.lesson.CourseID = courseID
	return l
}

func (l *LessonBuilder) WithClassID(classID string) *LessonBuilder {
	l.lesson.ClassID = classID
	return l
}

func (l *LessonBuilder) WithID(id string) *LessonBuilder {
	l.lesson.LessonID = id
	l.lesson.Persisted = true
	return l
}

func (l *LessonBuilder) WithName(name string) *LessonBuilder {
	l.lesson.Name = name
	return l
}

func (l *LessonBuilder) WithTimeRange(startTime, endTime time.Time) *LessonBuilder {
	l.lesson.StartTime = startTime
	l.lesson.EndTime = endTime
	return l
}

func (l *LessonBuilder) WithModificationTime(createdAt, updatedAt time.Time) *LessonBuilder {
	l.lesson.CreatedAt = createdAt
	l.lesson.UpdatedAt = updatedAt
	return l
}

func (l *LessonBuilder) WithDeletedTime(deletedAt *time.Time) *LessonBuilder {
	l.lesson.DeletedAt = deletedAt
	return l
}

func (l *LessonBuilder) WithMaterials(mediaIDs []string) *LessonBuilder {
	if len(mediaIDs) == 0 {
		return l
	}

	mediaIDs = utils.GetUniqueElementStringArray(mediaIDs)
	if l.lesson.Material != nil {
		l.lesson.Material.MediaIDs = mediaIDs
	} else {
		l.lesson.Material = &LessonMaterial{
			MediaIDs: mediaIDs,
		}
	}
	return l
}

func (l *LessonBuilder) WithTeachingMedium(medium LessonTeachingMedium) *LessonBuilder {
	l.lesson.TeachingMedium = medium
	return l
}

func (l *LessonBuilder) WithTeachingMethod(method LessonTeachingMethod) *LessonBuilder {
	l.lesson.TeachingMethod = method
	return l
}

func (l *LessonBuilder) WithTeacherIDs(teacherIDs []string) *LessonBuilder {
	teacherIDs = utils.GetUniqueElementStringArray(teacherIDs)
	l.lesson.Teachers = make(LessonTeachers, 0, len(teacherIDs))
	for _, id := range teacherIDs {
		l.lesson.Teachers = append(l.lesson.Teachers, &LessonTeacher{
			TeacherID: id,
		})
	}
	return l
}

func (l *LessonBuilder) WithSchedulingStatus(schedulingStatus LessonSchedulingStatus) *LessonBuilder {
	l.lesson.SchedulingStatus = schedulingStatus
	return l
}

func (l *LessonBuilder) WithLearners(learners LessonLearners) *LessonBuilder {
	l.lesson.Learners = learners
	return l
}

func (l *LessonBuilder) WithSchedulerID(schedulerID string) *LessonBuilder {
	l.lesson.SchedulerID = schedulerID
	return l
}

func (l *LessonBuilder) WithIsLocked(isLocked bool) *LessonBuilder {
	l.lesson.IsLocked = isLocked
	return l
}

func (l *LessonBuilder) WithCourseName(courseName string) *LessonBuilder {
	l.lesson.CourseName = courseName
	return l
}

func (l *LessonBuilder) WithClassName(className string) *LessonBuilder {
	l.lesson.ClassName = className
	return l
}

func (l *LessonBuilder) WithLocationName(locationName string) *LessonBuilder {
	l.lesson.LocationName = locationName
	return l
}

func (l *LessonBuilder) WithClassroomIDs(classroomIDs []string) *LessonBuilder {
	classroomIDs = utils.GetUniqueElementStringArray(classroomIDs)
	l.lesson.Classrooms = make(LessonClassrooms, 0, len(classroomIDs))
	for _, id := range classroomIDs {
		l.lesson.Classrooms = append(l.lesson.Classrooms, &LessonClassroom{
			ClassroomID: id,
		})
	}
	return l
}

func (l *LessonBuilder) WithZoomLink(zoomLink string) *LessonBuilder {
	l.lesson.ZoomLink = zoomLink
	return l
}

func (l *LessonBuilder) WithZoomOccurrenceID(zoomOccurrenceID string) *LessonBuilder {
	l.lesson.ZoomOccurrenceID = zoomOccurrenceID
	return l
}

func (l *LessonBuilder) WithZoomID(zoomID string) *LessonBuilder {
	l.lesson.ZoomID = zoomID
	return l
}

func (l *LessonBuilder) WithZoomAccountID(zoomAccountID string) *LessonBuilder {
	l.lesson.ZoomOwnerID = zoomAccountID
	return l
}

func (l *LessonBuilder) WithClassDoOwnerID(classDoOwnerID string) *LessonBuilder {
	l.lesson.ClassDoOwnerID = classDoOwnerID
	return l
}

func (l *LessonBuilder) WithClassDoLink(classDoLink string) *LessonBuilder {
	l.lesson.ClassDoLink = classDoLink
	return l
}

func (l *LessonBuilder) WithClassDoRoomID(classDoRoomID string) *LessonBuilder {
	l.lesson.ClassDoRoomID = classDoRoomID
	return l
}

func (l *LessonBuilder) AddLearner(ll *LessonLearner) *LessonBuilder {
	if len(l.lesson.Learners) == 0 {
		l.lesson.Learners = make(LessonLearners, 0, 1)
	}
	l.lesson.Learners = append(l.lesson.Learners, ll)
	return l
}

func (l *LessonBuilder) WithMasterDataPort(port MasterDataPort) *LessonBuilder {
	l.lesson.MasterDataPort = port
	return l
}

func (l *LessonBuilder) WithUserModulePort(port UserModulePort) *LessonBuilder {
	l.lesson.UserModulePort = port
	return l
}

func (l *LessonBuilder) WithMediaModulePort(port MediaModulePort) *LessonBuilder {
	l.lesson.MediaModulePort = port
	return l
}

func (l *LessonBuilder) WithLessonRepo(repo LessonRepo) *LessonBuilder {
	l.lesson.Repo = repo
	return l
}

func (l *LessonBuilder) WithDateInfoRepo(port DateInfoRepo) *LessonBuilder {
	l.lesson.DateInfoRepo = port
	return l
}

func (l *LessonBuilder) WithClassroomRepo(port ClassroomRepo) *LessonBuilder {
	l.lesson.ClassroomRepo = port
	return l
}

func (l *LessonBuilder) WithLessonCapacity(cap int32) *LessonBuilder {
	l.lesson.LessonCapacity = cap
	return l
}

func (l *LessonBuilder) WithSchedulerInfo(schedulerInfo *SchedulerInfo) *LessonBuilder {
	l.lesson.SchedulerInfo = schedulerInfo
	return l
}

func (l *LessonBuilder) WithEndAt(endAt *time.Time) *LessonBuilder {
	l.lesson.EndAt = endAt
	return l
}

func (l *LessonBuilder) WithPreparationTime(minutes int32) *LessonBuilder {
	l.lesson.PreparationTime = minutes
	return l
}

func (l *LessonBuilder) WithBreakTime(minutes int32) *LessonBuilder {
	l.lesson.BreakTime = minutes
	return l
}

type SchedulerInfo struct {
	SchedulerID string
	Freq        string
}

type Lesson struct {
	LessonID         string
	Name             string
	LocationID       string
	CourseID         string
	ClassID          string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	StartTime        time.Time
	EndTime          time.Time
	SchedulingStatus LessonSchedulingStatus
	TeachingMedium   LessonTeachingMedium
	TeachingMethod   LessonTeachingMethod
	Learners         LessonLearners
	Teachers         LessonTeachers
	Material         *LessonMaterial
	SchedulerID      string
	IsLocked         bool
	DateInfos        []*calendar_dto.DateInfo
	Classrooms       LessonClassrooms
	LessonCapacity   int32
	PreparationTime  int32
	BreakTime        int32

	// for lessons on calendar api
	CourseName    string
	ClassName     string
	ClassroomName string
	LocationName  string

	// internal state
	Persisted           bool // true: lesson already exists in db
	PreSchedulingStatus LessonSchedulingStatus

	// ports
	MasterDataPort  MasterDataPort
	UserModulePort  UserModulePort
	MediaModulePort MediaModulePort
	Repo            LessonRepo
	DateInfoRepo    DateInfoRepo
	ClassroomRepo   ClassroomRepo

	// for zoom
	ZoomLink         string
	ZoomOwnerID      string
	ZoomID           string
	ZoomOccurrenceID string
	SchedulerInfo    *SchedulerInfo

	// for class do
	ClassDoLink    string
	ClassDoOwnerID string
	ClassDoRoomID  string

	// used by virtual classroom
	EndAt *time.Time
}

func (l *Lesson) GetDateInfoByDateAndCenterID(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, centerID string) ([]*calendar_dto.DateInfo, error) {
	dateInfo, err := l.DateInfoRepo.GetDateInfoByDateRangeAndLocationID(ctx, db, startDate, endDate, centerID)
	if err != nil {
		return nil, fmt.Errorf("could not get date info with time range %v - %v of center %s", startDate, endDate, centerID)
	}
	l.DateInfos = dateInfo
	return dateInfo, nil
}

func (l *Lesson) GetNonRegularDatesExceptFirstDate(ctx context.Context, dateInfo []*calendar_dto.DateInfo) map[string]string {
	excludeDates := make(map[string]string)
	if len(dateInfo) > 0 {
		for _, d := range dateInfo {
			dateType := d.DateTypeID
			location, _ := time.LoadLocation(d.TimeZone)
			lessonDate := l.StartTime.In(location).Format(Ymd)
			if dateType != string(calendar_constants.RegularDay) && d.Date.Format(Ymd) != lessonDate {
				excludeDates[d.Date.Format(calendar_constants.TimeLayout)] = d.TimeZone
			}
		}
	}
	return excludeDates
}

func (l *Lesson) CheckClosedDate(date time.Time, dateInfo []*calendar_dto.DateInfo) error {
	for _, d := range dateInfo {
		location, _ := time.LoadLocation(d.TimeZone)
		Date := date.In(location).Format(Ymd)
		if Date == d.Date.Format(Ymd) && d.DateTypeID == string(calendar_constants.ClosedDay) {
			return fmt.Errorf(" could not create lesson on closed date %s", d.DateTypeID)
		}
	}
	return nil
}

// IsValid checks the validity of a Lesson.
// Require lesson id, center id, start time, end time,
// scheduling status, teaching medium, teaching method
// For publish lesson
// Require at least 1 teacher
// Each learner only be appeared one time if have
// For draft lesson
// No require teacher or student
func (l *Lesson) IsValid(ctx context.Context, db database.Ext) error {
	if len(l.LessonID) == 0 {
		return fmt.Errorf("Lesson.LessonID cannot be empty")
	}

	if len(l.LocationID) == 0 {
		return fmt.Errorf("Lesson.LocationID cannot be empty")
	}

	if l.StartTime.IsZero() {
		return fmt.Errorf("start time could not be empty")
	}

	if l.EndTime.IsZero() {
		return fmt.Errorf("end time could not be empty")
	}

	if l.EndTime.Before(l.StartTime) {
		return fmt.Errorf("end time could not before start time")
	}

	if l.UpdatedAt.Before(l.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}

	if len(l.SchedulingStatus) == 0 {
		return fmt.Errorf("Lesson.SchedulingStatus cannot be empty")
	}

	if len(l.TeachingMedium) == 0 {
		return fmt.Errorf("Lesson.TeachingMedium cannot be empty")
	}

	if len(l.TeachingMethod) == 0 {
		return fmt.Errorf("Lesson.TeachingMethod cannot be empty")
	}

	if l.SchedulingStatus == LessonSchedulingStatusPublished {
		if len(l.Teachers) == 0 {
			return fmt.Errorf("Lesson.Teachers cannot be empty")
		}
	}
	if err := l.Learners.Validate(l.LocationID); err != nil {
		return fmt.Errorf("invalid learners: %w", err)
	}

	if err := l.Teachers.IsValid(); err != nil {
		return fmt.Errorf("invalid teachers: %w", err)
	}

	if l.Persisted {
		existedLesson, err := l.Repo.GetLessonByID(ctx, db, l.LessonID)
		if err != nil {
			return fmt.Errorf("could not get lesson id: %w", err)
		}
		// check for update teaching method: if a lesson is created, we won't be able
		// to update teaching_method later.
		if existedLesson.TeachingMethod != l.TeachingMethod {
			return fmt.Errorf("mismatched teaching method with lesson_id %s : in DB: %s, current entity:%s",
				existedLesson.LessonID, existedLesson.TeachingMethod, l.TeachingMethod)
		}
	}

	if _, err := l.MasterDataPort.GetLocationByID(ctx, db, l.LocationID); err != nil {
		return fmt.Errorf("could not get center by id %s: %w", l.LocationID, err)
	}

	teacherIDs := l.Teachers.GetIDs()
	if len(teacherIDs) > 0 {
		if err := l.UserModulePort.CheckTeacherIDs(ctx, teacherIDs); err != nil {
			return fmt.Errorf("could not get one of teachers %v: %w", teacherIDs, err)
		}
	}
	if len(l.Learners) > 0 {
		studentWithCourse := make([]string, 0, len(l.Learners)*2)
		for _, learner := range l.Learners {
			studentWithCourse = append(studentWithCourse, learner.LearnerID, learner.CourseID)
		}

		if err := l.UserModulePort.CheckStudentCourseSubscriptions(ctx, l.StartTime, studentWithCourse...); err != nil {
			return fmt.Errorf("could not get one of student course subscriptions: %w", err)
		}
	}

	if l.Material != nil && len(l.Material.MediaIDs) != 0 {
		medias, err := l.MediaModulePort.RetrieveMediasByIDs(ctx, l.Material.MediaIDs)
		if err != nil {
			return fmt.Errorf("mediaModulePort.RetrieveMediasByIDs: %w", err)
		}
		if len(medias.GetMediaIDs()) != len(l.Material.MediaIDs) {
			return fmt.Errorf("could not get one of medias %v: %v", l.Material.MediaIDs, medias.GetMediaIDs())
		}
	}
	// Lesson Group validations:
	// - teaching_method = "group"
	// - len(l.course_id) > 0
	// - l.course_id & l.class_id existed in DB
	// - saving draft course_id and class_id are not required
	if l.TeachingMethod == LessonTeachingMethodGroup {
		if len(l.CourseID) == 0 {
			if l.SchedulingStatus == LessonSchedulingStatusPublished {
				return fmt.Errorf("lesson with id %s have group teaching method but missing courseID", l.LessonID)
			}
		} else {
			if _, err := l.MasterDataPort.GetCourseByID(ctx, db, l.CourseID); err != nil {
				return fmt.Errorf("l.MasterDataPort.GetCourseByID: could not get course with id %s: %w", l.CourseID, err)
			}
		}
		if len(l.ClassID) > 0 {
			if _, err := l.MasterDataPort.GetClassByID(ctx, db, l.ClassID); err != nil {
				return fmt.Errorf(" l.MasterDataPort.GetClassByID: could not get class with id %s: %w", l.ClassID, err)
			}
		}
	}
	// validate date info of lesson
	if l.DateInfos == nil {
		if _, err := l.GetDateInfoByDateAndCenterID(ctx, db, l.StartTime, l.StartTime, l.LocationID); err != nil {
			return err
		}
	}
	if err := l.CheckClosedDate(l.StartTime, l.DateInfos); err != nil {
		return err
	}

	// validate classrooms
	if len(l.Classrooms) > 0 {
		classroomIDs := l.Classrooms.GetIDs()

		if err := l.ClassroomRepo.CheckClassroomIDs(ctx, db, classroomIDs); err != nil {
			return fmt.Errorf("failed in checking classroom IDs: %w", err)
		}
	}
	if l.TeachingMedium == "LESSON_TEACHING_MEDIUM_ZOOM" && l.SchedulingStatus != LessonSchedulingStatusDraft && l.ZoomLink == "" && l.ZoomOwnerID == "" {
		return fmt.Errorf("zoom link or zoom owner not should be empty")
	}
	return nil
}

func (l *Lesson) IsLock() bool {
	return l.IsLocked
}

func (l *Lesson) AddTeachers(teachers LessonTeachers) {
	l.Teachers = append(l.Teachers, teachers...)
}

func (l *Lesson) AddClassrooms(classrooms LessonClassrooms) {
	l.Classrooms = append(l.Classrooms, classrooms...)
}

func (l *Lesson) GetTeacherIDs() []string {
	ids := make([]string, 0, len(l.Teachers))
	for _, teacher := range l.Teachers {
		ids = append(ids, teacher.TeacherID)
	}
	return ids
}

func (l *Lesson) AddLearners(learners LessonLearners) {
	l.Learners = append(l.Learners, learners...)
}

func (l *Lesson) GetLearnersIDs() []string {
	ids := make([]string, 0, len(l.Learners))
	for _, learner := range l.Learners {
		ids = append(ids, learner.LearnerID)
	}
	return ids
}

func (l *Lesson) AddMaterials(mediaIDs []string) {
	if len(mediaIDs) == 0 {
		return
	}

	mediaIDs = utils.GetUniqueElementStringArray(mediaIDs)
	if l.Material != nil {
		l.Material.MediaIDs = mediaIDs
	} else {
		l.Material = &LessonMaterial{
			MediaIDs: mediaIDs,
		}
	}
}

// preInsert will do:
//   - Generate lesson id
func (l *Lesson) preInsert() {
	l.LessonID = idutil.ULIDNow()
}

func (l *Lesson) SaveOneTime() {
	if !l.Persisted {
		l.preInsert()
	}
}

func (l *Lesson) ResetID() {
	l.LessonID = ""
	l.Persisted = false
}

func (l *Lesson) CheckLessonValidForPublish() bool {
	if len(l.Teachers) == 0 {
		return false
	}
	if l.TeachingMethod == LessonTeachingMethodGroup {
		if len(l.CourseID) == 0 {
			return false
		}
	}
	return true
}

type UpdateLessonTeacherName struct {
	TeacherID string
	FullName  string
}

type ListLessonArgs struct {
	TeacherIDs  []string
	StudentIDs  []string
	CourseIDs   []string
	LocationIDs []string
	Grades      []int32
	GradesV2    []string
	FromDate    time.Time
	ToDate      time.Time
	TimeZone    string
	Dow         []DateOfWeek // date_of_weeks
	FromTime    string
	ToTime      string

	Limit       uint32
	KeyWord     string
	LessonTime  string
	CurrentTime time.Time
	Compare     string
	SchoolID    string
	LessonID    string
}

type ListStudentsByLessonArgs struct {
	LessonID string
	Limit    uint32

	// used for pagination
	UserName string
	UserID   string
}

type ListMediaByLessonArgs struct {
	LessonID string

	// used for pagination
	Limit  uint32
	Offset string
}

type QueryLesson struct {
	ClassID   string
	StartTime *time.Time
	EndTime   *time.Time
}

type UpdateLessonMemberReport struct {
	LessonID         string
	StudentID        string
	AttendanceStatus string
	AttendanceReason string
	AttendanceNotice string
	AttendanceNote   string
}

type UpdateLessonMemberName struct {
	LessonID      string
	StudentID     string
	UserFirstName string
	UserLastName  string
}
