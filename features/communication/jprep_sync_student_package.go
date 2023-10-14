package communication

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/notification/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/cucumber/godog"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type JprepSyncStudentPackageSuite struct {
	*common.NotificationSuite
	EventUpserts          []*npb.EventSyncStudentPackage
	ExpectedStudentCourse map[string]*entities.NotificationStudentCourse
	StudentIDs            []string
	CourseIDs             []string
}

func (c *SuiteConstructor) InitJprepSyncStudentPackage(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &JprepSyncStudentPackageSuite{
		NotificationSuite:     dep.notiCommonSuite,
		EventUpserts:          make([]*npb.EventSyncStudentPackage, 0),
		ExpectedStudentCourse: make(map[string]*entities.NotificationStudentCourse, 0),
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" of Jprep organization with default location logged in Back Office$`: s.StaffGrantedRoleOfJprepOrganizationWithDefaultLocationLoggedInBackOffice,
		`^school admin creates "([^"]*)" students$`:                                           s.CreatesNumberOfStudents,
		`^school admin creates "([^"]*)" courses$`:                                            s.CreatesNumberOfCourses,
		`^nats events are published$`:                                                         s.natsEventsArePublished,
		`^notification system must store student course data correctly with type "([^"]*)"$`:  s.notificationSystemMustStoreStudentCourseDataCorrectly,
		`^some valid upsert event sync student course from Yasuo$`:                            s.someValidEventSyncStudentCourseFromYasuo,
		`^school admin "([^"]*)" some student course$`:                                        s.schoolAdminSomeStudentCourse,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *JprepSyncStudentPackageSuite) someValidEventSyncStudentCourseFromYasuo(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	courseIDs := []string{}
	for _, course := range stepState.Courses {
		courseIDs = append(courseIDs, course.ID)
	}
	s.CourseIDs = courseIDs

	timeNow := time.Now()
	startDate := &timestamppb.Timestamp{
		Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
	}
	endDate := &timestamppb.Timestamp{
		Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
	}
	locationID := stepState.Organization.DefaultLocation.ID
	events := []*npb.EventSyncStudentPackage{}
	for _, student := range stepState.Students {
		s.StudentIDs = append(s.StudentIDs, student.ID)
		for _, courseID := range courseIDs {
			item := &entities.NotificationStudentCourse{}
			database.AllNullEntity(item)
			err := multierr.Combine(
				item.StudentID.Set(student.ID),
				item.CourseID.Set(courseID),
				item.LocationID.Set(locationID),
				item.StartAt.Set(startDate.AsTime()),
				item.EndAt.Set(endDate.AsTime()),
				item.DeletedAt.Set(nil),
			)
			if err != nil {
				return ctx, fmt.Errorf("multierr %v", err)
			}
			keyStr := fmt.Sprintf("%s-%s-%s", student.ID, courseID, locationID)
			s.ExpectedStudentCourse[keyStr] = item
		}
		events = append(events, &npb.EventSyncStudentPackage{
			StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
				{
					StudentId:  student.ID,
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					Packages: []*npb.EventSyncStudentPackage_Package{
						{
							CourseIds: courseIDs,
							StartDate: startDate,
							EndDate:   endDate,
						},
					},
				},
			},
		})
	}
	s.EventUpserts = events
	return ctx, nil
}

func (s *JprepSyncStudentPackageSuite) schoolAdminSomeStudentCourse(ctx context.Context, action string) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	locationID := stepState.Organization.DefaultLocation.ID
	for _, event := range s.EventUpserts {
		switch action {
		case "update":
			for _, studentPackage := range event.StudentPackages {
				for _, pkg := range studentPackage.Packages {
					endAt := time.Time{}
					for _, course := range pkg.CourseIds {
						keyStr := fmt.Sprintf("%s-%s-%s", studentPackage.StudentId, course, locationID)
						if val, ok := s.ExpectedStudentCourse[keyStr]; ok {
							currentEndAt := &time.Time{}
							err := val.EndAt.AssignTo(&currentEndAt)
							if err != nil {
								return ctx, err
							}
							endAt = currentEndAt.AddDate(0, 0, 1) // add one more day
							err = val.EndAt.Set(endAt)
							if err != nil {
								return ctx, err
							}
						}
					}
					pkg.EndDate = &timestamppb.Timestamp{
						Seconds: endAt.Unix(),
					}
				}
				studentPackage.ActionKind = npb.ActionKind_ACTION_KIND_UPSERTED
			}
		case "delete":
			for _, studentPackage := range event.StudentPackages {
				for _, pkg := range studentPackage.Packages {
					for _, course := range pkg.CourseIds {
						keyStr := fmt.Sprintf("%s-%s-%s", studentPackage.StudentId, course, locationID)
						if val, ok := s.ExpectedStudentCourse[keyStr]; ok {
							_ = val.DeletedAt.Set(time.Now())
						}
					}
				}
				studentPackage.ActionKind = npb.ActionKind_ACTION_KIND_DELETED
			}
		}
	}
	return ctx, nil
}

