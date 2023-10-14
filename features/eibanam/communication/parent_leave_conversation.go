package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/internal/golibs/try"
)

type ParentLeaveConversationSuite struct {
	*SupportChatSuite
}

func (s *ParentLeaveConversationSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^"([^"]*)" logins CMS$`:                                               s.LoginCms,
		`^"([^"]*)" has created "([^"]*)" with "([^"]*)" info$`:                s.CreateStudentWithParentInfo,
		`^"([^"]*)" has joined "([^"]*)" chat group$`:                          s.JoinedChatGroupAndSubscribeStream,
		`^"([^"]*)" has sent "([^"]*)" message to parent chat group$`:          s.SendMessageToParentChatGroup,
		`^"([^"]*)" logins on Learner App$`:                                    s.LoginLeanerApp,
		`^school admin removes the relationship of "([^"]*)" and "([^"]*)"$`:   s.SchoolAdminRemovesTheRelationshipOfStudentParent,
		`^"([^"]*)" sees "([^"]*)" message of "([^"]*)" with name and avatar$`: s.SeesMessageFromWithNameAndAvatar,
		`^"([^"]*)" sees parent chat group is removed on Learner App$`:         s.ParentSeesParentChatGroupIsRemovedOnLearnerApp,
		`^teacher sees "([^"]*)" message of "([^"]*)" with name and avatar$`:   s.TeacherSeesMessageFromWithNameAndAvatar,
		`^school admin add the relationship of "([^"]*)" and "([^"]*)"$`:       s.SchoolAdminAddsTheRelationshipOfStudentParent,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func NewParentLeaveConversationSuite(util *helper.CommunicationHelper) *ParentLeaveConversationSuite {
	support := &SupportChatSuite{
		util: util,
	}
	return &ParentLeaveConversationSuite{
		support,
	}
}

func (s *ParentLeaveConversationSuite) ParentSeesParentChatGroupIsRemovedOnLearnerApp(ctx context.Context, parname string) (context.Context, error) {
	st := s.SupportChatSuite.FromCtx(ctx)
	par := st.getTargetUserState(parname)
	parchat := st.getTargetState(parname)

	err := try.Do(func(attempt int) (bool, error) {
		chatlist, err := s.util.GetLearnerAppChat(ctx, par.Token)
		if err != nil {
			return false, err
		}
		parchat.LearnerAppChats = chatlist
		curstudent := st.getTargetUserState("student").ID
		for _, chat := range chatlist.Conversations {
			if chat.StudentId == curstudent {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("parent %s still see student chat group", parname)
			}
		}
		return false, nil
	})

	return s.SupportChatSuite.ToCtx(ctx, st), err
}
