package virtualclassroom

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) anExistingUserSigninSystem(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		err               error
		userID, userGroup string
	)

	switch user {
	case studentType:
		if len(stepState.StudentIds) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("must have at least one existing student account")
		}
		userID = stepState.StudentIds[0]
		userGroup = constant.UserGroupStudent
		stepState.CurrentStudentID = userID
	case teacherType:
		if len(stepState.TeacherIDs) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("must have at least one existing teacher account")
		}
		userID = stepState.TeacherIDs[0]
		userGroup = constant.UserGroupTeacher
		stepState.CurrentTeacherID = userID
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("user %s is not supported on this step", user)
	}

	stepState.AuthToken, err = s.CommonSuite.GenerateExchangeToken(userID, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = userID
	stepState.CurrentUserGroup = userGroup

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) useAlreadyHasExistingPrivateGroupOfConversationsWithUsers(ctx context.Context, groupCount, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var participants []string

	count, err := strconv.Atoi(groupCount)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert group count to a number: %w", err)
	}

	switch user {
	case studentType:
		if len(stepState.StudentIds) < count {
			return StepStateToContext(ctx, stepState), fmt.Errorf("must have at least %d existing student accounts", count)
		}
		participants = stepState.StudentIds
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("user %s is not supported on this step", user)
	}
	currentUser := stepState.CurrentUserID
	totalConversations := int(math.Floor(float64(len(participants)) / float64(count)))

	// create with the specified number of users per group
	// ex. groupCount = 2; total participants = 3 (to include the current user)
	for i := 0; i < totalConversations; i++ {
		currentIndex := count * i
		otherParticipants := participants[currentIndex : currentIndex+count]

		reqParticipants := []string{currentUser}
		reqParticipants = append(reqParticipants, otherParticipants...)

		tempCtx, err := s.userGetsConversationID(ctx, stepState.CurrentLessonID, reqParticipants, vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PRIVATE)
		tempStepState := StepStateFromContext(tempCtx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed in creating conversations in advance: %w", err)
		}
		if tempStepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed in creating conversations in advance: %w", tempStepState.ResponseErr)
		}
	}

	// create with many users
	return s.userGetsConversationID(ctx, stepState.CurrentLessonID, participants, vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PRIVATE)
}

func (s *suite) userGetsConversationID(ctx context.Context, lessonID string, participants []string, convType vpb.LiveLessonConversationType) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetConversationIDRequest{
		LessonId:         lessonID,
		ParticipantList:  participants,
		ConversationType: convType,
	}
	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomChatServiceClient(s.VirtualClassroomConn).
		GetConversationID(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsPrivateConversationIDwithAUser(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var participants []string

	switch user {
	case studentType:
		if len(stepState.StudentIds) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("must have at least one existing student account")
		}
		participants = append(participants, stepState.StudentIds[0])
	case teacherType:
		if len(stepState.TeacherIDs) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("must have at least one existing teacher account")
		}
		participants = append(participants, stepState.TeacherIDs[0])
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("user %s is not supported on this step", user)
	}

	return s.userGetsConversationID(ctx, stepState.CurrentLessonID, participants, vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PRIVATE)
}

func (s *suite) userGetsTheSamePrivateConversationIDwithAUser(ctx context.Context, user string) (context.Context, error) {
	return s.userGetsPrivateConversationIDwithAUser(ctx, user)
}

func (s *suite) userGetsPublicConversationID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	participants := []string{stepState.CurrentUserID}

	return s.userGetsConversationID(ctx, stepState.CurrentLessonID, participants, vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PUBLIC)
}

func (s *suite) userGetsTheSamePublicConversationID(ctx context.Context) (context.Context, error) {
	return s.userGetsPublicConversationID(ctx)
}

func (s *suite) userGetsNonEmptyConversationID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetConversationIDResponse)
	conversationID := response.GetConversationId()

	if len(conversationID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting conversation ID but got empty")
	}

	request := stepState.Request.(*vpb.GetConversationIDRequest)

	stepState.LiveLessonConversations = append(stepState.LiveLessonConversations,
		domain.LiveLessonConversation{
			ConversationID:   conversationID,
			LessonID:         request.GetLessonId(),
			ParticipantList:  request.GetParticipantList(),
			ConversationType: domain.LiveLessonConversationType(request.GetConversationType().String()),
		},
	)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedPrivateConversationID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetConversationIDResponse)
	actualConversationID := response.GetConversationId()

	if len(actualConversationID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting conversation ID but got empty")
	}

	request := stepState.Request.(*vpb.GetConversationIDRequest)
	actualLessonID := request.GetLessonId()
	actualParticiapnts := request.GetParticipantList()
	actualConversationType := domain.LiveLessonConversationType(request.GetConversationType().String())

	for _, con := range stepState.LiveLessonConversations {
		if con.LessonID == actualLessonID && con.ConversationType == actualConversationType && sliceutils.UnorderedEqual(actualParticiapnts, con.ParticipantList) {
			if con.ConversationID != actualConversationID {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the expected %s conversation ID %s for participants %s lesson %s returned a different conversation ID %s",
					con.ConversationType,
					con.ConversationID,
					con.ParticipantList,
					con.LessonID,
					actualConversationID,
				)
			}
			break // found match
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedPublicConversationID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetConversationIDResponse)
	actualConversationID := response.GetConversationId()

	if len(actualConversationID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting conversation ID but got empty")
	}

	request := stepState.Request.(*vpb.GetConversationIDRequest)
	actualLessonID := request.GetLessonId()
	actualConversationType := domain.LiveLessonConversationType(request.GetConversationType().String())

	for _, con := range stepState.LiveLessonConversations {
		if con.LessonID == actualLessonID && con.ConversationType == actualConversationType {
			if con.ConversationID != actualConversationID {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the expected %s conversation ID %s for lesson %s returned a different conversation ID %s",
					con.ConversationType,
					con.ConversationID,
					con.LessonID,
					actualConversationID,
				)
			}
			break // found match
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
