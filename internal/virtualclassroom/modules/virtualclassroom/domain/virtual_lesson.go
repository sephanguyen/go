package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
)

type (
	LessonSchedulingStatus string
	LessonTeachingMedium   string
	LessonTeachingMethod   string
	StudentAttendStatus    string
	DateOfWeek             int32
)

const (
	LessonSchedulingStatusDraft     LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_DRAFT"
	LessonSchedulingStatusPublished LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_PUBLISHED"
	LessonSchedulingStatusCompleted LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_COMPLETED"
	LessonSchedulingStatusCanceled  LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_CANCELED"

	LessonTeachingMediumOffline LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_OFFLINE"
	LessonTeachingMediumOnline  LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_ONLINE"
	LessonTeachingMediumZoom    LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_ZOOM"
	LessonTeachingMediumHybrid  LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_HYBRID"
	LessonTeachingMediumClassDo LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_CLASS_DO"

	LessonTeachingMethodIndividual LessonTeachingMethod = "LESSON_TEACHING_METHOD_INDIVIDUAL"
	LessonTeachingMethodGroup      LessonTeachingMethod = "LESSON_TEACHING_METHOD_GROUP"

	StudentAttendStatusEmpty          StudentAttendStatus = "STUDENT_ATTEND_STATUS_EMPTY"
	StudentAttendStatusAttend         StudentAttendStatus = "STUDENT_ATTEND_STATUS_ATTEND"
	StudentAttendStatusAbsent         StudentAttendStatus = "STUDENT_ATTEND_STATUS_ABSENT"
	StudentAttendStatusLate           StudentAttendStatus = "STUDENT_ATTEND_STATUS_LATE"
	StudentAttendStatusInformedAbsent StudentAttendStatus = "STUDENT_ATTEND_STATUS_INFORMED_ABSENT"
	StudentAttendStatusInformedLate   StudentAttendStatus = "STUDENT_ATTEND_STATUS_INFORMED_LATE"

	Monday    DateOfWeek = 0
	Tuesday   DateOfWeek = 1
	Wednesday DateOfWeek = 2
	Thurday   DateOfWeek = 3
	Friday    DateOfWeek = 4
	Saturday  DateOfWeek = 5
	Sunday    DateOfWeek = 6
)

type LessonMaterial struct {
	MediaIDs []string
}
type VirtualLesson struct {
	LessonID         string
	Name             string
	CenterID         string
	CourseID         string
	TeacherID        string
	ClassID          string
	StartTime        time.Time
	EndTime          time.Time
	SchedulingStatus LessonSchedulingStatus
	TeachingMedium   LessonTeachingMedium
	TeachingMethod   LessonTeachingMethod
	Learners         LessonLearners
	Teachers         LessonTeachers
	Material         *LessonMaterial
	LessonGroupID    string
	SchedulerID      string
	RoomState        OldLessonRoomState
	RoomID           string
	EndAt            *time.Time
	ZoomLink         string
	LessonCapacity   int32
	ClassDoOwnerID   string
	ClassDoLink      string

	// default timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	virtualLessonPort VirtualLessonPort

	LearnerIDs
	TeacherIDs
}

type OldLessonRoomState struct {
	CurrentMaterial *CurrentMaterial `json:"current_material,omitempty"`
	CurrentPolling  *CurrentPolling  `json:"current_polling,omitempty"`
	Recording       *RecordingState  `json:"recording,omitempty"`
}

func (l *OldLessonRoomState) IsValid() error {
	if l.CurrentMaterial != nil {
		if err := l.CurrentMaterial.IsValid(); err != nil {
			return fmt.Errorf("invalid current_material: %v", err)
		}
	}
	return nil
}

type LearnerIDs struct {
	LearnerIDs []string
}

func (ids LearnerIDs) HaveID(id string) bool {
	for _, v := range ids.LearnerIDs {
		if v == id {
			return true
		}
	}

	return false
}

type TeacherIDs struct {
	TeacherIDs []string
}

func (ids TeacherIDs) HaveID(id string) bool {
	for _, v := range ids.TeacherIDs {
		if v == id {
			return true
		}
	}

	return false
}

