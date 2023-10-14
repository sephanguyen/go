package conversationmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/conversationmgmt/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/cucumber/godog"
)

type GetAgoraAppInfoSuite struct {
	*common.ConversationMgmtSuite
	studentToken string
}

func (c *SuiteConstructor) InitGetAgoraAppInfo(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &GetAgoraAppInfoSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^current staff creates "([^"]*)" students$`: s.CreatesNumberOfStudents,
		`^student login to Learner App$`:             s.studentLoginToLearnerApp,
		`^student call GetAgoraInfo API$`:            s.studentCallGetAgoraInfoAPI,
		`^returns "([^"]*)" status code$`:            s.CheckReturnStatusCode,
		`^GetAgoraInfo API return correct data$`:     s.getAgoraInfoAPIReturnCorrectData,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *GetAgoraAppInfoSuite) studentLoginToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	studentID := commonState.Students[0].ID
	userGroup := commonState.Students[0].Group
	var err error
	s.studentToken, err = s.GenerateExchangeTokenCtx(ctx, studentID, userGroup)
	if err != nil {
		return ctx, fmt.Errorf("failed GenerateExchangeTokenCtx: %+v", err)
	}
	return ctx, nil
}

func (s *GetAgoraAppInfoSuite) studentCallGetAgoraInfoAPI(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	studentContextWithToken := common.ContextWithToken(ctx, s.studentToken)

	commonState.Response, commonState.ResponseErr = cpb.NewAgoraUserMgmtServiceClient(s.ConversationMgmtGRPCConn).GetAppInfo(
		studentContextWithToken,
		&cpb.GetAppInfoRequest{})

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *GetAgoraAppInfoSuite) getAgoraInfoAPIReturnCorrectData(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	resp := commonState.Response.(*cpb.GetAppInfoResponse)
	if resp == nil {
		return ctx, fmt.Errorf("response GetAppInfoResponse is nil")
	}

	if resp.CurrentUserToken == "" {
		return ctx, fmt.Errorf("expected agora user token")
	}

	if resp.AppKey == "" {
		return ctx, fmt.Errorf("expected agora app key")
	}

	if resp.TokenExpiredAt == 0 {
		return ctx, fmt.Errorf("expected token expire time > 0")
	}
	return ctx, nil
}
