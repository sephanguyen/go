package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
)

type CreateLesson struct {
	Lesson   *domain.Lesson
	TimeZone string
}

type ZoomOccurrence struct {
	OccurrenceID string
	StartTime    string
}
type ZoomInfo struct {
	ZoomAccountID   string
	ZoomLink        string
	ZoomID          string
	ZoomOccurrences []*ZoomOccurrence
}

type CreateRecurringLesson struct {
	Lesson   *domain.Lesson
	RRuleCmd RecurrenceRuleCommand
	TimeZone string
	ZoomInfo *ZoomInfo
}

type RecurrenceRuleCommand struct {
	StartTime time.Time
	EndTime   time.Time
	UntilDate time.Time
}
type ImportLessonCommand struct {
	Lesson       *domain.Lesson
	StartTime    time.Time
	EndTime      time.Time
	UntilDate    time.Time
	SavingMethod lpb.CreateLessonSavingMethod
}
type ImportLessonPayload struct {
	Payloads                 []ImportLessonCommand
	Scanner                  scanner.CSVScanner
	Timezone                 string
	PartnerInternalIDs       []string
	CenterByPartnerID        map[string]*domain.Location
	IsUsingVersion2          bool
	StudentIDWithCourseIDs   []string // "student_id_1", "course_id_1", "student_id_2", "course_id_2"
	StudentCourseAccessPaths map[string][]string

	// port
	MasterDataPort          infrastructure.MasterDataPort
	DateInfoRepo            infrastructure.DateInfoRepo
	UserModulePort          infrastructure.UserModulePort
	StudentSubscriptionRepo user_infras.StudentSubscriptionRepo
}

func NewImportLessonPayload() *ImportLessonPayload {
	return &ImportLessonPayload{}
}

func (p *ImportLessonPayload) WithTimeZone(timezone string) *ImportLessonPayload {
	p.Timezone = timezone
	return p
}

func (p *ImportLessonPayload) WithScanner(sc scanner.CSVScanner) *ImportLessonPayload {
	p.Scanner = sc
	return p
}

func (p *ImportLessonPayload) WithPartnerInternalIDs(pIDs []string) *ImportLessonPayload {
	p.PartnerInternalIDs = pIDs
	return p
}

func (p *ImportLessonPayload) WithStudentCourseIDs(ids []string) *ImportLessonPayload {
	p.StudentIDWithCourseIDs = ids
	return p
}

func (p *ImportLessonPayload) WithVersion2(isVersion2 bool) *ImportLessonPayload {
	p.IsUsingVersion2 = isVersion2
	return p
}

func (p *ImportLessonPayload) WithMasterDataPort(port infrastructure.MasterDataPort) *ImportLessonPayload {
	p.MasterDataPort = port
	return p
}

func (p *ImportLessonPayload) WithUserModulePort(port infrastructure.UserModulePort) *ImportLessonPayload {
	p.UserModulePort = port
	return p
}

func (p *ImportLessonPayload) WithDateInfoRepo(port infrastructure.DateInfoRepo) *ImportLessonPayload {
	p.DateInfoRepo = port
	return p
}

func (p *ImportLessonPayload) WithStudentSubscriptionRepo(port user_infras.StudentSubscriptionRepo) *ImportLessonPayload {
	p.StudentSubscriptionRepo = port
	return p
}

func (p *ImportLessonPayload) prepareData(ctx context.Context, db database.Ext) map[int]error {
	errors := make(map[int]error)
	centerByPartnerID, err := p.MasterDataPort.GetLowestLocationsByPartnerInternalIDs(ctx, db, p.PartnerInternalIDs)
	if err != nil || len(centerByPartnerID) == 0 {
		errors[0] = fmt.Errorf("invalid all partner_internal_id")
		return errors
	}
	p.CenterByPartnerID = centerByPartnerID

	if p.IsUsingVersion2 && len(p.StudentIDWithCourseIDs) > 0 {
		locationIDs := make([]string, len(p.PartnerInternalIDs))
		for _, loc := range centerByPartnerID {
			locationIDs = append(locationIDs, loc.LocationID)
		}
		studentSubscriptions, err := p.StudentSubscriptionRepo.GetStudentCourseSubscriptions(ctx, db, locationIDs, p.StudentIDWithCourseIDs...)
		if err != nil || len(studentSubscriptions) == 0 {
			errors[0] = fmt.Errorf("cannot get student course subscriptions")
			return errors
		}

		studentSubMap := make(map[string][]string)
		for _, ss := range studentSubscriptions {
			studentCourseID := fmt.Sprintf("%s/%s", ss.StudentID, ss.CourseID)
			studentSubMap[studentCourseID] = ss.LocationIDs
		}
		p.StudentCourseAccessPaths = studentSubMap
	}
	return errors
}