func (v *VirtualLesson) IsValid(ctx context.Context) error {
	if len(v.LessonID) == 0 {
		return fmt.Errorf("VirtualLesson.isValid: Lesson ID cannot be empty")
	}

	if ok, err := v.virtualLessonPort.IsLessonMediumOnline(ctx, v.LessonID); err == nil {
		if !ok {
			return fmt.Errorf("VirtualLesson.isValid: expected lesson id %s have medium attribute is online but is not", v.LessonID)
		}
	} else {
		return fmt.Errorf("VirtualLesson.isValid: lessonPort.IsLessonMediumOnline: %w", err)
	}

	return nil
}

func (v *VirtualLesson) CheckLessonMemberIDs(ctx context.Context, userIDs []string) error {
	res, err := v.virtualLessonPort.CheckLessonMemberIDs(ctx, v.LessonID, userIDs)
	if err != nil {
		return fmt.Errorf("VirtualLesson.CheckLessonMemberIDs: Lesson.lessonRepo.CheckLessonMemberIDs: %w", err)
	}

	memberIDs := golibs.StringSliceToMap(res)
	for _, userID := range userIDs {
		if _, ok := memberIDs[userID]; !ok {
			return fmt.Errorf("user ID %s is not belong lesson %s", userID, v.LessonID)
		}
	}

	return nil
}

func (v *VirtualLesson) AddTeacherIDs(teacherIDs []string) {
	v.TeacherIDs.TeacherIDs = teacherIDs
}

func GetLessonIDs(lessons []VirtualLesson) []string {
	lessonIDs := make([]string, 0, len(lessons))
	for _, lesson := range lessons {
		lessonIDs = append(lessonIDs, lesson.LessonID)
	}
	return lessonIDs
}

func NewVirtualLesson() *VirtualLessonBuilder {
	return &VirtualLessonBuilder{
		virtualLesson: &VirtualLesson{},
	}
}

type VirtualLessonBuilder struct {
	virtualLesson *VirtualLesson
}

func (l *VirtualLessonBuilder) WithLessonID(id string) *VirtualLessonBuilder {
	l.virtualLesson.LessonID = id
	return l
}

func (l *VirtualLessonBuilder) WithDeletedAtTime(deletedAt *time.Time) *VirtualLessonBuilder {
	l.virtualLesson.DeletedAt = deletedAt
	return l
}

func (l *VirtualLessonBuilder) WithModificationTime(createdAt, updatedAt time.Time) *VirtualLessonBuilder {
	l.virtualLesson.CreatedAt = createdAt
	l.virtualLesson.UpdatedAt = updatedAt
	return l
}

func (l *VirtualLessonBuilder) WithCenterID(centerID string) *VirtualLessonBuilder {
	l.virtualLesson.CenterID = centerID
	return l
}
func (l *VirtualLessonBuilder) WithCourseID(courseID string) *VirtualLessonBuilder {
	l.virtualLesson.CourseID = courseID
	return l
}
func (l *VirtualLessonBuilder) WithTeacherID(teacherID string) *VirtualLessonBuilder {
	l.virtualLesson.TeacherID = teacherID
	return l
}
func (l *VirtualLessonBuilder) WithClassID(classID string) *VirtualLessonBuilder {
	l.virtualLesson.ClassID = classID
	return l
}
func (l *VirtualLessonBuilder) WithLessonGroupID(lessonGroupID string) *VirtualLessonBuilder {
	l.virtualLesson.LessonGroupID = lessonGroupID
	return l
}

func (l *VirtualLessonBuilder) WithRoomState(roomState *OldLessonRoomState) *VirtualLessonBuilder {
	l.virtualLesson.RoomState = *roomState
	return l
}

func (l *VirtualLessonBuilder) WithName(name string) *VirtualLessonBuilder {
	l.virtualLesson.Name = name
	return l
}

func (l *VirtualLessonBuilder) WithTimeRange(startTime, endTime time.Time) *VirtualLessonBuilder {
	l.virtualLesson.StartTime = startTime
	l.virtualLesson.EndTime = endTime
	return l
}

