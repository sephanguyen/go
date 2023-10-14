package communication

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/cucumber/godog"
)

type JprepSyncStudentClassSuite struct {
	*common.NotificationSuite
	// EventUpserts          []*npb.EventSyncStudentPackage
	EventUpserts        []*pb.EvtClassRoom
	ExpectedClassMember map[string]*entities.NotificationClassMember
	StudentIDs          []string
	CourseIDs           []string
	ClassIDs            []string
	mapCourseAndClass   map[string][]int32
	mapClassAndCourse   map[int32]string
}

func (c *SuiteConstructor) InitJprepSyncStudentClass(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &JprepSyncStudentClassSuite{
		NotificationSuite:   dep.notiCommonSuite,
		EventUpserts:        make([]*pb.EvtClassRoom, 0),
		ExpectedClassMember: make(map[string]*entities.NotificationClassMember, 0),
		mapCourseAndClass:   make(map[string][]int32, 0),
		mapClassAndCourse:   make(map[int32]string, 0),
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" of Jprep organization with default location logged in Back Office$`: s.StaffGrantedRoleOfJprepOrganizationWithDefaultLocationLoggedInBackOffice,
		`^school admin creates "([^"]*)" students$`:                                           s.CreatesNumberOfStudents,
		`^school admin creates "([^"]*)" courses with "([^"]*)" classes$`:                     s.CreatesNumberOfCoursesWithClass,
		`^nats events are published$`:                                                         s.natsEventsArePublished,
		`^notification system must store student class data correctly$`:                       s.notificationSystemMustStoreStudentClassDataCorrectly,
		`^some valid upsert event sync student class from Bob$`:                               s.someValidUpsertEventSyncStudentClassFromBob,
		`^school admin "([^"]*)" some student class$`:                                         s.schoolAdminSomeStudentClass,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *JprepSyncStudentClassSuite) someValidUpsertEventSyncStudentClassFromBob(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	for _, course := range stepState.Courses {
		s.CourseIDs = append(s.CourseIDs, course.ID)
		for _, class := range course.Classes {
			intClassID, err := strconv.ParseInt(class.ID, 10, 32)
			if err != nil {
				return ctx, fmt.Errorf("failed convert ClassID str to int: %v", err)
			}
			s.mapCourseAndClass[course.ID] = append(s.mapCourseAndClass[course.ID], int32(intClassID))
			s.mapClassAndCourse[int32(intClassID)] = course.ID
			s.ClassIDs = append(s.ClassIDs, class.ID)
		}
	}
	locationID := stepState.Organization.DefaultLocation.ID
	for _, student := range stepState.Students {
		s.StudentIDs = append(s.StudentIDs, student.ID)
		for courseID, classIDs := range s.mapCourseAndClass {
			for _, classID := range classIDs {
				evt := &pb.EvtClassRoom{
					Message: &pb.EvtClassRoom_JoinClass_{
						JoinClass: &pb.EvtClassRoom_JoinClass{
							ClassId:   classID,
							UserId:    student.ID,
							UserGroup: pb.UserGroup(pb.UserGroup_value["USER_GROUP_STUDENT"]),
						},
					},
				}
				keyStr := fmt.Sprintf("%s-%d-%s-%s", student.ID, classID, locationID, courseID)
				item, err := mappers.EventJoinClassRoomToNotificationClassMemberEnt(evt.GetJoinClass(), courseID, locationID)
				if err != nil {
					return ctx, fmt.Errorf("multierr %v", err)
				}
				s.ExpectedClassMember[keyStr] = item
				s.EventUpserts = append(s.EventUpserts, evt)
			}
		}
	}
	return ctx, nil
}

func (s *JprepSyncStudentClassSuite) schoolAdminSomeStudentClass(ctx context.Context, action string) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	locationID := stepState.Organization.DefaultLocation.ID
	now := time.Now()
	for _, event := range s.EventUpserts {
		// nolint
		if rand.Intn(2) == 1 {
			msg := event.GetJoinClass()
			courseID := s.mapClassAndCourse[int32(msg.ClassId)]
			keyStr := fmt.Sprintf("%s-%d-%s-%s", msg.UserId, msg.ClassId, locationID, courseID)
			item, err := mappers.EventJoinClassRoomToNotificationClassMemberEnt(event.GetJoinClass(), courseID, locationID)
			if err != nil {
				return ctx, fmt.Errorf("multierr %v", err)
			}
			_ = item.DeletedAt.Set(now)
			s.ExpectedClassMember[keyStr] = item

			newMsg := &pb.EvtClassRoom_LeaveClass_{
				LeaveClass: &pb.EvtClassRoom_LeaveClass{
					ClassId:  msg.ClassId,
					UserIds:  []string{msg.UserId},
					IsKicked: false,
				},
			}
			event.Message = newMsg
		}
	}
	return ctx, nil
}

