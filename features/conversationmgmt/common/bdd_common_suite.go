package common

import (
	"context"

	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	"github.com/manabie-com/backend/features/conversationmgmt/common/helpers"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
)

type StepState struct {
	Request     interface{}
	Response    interface{}
	ResponseErr error

	AuthToken             string
	CurrentUserGroup      string
	CurrentGrandtedRoles  []string
	CurrentUserID         string
	CurrentResourcePath   string
	CurrentOrganicationID int32

	CurrentStaff *entities.Staff
	Organization *entities.Organization

	Students       []*entities.Student
	Courses        []*entities.Course
	Classes        []*entities.Class
	Tags           []*entities.Tag
	GradeMasters   []*entities.GradeMaster
	GradeAssigneds []*entities.GradeMaster

	Conversations []*entities.Conversation

	Schools        []*entities.School
	CurrentSchools []*entities.School

	// system notifications
	PayloadSystemNotifications []*payload.UpsertSystemNotification
	TokenOfSentRecipient       string

	MapCourseIDAndStudentIDs map[string][]string
	MapStudentIDAndParentIDs map[string][]string

	// Notification *cpb.Notification

	MultiTenants []*context.Context

	ClientID string
}

type StepStateKey struct{}

func StepStateFromContext(ctx context.Context) *StepState {
	state := ctx.Value(StepStateKey{})
	if state == nil {
		return &StepState{}
	}
	return state.(*StepState)
}

func StepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, StepStateKey{}, state)
}

type ConversationMgmtSuite struct {
	*StepState
	*helpers.ConversationMgmtHelper
}
