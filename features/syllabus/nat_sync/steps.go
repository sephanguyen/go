package nat_sync

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/nats-io/nats.go"
	"go.uber.org/multierr"
)

type StepState struct {
	Token         string
	Response      interface{}
	Request       interface{}
	ResponseErr   error
	CourseClass   []*entities.CourseClass
	CourseStudent []*entities.CourseStudent
	StudentIDs    []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<jprep_sync>a signed in "([^"]*)"$`:                               s.aSignedIn,
		`^<jprep_sync_course_class>generate course class$`:                  s.generateCoureClass,
		`^<jprep_sync_course_class>NAT JETSTREAM send a request$`:           s.NatSendARequestSyncCourseClass,
		`^<jprep_sync_course_class> store correct request to our system$`:   s.StoreCorrectResultFromSyncCourseClassRequest,
		`^<jprep_sync_course_student>generate course student$`:              s.generateCoureStudent,
		`^<jprep_sync_course_student>NAT JETSTREAM send a request$`:         s.NatSendARequestSyncCourseStudentV2,
		`^<jprep_sync_course_student> store correct request to our system$`: s.StoreCorrectResultFromSyncCourseStudentRequest,
	}

	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	_, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, arg)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.Token = authToken
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsStatusCode(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), utils.ValidateStatusCode(stepState.ResponseErr, arg)
}

func (s *Suite) generateCoureClass(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	courseClasses := []*entities.CourseClass{
		{ID: database.Text("course-1-111"), CourseID: database.Text("course-1"), ClassID: database.Text("111")},
		{ID: database.Text("course-2-222"), CourseID: database.Text("course-2"), ClassID: database.Text("222")},
		{ID: database.Text("course-3-333"), CourseID: database.Text("course-3"), ClassID: database.Text("333")},
	}

	stepState.CourseClass = courseClasses
	repo := &repositories.CourseClassRepo{}
	if err := repo.BulkUpsert(ctx, s.EurekaDB, courseClasses); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) generateCoureStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	courseStudentMap := map[string][]string{}
	stepState.StudentIDs = []string{idutil.ULIDNow(), idutil.ULIDNow()}
	courseStudentMap[stepState.StudentIDs[0]] = []string{"course-1", "course-2"}
	courseStudentMap[stepState.StudentIDs[1]] = []string{"course-3", "course-4"}
	courseStudents := []*entities.CourseStudent{}
	for studentID, courseIDs := range courseStudentMap {
		for _, courseID := range courseIDs {
			courseStudent := &entities.CourseStudent{}
			database.AllNullEntity(courseStudent)
			err := multierr.Combine(
				courseStudent.ID.Set(idutil.ULIDNow()),
				courseStudent.StudentID.Set(studentID),
				courseStudent.CourseID.Set(courseID),
				courseStudent.CreatedAt.Set(timeutil.Now()),
				courseStudent.UpdatedAt.Set(timeutil.Now()),
				courseStudent.StartAt.Set(time.Now().Add(-24*time.Hour)),
				courseStudent.EndAt.Set(time.Now().AddDate(0, 0, 10)),
			)
			if err != nil {
				return nil, fmt.Errorf("err set CourseStudent: %w", err)
			}
			courseStudents = append(courseStudents, courseStudent)
		}
	}

	stepState.CourseStudent = courseStudents
	repo := &repositories.CourseStudentRepo{}
	if _, err := repo.BulkUpsert(ctx, s.EurekaDB, courseStudents); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

const (
	JPREPCourseStudentStreamName     = constants.StreamSyncStudentPackage
	JPREPCourseStudentStreamSubjects = constants.DeliverSyncStudentPackageEureka
)

func (s *Suite) CreateStream(name string, subject string) error {
	stream, err := s.JSM.GetJS().StreamInfo(name)

	if err != nil {
		return err
	}

	if stream == nil {
		_, err = s.JSM.GetJS().AddStream(&nats.StreamConfig{
			Name:     name,
			Subjects: []string{subject},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Suite) PublishMessage(ctx context.Context, data []byte, subjectName string) error {
	_, err := s.JSM.PublishAsyncContext(ctx, subjectName, data)
	if err != nil {
		return fmt.Errorf("publishMessage error: %w", err)
	}
	return nil
}