func (s *JprepSyncStudentClassSuite) natsEventsArePublished(ctx context.Context) (context.Context, error) {
	for _, event := range s.EventUpserts {
		data, err := event.Marshal()
		if err != nil {
			return ctx, fmt.Errorf("failed marshal event data: %v", err)
		}
		_, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectClassUpserted, data)
		if err != nil {
			return ctx, fmt.Errorf("failed pushlish SubjectClassUpserted: %v", err)
		}
	}
	return ctx, nil
}

func (s *JprepSyncStudentClassSuite) notificationSystemMustStoreStudentClassDataCorrectly(ctx context.Context) (context.Context, error) {
	time.Sleep(5 * time.Second)
	stepState := common.StepStateFromContext(ctx)
	orgID := strconv.Itoa(int(stepState.Organization.ID))
	e := &entities.NotificationClassMember{}
	query := fmt.Sprintf(`
		SELECT %s
		FROM notification_class_members ncm
		WHERE ncm.resource_path = $1 AND ncm.class_id = ANY($2::TEXT[])
		AND ncm.location_id = $3
		AND ncm.student_id = ANY($4::TEXT[])
	`, strings.Join(database.GetFieldNames(e), ","))
	err := try.Do(func(attempt int) (bool, error) {
		rows, err := s.BobDBConn.Query(ctx, query,
			database.Text(orgID), database.TextArray(s.ClassIDs), database.Text(stepState.Organization.DefaultLocation.ID), database.TextArray(s.StudentIDs))
		if err != nil {
			return false, err
		}

		defer rows.Close()

		actualNotiClassMembers := make([]*entities.NotificationClassMember, 0)
		for rows.Next() {
			item := &entities.NotificationClassMember{}
			err = rows.Scan(database.GetScanFields(item, database.GetFieldNames(item))...)
			if err != nil {
				return false, fmt.Errorf("err scan: %v", err)
			}
			actualNotiClassMembers = append(actualNotiClassMembers, item)
		}

		if err := rows.Err(); err != nil {
			return false, err
		}

		if len(actualNotiClassMembers) == 0 {
			retry := attempt < 10

			if retry {
				time.Sleep(2 * time.Second)

				return true, fmt.Errorf("sync jprep student course temporarily failed, retrying")
			}
			return false, fmt.Errorf("sync or query failed")
		}

		if len(actualNotiClassMembers) != len(s.ExpectedClassMember) {
			return false, fmt.Errorf("error sync data JPREP student course: expected %d, actual %d", len(s.ExpectedClassMember), len(actualNotiClassMembers))
		}

		for _, item := range actualNotiClassMembers {
			keyStr := fmt.Sprintf("%s-%s-%s-%s", item.StudentID.String, item.ClassID.String, item.LocationID.String, item.CourseID.String)
			if expected, ok := s.ExpectedClassMember[keyStr]; ok {
				if expected.DeletedAt.Status != item.DeletedAt.Status {
					return false, fmt.Errorf("error sync data JPREP student course: deleted_at diff(%s, %s)", expected.DeletedAt.Time, item.DeletedAt.Time)
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
