package domain

import (
	"fmt"
	"time"
)

type (
	LessonLearner struct {
		LearnerID        string
		CourseID         string
		LocationID       string
		AttendStatus     StudentAttendStatus
		AttendanceNotice StudentAttendanceNotice
		AttendanceReason StudentAttendanceReason
		AttendanceNote   string
		Grade            string
		CourseName       string
		LearnerName      string
		*Reallocate
	}
	Reallocate struct {
		OriginalLessonID string
	}
)

func NewLessonLearner(learnerId, courseId, locationId, attendStatus, attendNotice, attendReason, attendNote string) *LessonLearner {
	ll := &LessonLearner{
		LearnerID:        learnerId,
		CourseID:         courseId,
		LocationID:       locationId,
		AttendStatus:     StudentAttendStatus(attendStatus),
		AttendanceNotice: StudentAttendanceNotice(attendNotice),
		AttendanceReason: StudentAttendanceReason(attendReason),
		AttendanceNote:   attendNote,
	}
	return ll
}

func (l LessonLearner) Validate() error {
	if len(l.LearnerID) == 0 {
		return fmt.Errorf("Lesson.Learner.LearnerID cannot be empty")
	}

	if len(l.CourseID) == 0 {
		return fmt.Errorf("Lesson.Learner.CourseID of learner %s cannot be empty", l.LearnerID)
	}

	if len(l.AttendStatus) == 0 {
		return fmt.Errorf("Lesson.Learner.AttendStatus of learner %s cannot be empty", l.LearnerID)
	}

	if len(l.AttendanceNotice) == 0 {
		return fmt.Errorf("Lesson.Learner.AttendanceNotice of learner %s cannot be empty", l.LearnerID)
	}

	if len(l.AttendanceReason) == 0 {
		return fmt.Errorf("Lesson.Learner.AttendanceReason of learner %s cannot be empty", l.LearnerID)
	}

	return nil
}

func (l *LessonLearner) EmptyAttendanceInfo() {
	l.AttendStatus = StudentAttendStatusEmpty
	l.AttendanceNote = ""
	l.AttendanceNotice = NoticeEmpty
	l.AttendanceReason = ReasonEmpty
}

func (l *LessonLearner) SetAttendanceInfoIfPresentOnMap(attendanceMap map[string]*LessonLearner) {
	if attendanceInfo, ok := attendanceMap[l.LearnerID]; ok {
		l.AttendStatus = attendanceInfo.AttendStatus
		l.AttendanceNote = attendanceInfo.AttendanceNote
		l.AttendanceNotice = attendanceInfo.AttendanceNotice
		l.AttendanceReason = attendanceInfo.AttendanceReason
	} else {
		l.EmptyAttendanceInfo()
	}
}

func (l *LessonLearner) AddGrade(grade string) {
	l.Grade = grade
}

func (l *LessonLearner) AddCourseName(courseName string) {
	l.CourseName = courseName
}

func (l *LessonLearner) AddLearnerName(learnerName string) {
	l.LearnerName = learnerName
}

func (l *LessonLearner) AddReallocate(lessonID string) {
	if len(lessonID) > 0 {
		l.Reallocate = &Reallocate{
			OriginalLessonID: lessonID,
		}
	}
}

func (l *LessonLearner) IsReallocateStatus() bool {
	return l.AttendStatus == StudentAttendStatusReallocate
}

func (l *LessonLearner) IsReallocated() bool {
	return l.Reallocate != nil
}

func (l *LessonLearner) StudentWithCourse() string {
	return l.LearnerID + "-" + l.CourseID
}

func (l *LessonLearner) IsValidForAllocateToLesson(endDate, lessonDate time.Time, tz *time.Location) bool {
	return lessonDate.In(tz).Format(Ymd) <= endDate.In(tz).Format(Ymd)
}

type LessonLearners []*LessonLearner

func (l LessonLearners) Validate(locationID string) error {
	for i := range l {
		if err := l[i].Validate(); err != nil {
			return err
		}
		if l[i].Reallocate != nil {
			continue
		}
		if l[i].LocationID != locationID {
			return fmt.Errorf("locationID of learner %s must be %s but got %s", l[i].LearnerID, locationID, l[i].LocationID)
		}
	}

	learnerWithCourse := make(map[string]string)
	for _, learner := range l {
		if courseID, ok := learnerWithCourse[learner.LearnerID]; ok {
			return fmt.Errorf("duplicated learner %s and course %s", learner.LearnerID, courseID)
		}
		learnerWithCourse[learner.LearnerID] = learner.CourseID
	}

	return nil
}