func (s *JprepSyncStudentPackageSuite) natsEventsArePublished(ctx context.Context) (context.Context, error) {
	for _, event := range s.EventUpserts {
		data, err := proto.Marshal(event)
		if err != nil {
			return ctx, fmt.Errorf("failed marshal event data: %v", err)
		}
		_, err = s.JSM.PublishContext(ctx, constants.SubjectSyncStudentPackage, data)
		if err != nil {
			return ctx, fmt.Errorf("failed pushlish SubjectSyncStudentPackage: %v", err)
		}
	}
	return ctx, nil
}

func (s *JprepSyncStudentPackageSuite) notificationSystemMustStoreStudentCourseDataCorrectly(ctx context.Context, typeEvent string) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	orgID := strconv.Itoa(int(stepState.Organization.ID))
	e := &entities.NotificationStudentCourse{}
	query := fmt.Sprintf(`
		SELECT %s
		FROM notification_student_courses nsc
		WHERE nsc.resource_path = $1 AND nsc.course_id = ANY($2::TEXT[])
			AND nsc.location_id = $3
			AND nsc.student_id = ANY($4::TEXT[])
	`, strings.Join(database.GetFieldNames(e), ","))

	if typeEvent == "upsert" {
		query += " AND deleted_at IS NULL"
	} else {
		query += " AND deleted_at IS NOT NULL"
	}

	err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(350 * time.Millisecond)
		rows, err := s.BobDBConn.Query(ctx, query,
			database.Text(orgID), database.TextArray(s.CourseIDs), database.Text(stepState.Organization.DefaultLocation.ID), database.TextArray(s.StudentIDs))
		if err != nil {
			return false, err
		}

		defer rows.Close()

		actualNotiStudentCourses := make([]*entities.NotificationStudentCourse, 0)
		for rows.Next() {
			item := &entities.NotificationStudentCourse{}
			err = rows.Scan(database.GetScanFields(item, database.GetFieldNames(item))...)
			if err != nil {
				return false, fmt.Errorf("err scan: %v", err)
			}
			actualNotiStudentCourses = append(actualNotiStudentCourses, item)
		}

		if err := rows.Err(); err != nil {
			return false, err
		}

		if len(actualNotiStudentCourses) == 0 {
			retry := attempt < 10

			if retry {
				time.Sleep(2 * time.Second)

				return true, fmt.Errorf("sync jprep student course temporarily failed, retrying")
			}
			return false, fmt.Errorf("sync or query failed")
		}

		if len(actualNotiStudentCourses) != len(s.ExpectedStudentCourse) {
			return false, fmt.Errorf("error sync data JPREP student course: expected %d, actual %d", len(s.ExpectedStudentCourse), len(actualNotiStudentCourses))
		}

		aDay := time.Duration(24 * time.Hour)
		for _, item := range actualNotiStudentCourses {
			keyStr := fmt.Sprintf("%s-%s-%s", item.StudentID.String, item.CourseID.String, item.LocationID.String)
			if expected, ok := s.ExpectedStudentCourse[keyStr]; ok {
				if expected.DeletedAt.Status != item.DeletedAt.Status {
					return false, fmt.Errorf("error sync data JPREP student course: deleted_at diff(%s, %s)", expected.DeletedAt.Time, item.DeletedAt.Time)
				}
				if expected.EndAt.Status != item.EndAt.Status ||
					(expected.EndAt.Time.Sub(item.EndAt.Time) != 0 && expected.EndAt.Time.Sub(item.EndAt.Time) != aDay) {
					return false, fmt.Errorf("error sync data JPREP student course: end_at diff(%s, %s)", expected.EndAt.Time, item.EndAt.Time)
				}
			} else {
				return false, fmt.Errorf("expect to find %s", keyStr)
			}
		}

		return false, nil
	})

	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
