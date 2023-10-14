package usermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) schoolAdminReallocateTheStudentsEnrollmentStatus(ctx context.Context, dataType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	orgID := OrgIDFromCtx(ctx)
	student := stepState.Response.(*upb.UpsertStudentResponse)
	evt := reallocateStudentEnrollmentStatusReq(student.StudentProfiles[0].Id, getChildrenLocation(orgID)[0])

	switch dataType {
	case "student does not existed":
		evt.StudentEnrollmentStatus[0].StudentId = idutil.ULIDNow()
	case "location does not existed":
		evt.StudentEnrollmentStatus[0].LocationId = idutil.ULIDNow()
	case "enrollment status is not temporary":
		evt.StudentEnrollmentStatus[0].EnrollmentStatus = 2
	case "start date after end date":
		evt.StudentEnrollmentStatus[0].EndDate = timestamppb.New(time.Now().Add(-24 * time.Hour))
	case "temporary status":
		evt.StudentEnrollmentStatus[0].EnrollmentStatus = npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return ctx, errors.Wrap(err, "json.Marshal failed")
	}
	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectEnrollmentStatusAssignmentCreated, data)
	if err != nil {
		return ctx, nats.HandlePushMsgFail(ctx, fmt.Errorf("reallocateStudentEnrollmentStatus: JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	stepState.Request1 = evt

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsEnrollmentStatusWasReallocatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request1.(*npb.LessonReallocateStudentEnrollmentStatusEvent)
	studentID := req.StudentEnrollmentStatus[0].StudentId
	locationID := req.StudentEnrollmentStatus[0].LocationId
	studentEnrollmentStatus, err := tryToGetStudentEnrollmentStatus(ctx, s.BobDBTrace, studentID, locationID)
	if err != nil {
		return ctx, err
	}
	if len(studentEnrollmentStatus) == 0 {
		return ctx, fmt.Errorf("reallocate student enrollment status failed")
	}

	switch ses := studentEnrollmentStatus[0]; {
	case !ses.StartDate().Time().Round(time.Second).Equal(req.StudentEnrollmentStatus[0].StartDate.AsTime().Round(time.Second)):
		return ctx, fmt.Errorf("expected start date must be equal with request")
	case !ses.EndDate().Time().Round(time.Second).Equal(req.StudentEnrollmentStatus[0].EndDate.AsTime().Round(time.Second)):
		return ctx, fmt.Errorf("expected end date date must be equal with request")
	case ses.EnrollmentStatus().String() != upb.StudentEnrollmentStatus_name[int32(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY)]:
		return ctx, fmt.Errorf("expected enrollment status must be temporary, but got %s", ses.EnrollmentStatus().String())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminCanNotReallocateTheStudentsEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request1.(*npb.LessonReallocateStudentEnrollmentStatusEvent)
	studentID := req.StudentEnrollmentStatus[0].StudentId
	locationID := req.StudentEnrollmentStatus[0].LocationId
	studentEnrollmentStatus, err := tryToGetStudentEnrollmentStatus(ctx, s.BobDBTrace, studentID, locationID)
	if err != nil {
		return ctx, err
	}
	if len(studentEnrollmentStatus) > 0 {
		return ctx, fmt.Errorf("expected can not reallocate student enrollment status")
	}

	return StepStateToContext(ctx, stepState), nil
}

func tryToGetStudentEnrollmentStatus(ctx context.Context, db database.Ext, studentID, locationID string) ([]entity.DomainEnrollmentStatusHistory, error) {
	studentEnrollmentStatus := entity.DomainEnrollmentStatusHistories{}
	err := try.Do(func(attempt int) (bool, error) {
		students, err := (&repository.DomainEnrollmentStatusHistoryRepo{}).GetByStudentIDAndLocationID(ctx, db, studentID, locationID, false)
		if err == nil && len(students) > 0 {
			studentEnrollmentStatus = students
			return false, nil
		}

		if attempt < retryTimes {
			time.Sleep(200 * time.Millisecond)
			return true, fmt.Errorf("can't find student enrollment status")
		}

		return false, err
	})

	return studentEnrollmentStatus, err
}

func reallocateStudentEnrollmentStatusReq(studentID string, locationID string) *npb.LessonReallocateStudentEnrollmentStatusEvent {
	return &npb.LessonReallocateStudentEnrollmentStatusEvent{
		StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
			{
				StudentId:        studentID,
				LocationId:       locationID,
				StartDate:        timestamppb.Now(),
				EndDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
				EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
			},
		},
	}
}