func (l *VirtualLessonBuilder) WithDeletedTime(deletedAt *time.Time) *VirtualLessonBuilder {
	l.virtualLesson.DeletedAt = deletedAt
	return l
}

func (l *VirtualLessonBuilder) WithMaterials(mediaIDs []string) *VirtualLessonBuilder {
	if len(mediaIDs) == 0 {
		return l
	}

	mediaIDs = golibs.GetUniqueElementStringArray(mediaIDs)
	if l.virtualLesson.Material != nil {
		l.virtualLesson.Material.MediaIDs = mediaIDs
	} else {
		l.virtualLesson.Material = &LessonMaterial{
			MediaIDs: mediaIDs,
		}
	}
	return l
}

func (l *VirtualLessonBuilder) WithTeachingMedium(medium LessonTeachingMedium) *VirtualLessonBuilder {
	l.virtualLesson.TeachingMedium = medium
	return l
}

func (l *VirtualLessonBuilder) WithTeachingMethod(method LessonTeachingMethod) *VirtualLessonBuilder {
	l.virtualLesson.TeachingMethod = method
	return l
}

func (l *VirtualLessonBuilder) WithTeacherIDs(teacherIDs []string) *VirtualLessonBuilder {
	teacherIDs = golibs.GetUniqueElementStringArray(teacherIDs)
	l.virtualLesson.Teachers = make(LessonTeachers, 0, len(teacherIDs))
	for _, id := range teacherIDs {
		l.virtualLesson.Teachers = append(l.virtualLesson.Teachers, &LessonTeacher{
			TeacherID: id,
		})
	}
	l.virtualLesson.TeacherIDs.TeacherIDs = teacherIDs
	return l
}

func (l *VirtualLessonBuilder) WithLearnerIDs(learnerIDs []string) *VirtualLessonBuilder {
	l.virtualLesson.LearnerIDs.LearnerIDs = learnerIDs
	return l
}

func (l *VirtualLessonBuilder) WithSchedulingStatus(schedulingStatus LessonSchedulingStatus) *VirtualLessonBuilder {
	l.virtualLesson.SchedulingStatus = schedulingStatus
	return l
}

func (l *VirtualLessonBuilder) WithLearners(learners LessonLearners) *VirtualLessonBuilder {
	l.virtualLesson.Learners = learners
	return l
}

func (l *VirtualLessonBuilder) WithSchedulerID(schedulerID string) *VirtualLessonBuilder {
	l.virtualLesson.SchedulerID = schedulerID
	return l
}

func (l *VirtualLessonBuilder) WithRoomID(roomID string) *VirtualLessonBuilder {
	l.virtualLesson.RoomID = roomID
	return l
}

func (l *VirtualLessonBuilder) WithZoomLink(zoomLink string) *VirtualLessonBuilder {
	l.virtualLesson.ZoomLink = zoomLink
	return l
}

func (l *VirtualLessonBuilder) WithEndAt(endAt *time.Time) *VirtualLessonBuilder {
	l.virtualLesson.EndAt = endAt
	return l
}

func (l *VirtualLessonBuilder) WithLessonCapacity(cap int32) *VirtualLessonBuilder {
	l.virtualLesson.LessonCapacity = cap
	return l
}

func (l *VirtualLessonBuilder) WithClassDoOwnerID(ownerID string) *VirtualLessonBuilder {
	l.virtualLesson.ClassDoOwnerID = ownerID
	return l
}

func (l *VirtualLessonBuilder) WithClassDoLink(link string) *VirtualLessonBuilder {
	l.virtualLesson.ClassDoLink = link
	return l
}

// BuildDraft will return a lesson but not valid data
func (l *VirtualLessonBuilder) BuildDraft() *VirtualLesson {
	return l.virtualLesson
}

type VirtualLessonPort interface {
	IsLessonMediumOnline(ctx context.Context, lessonID string) (bool, error)
	CheckLessonMemberIDs(ctx context.Context, lessonID string, userIDs []string) (memberIDs []string, err error)
}

type ListMediaByLessonArgs struct {
	LessonID string

	// used for pagination
	Limit  uint32
	Offset string
}
