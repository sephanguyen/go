package entryexitmgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"google.golang.org/protobuf/proto"
)

func (s *suite) aEvtUserWithMessage(ctx context.Context, event string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SchoolID = strconv.Itoa(int(stepState.CurrentSchoolID))
	if stepState.StudentID == "" {
		stepState.StudentID = idutil.ULIDNow()
		_, err := s.aValidUser(StepStateToContext(ctx, stepState), s.BobDBTrace, withID(stepState.StudentID), withUserGroup(cpb.UserGroup_USER_GROUP_STUDENT.String()), withRole(userConstant.RoleStudent))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		time.Sleep(1 * time.Second) // wait for kafka sync
	}

	if event == "CreateStudent" {
		stepState.Request = &bpb.EvtUser{
			Message: &bpb.EvtUser_CreateStudent_{
				CreateStudent: &bpb.EvtUser_CreateStudent{
					StudentId:   stepState.StudentID,
					StudentName: stepState.StudentName,
					SchoolId:    stepState.SchoolID,
				},
			},
		}
	}
	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) yasuoSendEventEvtUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	msg := stepState.Request.(*bpb.EvtUser)
	data, err := proto.Marshal(msg)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = s.JSM.TracedPublish(contextWithToken(ctx), "nats.TracedPublish", constants.SubjectUserCreated, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentMustHaveQrcode(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*bpb.EvtUser)
	studentEvent := req.GetCreateStudent()

	if err := try.Do(func(attempt int) (bool, error) {
		query := `SELECT count(student_id) FROM student_qr s WHERE s.student_id = $1`
		var count int

		if err := s.EntryExitMgmtDBTrace.QueryRow(ctx, query, studentEvent.StudentId).Scan(&count); err != nil {
			return false, err
		}
		if count == 1 {
			return false, nil
		}
		if count > 1 {
			return false, fmt.Errorf("unexpected %d qrcode created for student", count)
		}
		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("qrcode not created")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentMustHaveNoQrcode(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*bpb.EvtUser)
	studentEvent := req.GetCreateStudent()

	if err := try.Do(func(attempt int) (bool, error) {
		query := `SELECT count(student_id) FROM student_qr s WHERE s.student_id = $1`
		var count int

		if err := s.EntryExitMgmtDBTrace.QueryRow(ctx, query, studentEvent.StudentId).Scan(&count); err != nil {
			return false, err
		}

		if count == 0 {
			return false, nil
		}

		if count >= 1 {
			return false, fmt.Errorf("unexpected %d qrcode created for student", count)
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("qrcode has been created for this student: %v", studentEvent.StudentId)
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
