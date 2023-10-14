package common

import (
	"context"

	"github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/features/communication/common/helpers"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
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

	Schools        []*entities.School
	CurrentSchools []*entities.School

	// system notifications
	PayloadSystemNotifications []*payload.UpsertSystemNotification
	TokenOfSentRecipient       string

	MapCourseIDAndStudentIDs map[string][]string
	MapStudentIDAndParentIDs map[string][]string

	Notification *cpb.Notification

	MultiTenants []*context.Context

	ClientID string

	NatsNotification *ypb.NatsCreateNotificationRequest

	Questionnaire         *cpb.Questionnaire
	QuestionnaireTemplate *npb.QuestionnaireTemplate
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

type NotificationSuite struct {
	*StepState
	*helpers.CommunicationHelper
}
