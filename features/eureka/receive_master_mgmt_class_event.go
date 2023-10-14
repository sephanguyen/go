package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	master_mgmt_proto "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) aValidEvent_JoinMasterMgmtClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, event := s.aEvtJoinMasterMgmtClass(ctx, fmt.Sprintf("userId-%v", rand.Int31n(200)), fmt.Sprintf("classId-%v", rand.Int31n(200)))
	stepState.Event = event
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidEvent_LeaveMasterMgmtClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userId := fmt.Sprintf("userId-%v", rand.Int31n(200))
	classId := fmt.Sprintf("classId-%v", rand.Int31n(200))
	ctx, event := s.aEvtJoinMasterMgmtClass(ctx, userId, classId)
	stepState.Event = event
	if ctx, err := s.sendEventToNatsJS(ctx, "MasterMgmtClassEvent", constants.SubjectMasterMgmtClassUpserted); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, event = s.aEvtLeaveMasterMgmtClass(ctx, userId, classId)
	stepState.Event = event
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidEvent_CreateCourseMasterMgmtClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	cc := &master_mgmt_proto.EvtClass_CreateClass{
		CourseId:   fmt.Sprintf("courseId-%v", rand.Int31n(200)),
		ClassId:    fmt.Sprintf("classId-%v", rand.Int31n(200)),
		Name:       fmt.Sprintf("name-%v", rand.Int31n(200)),
		LocationId: fmt.Sprintf("location-%v", rand.Int31n(200)),
	}
	ctx, event := s.aEvtCreateCourseMasterMgmtClass(ctx, cc)
	stepState.Event = event
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aEvtCreateCourseMasterMgmtClass(ctx context.Context, courseClass *master_mgmt_proto.EvtClass_CreateClass) (context.Context, *master_mgmt_proto.EvtClass) {
	stepState := StepStateFromContext(ctx)
	event := &master_mgmt_proto.EvtClass{
		Message: &master_mgmt_proto.EvtClass_CreateClass_{
			CreateClass: courseClass,
		},
	}
	return StepStateToContext(ctx, stepState), event
}

func (s *suite) aEvtJoinMasterMgmtClass(ctx context.Context, userId, classId string) (context.Context, *master_mgmt_proto.EvtClass) {
	stepState := StepStateFromContext(ctx)
	event := &master_mgmt_proto.EvtClass{
		Message: &master_mgmt_proto.EvtClass_JoinClass_{
			JoinClass: &master_mgmt_proto.EvtClass_JoinClass{
				UserId:  userId,
				ClassId: classId,
			},
		},
	}
	return StepStateToContext(ctx, stepState), event
}

func (s *suite) aEvtLeaveMasterMgmtClass(ctx context.Context, userId, classId string) (context.Context, *master_mgmt_proto.EvtClass) {
	stepState := StepStateFromContext(ctx)
	event := &master_mgmt_proto.EvtClass{
		Message: &master_mgmt_proto.EvtClass_LeaveClass_{
			LeaveClass: &master_mgmt_proto.EvtClass_LeaveClass{
				UserId:  userId,
				ClassId: classId,
			},
		},
	}
	return StepStateToContext(ctx, stepState), event
}

func (s *suite) eurekaMustUpdateMasterMgmtClassMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var query string
	var err error
	var count int
	// sleep for consumer
	time.Sleep(1 * time.Second)

	switch v := stepState.Event.(*master_mgmt_proto.EvtClass).Message.(type) {
	case *master_mgmt_proto.EvtClass_CreateClass_:
		query = "SELECT count(*) FROM course_classes WHERE course_id = ANY($1) AND class_id = ANY($2)"
		err = s.DB.QueryRow(ctx, query, []string{v.CreateClass.CourseId}, []string{v.CreateClass.ClassId}).Scan(&count)
	case *master_mgmt_proto.EvtClass_JoinClass_:
		query = "SELECT count(*) FROM class_students WHERE student_id = ANY($1) AND class_id = ANY($2)"
		err = s.DB.QueryRow(ctx, query, []string{v.JoinClass.UserId}, []string{v.JoinClass.ClassId}).Scan(&count)
	case *master_mgmt_proto.EvtClass_LeaveClass_:
		query = "SELECT count(*) FROM class_students WHERE student_id = ANY($1) AND class_id = ANY($2) AND deleted_at is not null"
		err = s.DB.QueryRow(ctx, query, []string{v.LeaveClass.UserId}, []string{v.LeaveClass.ClassId}).Scan(&count)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update master mgmt class member not correct")
	}

	return StepStateToContext(ctx, stepState), nil
}
