package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type TagCheckTagNameExistSuite struct {
	*common.NotificationSuite
	response *npb.CheckExistTagNameResponse
	tagName  string
}

func (c *SuiteConstructor) InitTagCheckTagNameExist(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &TagCheckTagNameExistSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^a tag "([^"]*)" in database$`: s.aTagInDatabase,
		`^admin send check request$`:    s.adminSendCheckRequest,
		`^check response is "([^"]*)"$`: s.checkResponseIs,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *TagCheckTagNameExistSuite) aTagInDatabase(ctx context.Context, isExist string) (context.Context, error) {
	s.tagName = idutil.ULIDNow()
	if isExist == "false" {
		return ctx, nil // do nothing
	}
	return s.CreatesTagsWithNames(ctx, s.tagName)
}

func (s *TagCheckTagNameExistSuite) adminSendCheckRequest(ctx context.Context) (context.Context, error) {
	checkExistRequest := &npb.CheckExistTagNameRequest{
		TagName: s.tagName,
	}
	var err error
	s.response, err = npb.NewTagMgmtReaderServiceClient(s.NotificationMgmtGRPCConn).CheckExistTagName(
		ctx,
		checkExistRequest)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *TagCheckTagNameExistSuite) checkResponseIs(ctx context.Context, response string) (context.Context, error) {
	checkResponse := false
	if response == "true" {
		checkResponse = true
	}
	if s.response == nil {
		return ctx, fmt.Errorf("checkResponseIs: did not received response")
	}
	if s.response.IsExist != checkResponse {
		return ctx, fmt.Errorf("expected response to be: %v, got %v", checkResponse, s.response.IsExist)
	}
	return ctx, nil
}
