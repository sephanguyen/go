package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/try"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aValidEvent_JoinClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, event := s.aEvtJoinClass(ctx, "userId-join", rand.Int31n(100))
	stepState.Event = event
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidEvent_LeaveClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userIds := []string{"userId", "userId1", "userId2"}
	classId := rand.Int31n(200)
	for _, item := range userIds {
		ctx, event := s.aEvtJoinClass(ctx, item, classId)
		stepState.Event = event
		if ctx, err := s.sendEventToNatsJS(ctx, "ClassEvent", constants.SubjectClassUpserted); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	ctx, event := s.aEvtLeaveClass(ctx, userIds, classId)
	stepState.Event = event
	if ctx, err := s.sendEventToNatsJS(ctx, "ClassEvent", constants.SubjectClassUpserted); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aEvtJoinClass(ctx context.Context, userId string, classId int32) (context.Context, *bobproto.EvtClassRoom) {
	stepState := StepStateFromContext(ctx)
	event := &bobproto.EvtClassRoom{
		Message: &bobproto.EvtClassRoom_JoinClass_{
			JoinClass: &bobproto.EvtClassRoom_JoinClass{
				UserId:  userId,
				ClassId: classId,
			},
		},
	}
	return StepStateToContext(ctx, stepState), event
}

func (s *suite) aEvtLeaveClass(ctx context.Context, userIds []string, classId int32) (context.Context, *bobproto.EvtClassRoom) {
	stepState := StepStateFromContext(ctx)
	event := &bobproto.EvtClassRoom{
		Message: &bobproto.EvtClassRoom_LeaveClass_{
			LeaveClass: &bobproto.EvtClassRoom_LeaveClass{
				UserIds: userIds,
				ClassId: classId,
			},
		},
	}
	return StepStateToContext(ctx, stepState), event
}

func (s *suite) eurekaMustUpsertClassMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	event := stepState.Event.(*bobproto.EvtClassRoom).Message.(*bobproto.EvtClassRoom_JoinClass_)
	count := 0
	query := fmt.Sprintf("SELECT count(*) FROM class_students WHERE student_id = ANY($1) AND class_id = ANY($2)")
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(250 * time.Millisecond)

		err = s.DB.QueryRow(ctx, query, []string{event.JoinClass.UserId}, []string{strconv.Itoa(int(event.JoinClass.ClassId))}).Scan(&count)
		if err != nil {
			return true, err
		}
		if count != 1 {
			return true, fmt.Errorf("Eureka does not create class member correctly")
		}
		return attempt < 5, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustUpdateClassMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	event := stepState.Event.(*bobproto.EvtClassRoom).Message.(*bobproto.EvtClassRoom_LeaveClass_).LeaveClass
	count := 0
	query := fmt.Sprintf("SELECT count(*) FROM class_members WHERE course_id = ANY($1) AND class_id = ANY($2) AND deleted_at is not null")
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(350 * time.Microsecond)
		err = s.DB.QueryRow(ctx, query, event.UserIds, []string{strconv.Itoa(int(event.GetClassId()))}).Scan(&count)
		if err != nil {
			return true, nil
		}
		if count != len(event.UserIds) {
			return true, fmt.Errorf("Eureka does not update class member correctly")
		}
		return attempt < 5, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