func (l LessonLearners) GetLearnerIDs() []string {
	ids := make([]string, 0, len(l))
	for _, u := range l {
		ids = append(ids, u.LearnerID)
	}
	return ids
}

func (l LessonLearners) GetCourseIDsOfLessonLearners() []string {
	ids := make(map[string]bool)
	for _, learner := range l {
		ids[learner.CourseID] = true
	}

	res := make([]string, 0, len(ids))
	for id := range ids {
		res = append(res, id)
	}

	return res
}

func (l LessonLearners) GetStudentNoPendingReallocate(oldLearner LessonLearners) []string {
	newStudentMap := make(map[string]StudentAttendStatus)
	for _, item := range l {
		newStudentMap[item.LearnerID] = item.AttendStatus
	}

	var studentIds []string
	for _, item := range oldLearner {
		newStatus, existed := newStudentMap[item.LearnerID]
		if !existed {
			continue
		}
		if item.AttendStatus == StudentAttendStatusReallocate {
			if newStatus != StudentAttendStatusReallocate {
				studentIds = append(studentIds, item.LearnerID)
			}
		}
	}
	return studentIds
}

func (l LessonLearners) GetStudentReallocate(studentAttendanceStatus map[string]StudentAttendStatus) []string {
	var studentIds []string
	for _, item := range l {
		newStatus, existed := studentAttendanceStatus[item.LearnerID]
		if !existed {
			continue
		}
		if item.AttendStatus != StudentAttendStatusReallocate && newStatus == StudentAttendStatusReallocate {
			studentIds = append(studentIds, item.LearnerID)
		}
	}
	return studentIds
}

func (l LessonLearners) GetStudentUnReallocate(studentAttendanceStatus map[string]StudentAttendStatus) []string {
	var studentID []string
	for _, item := range l {
		newStatus, existed := studentAttendanceStatus[item.LearnerID]
		if !existed {
			continue
		}
		if item.AttendStatus == StudentAttendStatusReallocate && newStatus != StudentAttendStatusReallocate {
			studentID = append(studentID, item.LearnerID)
		}
	}
	return studentID
}

func (l LessonLearners) GroupByLearnerID() map[string]*LessonLearner {
	learnersMap := make(map[string]*LessonLearner, len(l))
	for _, item := range l {
		learnersMap[item.LearnerID] = item
	}
	return learnersMap
}

func (l LessonLearners) GetReallocateStudentRemoved(oldLearner LessonLearners) []string {
	newStudentMap := make(map[string]*LessonLearner)
	for _, item := range l {
		newStudentMap[item.LearnerID] = item
	}
	var studentIds []string
	for _, item := range oldLearner {
		newStudent, existed := newStudentMap[item.LearnerID]
		if !existed || (item.Reallocate != nil && newStudent.Reallocate == nil) {
			studentIds = append(studentIds, item.LearnerID)
		}
	}
	return studentIds
}

func (l LessonLearners) GetStudentReallocatedDiffLocation(locationID string) []string {
	userReallocated := []string{}
	for _, learner := range l {
		if learner.Reallocate != nil {
			if learner.LocationID != locationID {
				userReallocated = append(userReallocated, learner.LearnerID)
			}
		}
	}
	return userReallocated
}

func (l LessonLearners) GetLearnerByID(studentID string) *LessonLearner {
	for _, u := range l {
		if u.LearnerID == studentID {
			return u
		}
	}
	return nil
}

func (l LessonLearners) GetStudentRemoved(oldLearners LessonLearners) []string {
	newStudentMap := make(map[string]*LessonLearner)
	for _, item := range l {
		newStudentMap[item.LearnerID] = item
	}
	var studentIds []string
	for _, item := range oldLearners {
		if student, exists := newStudentMap[item.LearnerID]; !exists {
			studentIds = append(studentIds, item.LearnerID)
		} else if student.IsReallocated() {
			studentIds = append(studentIds, item.LearnerID)
		}
	}
	return studentIds
}

func (l LessonLearners) IsChangedStudentInfo(oldLearners LessonLearners) bool {
	if len(l) != len(oldLearners) {
		return true
	}
	learners := make(map[string]*LessonLearner, len(l))
	for _, u := range l {
		learners[u.LearnerID] = u
	}
	for _, oldLearner := range oldLearners {
		newLearner, exist := learners[oldLearner.LearnerID]
		if !exist {
			return true
		}
		if newLearner.IsReallocated() != oldLearner.IsReallocated() {
			return true
		}
	}
	return false
}
