package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *suite) removeParentSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleRemoveParent := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &pb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *pb.RemoveParentFromStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *pb.EvtUser_ParentRemovedFromStudent_:
				if req.StudentId == msg.ParentRemovedFromStudent.StudentId && req.ParentId == msg.ParentRemovedFromStudent.ParentId {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleRemoveParent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("removeParentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentsDataToRemove(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.addParentDataToRemoveParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addParentDataToRemoveParentReq(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.createNewParents(ctx, schoolAdminType); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	parentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles[0].Parent.UserProfile.UserId
	studentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).StudentId

	stepState.ParentIDs = append(stepState.ParentIDs, parentID)

	stepState.Request = &pb.RemoveParentFromStudentRequest{
		StudentId: studentID,
		ParentId:  parentID,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentsWereRemoveSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	if err := s.validateRemoveParent(ctx); err != nil {
		return ctx, fmt.Errorf("validateRemoveParent: %s", err.Error())
	}

	req := stepState.Request.(*pb.RemoveParentFromStudentRequest)
	if err := s.validParentAccessPath(ctx, req.StudentId); err != nil {
		return ctx, errors.Wrap(err, "validParentAccessPath")
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return StepStateToContext(ctx, stepState), nil
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) validateRemoveParent(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.RemoveParentFromStudentRequest)
	const query = `SELECT COUNT(*) 
FROM student_parents 
WHERE student_id = $1 AND parent_id = $2 AND deleted_at IS NULL
`
	row := s.BobDBTrace.QueryRow(ctx, query, req.StudentId, req.ParentId)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("remove parent not success")
	}
	return nil
}

func (s *suite) parentsDataWithoutRelationshipToRemove(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.createNewParents(ctx, schoolAdminType); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	parentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles[0].Parent.UserProfile.UserId
	stepState.ParentIDs = append(stepState.ParentIDs, parentID)

	err = s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.createNewParents(ctx, schoolAdminType); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	studentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).StudentId

	stepState.Request = &pb.RemoveParentFromStudentRequest{
		StudentId: studentID,
		ParentId:  parentID,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createDataParentWithMultipleStudents(ctx context.Context, numberStudents int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// Create new parent
	err := s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.createNewParents(ctx, schoolAdminType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	parentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles[0].Parent.UserProfile.UserId
	studentID := ""

	for i := 0; i < numberStudents; i++ {
		student, err := s.createStudent(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentID = student.ID.String
		// Assign created parent to diff students
		updateParentReq := updateParentReq(student.ID.String, parentID)
		stepState.Request = updateParentReq
		ctx, err = s.updateNewParents(ctx, schoolAdminType)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	stepState.Request = &pb.RemoveParentFromStudentRequest{
		StudentId: studentID,
		ParentId:  parentID,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) removeParentFromStudentWithConditions(ctx context.Context, account string, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	request := stepState.Request.(*pb.RemoveParentFromStudentRequest)
	stepState.RequestSentAt = time.Now()
	switch condition {
	case "invalid empty studentID":
		request = &pb.RemoveParentFromStudentRequest{StudentId: "", ParentId: stepState.Request.(*pb.RemoveParentFromStudentRequest).ParentId}
	case "invalid empty parentID":
		request = &pb.RemoveParentFromStudentRequest{StudentId: stepState.Request.(*pb.RemoveParentFromStudentRequest).StudentId, ParentId: ""}
	case "invalid un-exist parentID":
		request = &pb.RemoveParentFromStudentRequest{StudentId: stepState.Request.(*pb.RemoveParentFromStudentRequest).StudentId, ParentId: "UnExist"}
	case "invalid un-exist studentID":
		request = &pb.RemoveParentFromStudentRequest{StudentId: "UnExist", ParentId: stepState.Request.(*pb.RemoveParentFromStudentRequest).ParentId}
	}
	ctx, err = s.removeParentSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.removeParentSubscription: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).RemoveParentFromStudent(contextWithToken(ctx), request)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receiveCodeStatusAndMessage(ctx context.Context, codeStatus string, message string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != codeStatus {
		return ctx, fmt.Errorf("expecting %s, got %s status code", codeStatus, stt.Code().String())
	}

	if stt.Message() != message {
		return ctx, fmt.Errorf("expecting %s, got message: %s", message, stt.Message())
	}
	return ctx, nil
}

func (s *suite) verifyParentInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.RemoveParentFromStudentRequest)

	studentParentRepo := repository.StudentParentRepo{}

	studentParentsDB, err := studentParentRepo.FindStudentParentsByParentID(ctx, s.BobDB, req.ParentId)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentParentRepo.FindStudentParentsByParentID: %w", err)
	}

	if len(studentParentsDB) == 0 {
		return ctx, fmt.Errorf("expecting studentParent in database > 0, got: 0")
	}

	for _, studentParentDB := range studentParentsDB {
		if studentParentDB.StudentID.String != req.StudentId {
			return ctx, fmt.Errorf("expecting StudentID in database: %s, got: %s", req.StudentId, studentParentDB.StudentID.String)
		}
	}
	return ctx, nil
}