func (p *ImportLessonPayload) buildImportLessonPayload(ctx context.Context) map[int]error {
	sc := p.Scanner
	errors := make(map[int]error)

	for sc.Scan() {
		pID := sc.Text("partner_internal_id")
		currentRow := sc.GetCurRow()
		if _, ok := p.CenterByPartnerID[pID]; !ok {
			errors[currentRow] = fmt.Errorf("could not get center by partner_internal_id")
			continue
		}
		centerID := p.CenterByPartnerID[pID].LocationID

		startDateTime, err1 := timeutil.ParsingTimeFromYYYYMMDDStr(sc.RawText("start_date_time"), p.Timezone)
		endDateTime, err2 := timeutil.ParsingTimeFromYYYYMMDDStr(sc.RawText("end_date_time"), p.Timezone)
		err := multierr.Combine(err1, err2)
		if err != nil {
			errors[currentRow] = fmt.Errorf("could not parsing time")
			continue
		}
		if startDateTime.After(endDateTime) || startDateTime.Equal(endDateTime) {
			errors[currentRow] = fmt.Errorf("start time should difference or earlier than end time")
			continue
		}
		endTime := time.Date(
			startDateTime.Year(), startDateTime.Month(), startDateTime.Day(),
			endDateTime.Hour(), endDateTime.Minute(), endDateTime.Second(), endDateTime.Nanosecond(),
			endDateTime.Location())

		savingMethod := lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME
		if startDateTime.Format(domain.YYYMMDD) != endDateTime.Format(domain.YYYMMDD) {
			savingMethod = lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE
		}

		// in this phase, will not create recurring lesson by import file CSV
		if savingMethod != lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME {
			errors[currentRow] = fmt.Errorf("start date not matches with end date")
			continue
		}

		teachingMethod, ok := domain.MapValueLessonTeachingMethod[sc.Text("teaching_method")]
		if !ok {
			errors[currentRow] = fmt.Errorf("invalid teaching method value")
			continue
		}

		teachingMedium := domain.LessonTeachingMediumOffline
		if p.IsUsingVersion2 && sc.Text("teaching_medium") != "" {
			teachingMedium, ok = domain.MapValueLessonTeachingMedium[sc.Text("teaching_medium")]
			if !ok {
				errors[currentRow] = fmt.Errorf("invalid teaching medium value")
				continue
			}
		}

		builder := domain.NewLesson().
			WithLocationID(centerID).
			WithTimeRange(startDateTime, endTime).
			WithTeachingMedium(teachingMedium).
			WithTeachingMethod(teachingMethod).
			WithMasterDataPort(p.MasterDataPort).
			WithUserModulePort(p.UserModulePort).
			WithModificationTime(time.Now(), time.Now()).
			WithDateInfoRepo(p.DateInfoRepo).
			WithSchedulingStatus(domain.LessonSchedulingStatusDraft)

		if p.IsUsingVersion2 {
			for _, sCourseID := range strings.Split(sc.Text("student_course_ids"), "_") {
				if sCourseID == "" {
					continue
				}
				studentCourseAP, ok := p.StudentCourseAccessPaths[sCourseID]
				if !ok {
					errors[currentRow] = fmt.Errorf("cannot get student course subscription access path (%s)", sCourseID)
					break
				}

				if !slices.Contains(studentCourseAP, centerID) {
					errors[currentRow] = fmt.Errorf("student with this course (%s) must have the location same with lesson's location", sCourseID)
					break
				}

				if studentCourseIDs := strings.Split(sCourseID, "/"); len(studentCourseIDs) > 1 {
					learner := domain.LessonLearner{
						LearnerID:  studentCourseIDs[0],
						CourseID:   studentCourseIDs[1],
						LocationID: centerID,
					}
					learner.EmptyAttendanceInfo()
					builder.AddLearner(&learner)
				}
			}
			if sc.Text("teacher_ids") != "" {
				teacherIDs := strings.Split(sc.Text("teacher_ids"), "_")
				builder.WithTeacherIDs(teacherIDs)
			}
		}

		lesson := builder.BuildDraft()

		if p.IsUsingVersion2 {
			if len(lesson.Learners) > 0 {
				if err := lesson.Learners.Validate(centerID); err != nil {
					errors[currentRow] = fmt.Errorf("invalid learners: %w", err)
				}
				studentWithCourse := make([]string, 0, len(lesson.Learners)*2)
				for _, learner := range lesson.Learners {
					studentWithCourse = append(studentWithCourse, learner.LearnerID, learner.CourseID)
				}

				if err := lesson.UserModulePort.CheckStudentCourseSubscriptions(ctx, lesson.StartTime, studentWithCourse...); err != nil {
					errors[currentRow] = fmt.Errorf("could not get one of student course subscriptions: %w", err)
				}
			}

			teacherIDs := lesson.Teachers.GetIDs()
			if len(teacherIDs) > 0 {
				if err := lesson.Teachers.IsValid(); err != nil {
					errors[currentRow] = fmt.Errorf("invalid teachers: %w", err)
				}
				if err := lesson.UserModulePort.CheckTeacherIDs(ctx, teacherIDs); err != nil {
					errors[currentRow] = fmt.Errorf("could not get one of teachers %v: %w", teacherIDs, err)
				}
			}
		}

		payload := ImportLessonCommand{
			Lesson:       lesson,
			SavingMethod: savingMethod,
			StartTime:    startDateTime,
			EndTime:      endTime,
			UntilDate:    endDateTime,
		}
		p.Payloads = append(p.Payloads, payload)
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}
