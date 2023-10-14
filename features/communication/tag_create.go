package communication

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TagCreateSuite struct {
	*common.NotificationSuite
	validTag   *entities.Tag
	existError error
}

func (c *SuiteConstructor) InitTagCreate(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &TagCreateSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^admin upsert tag$`:                             s.adminUpsertTag,
		`^admin upsert tag with updated data$`:           s.adminUpsertTagWithUpdatedData,
		`^return error tag name existed$`:                s.returnErrorTagNameExisted,
		`^tag data is stored in the database correctly$`: s.tagDataIsStoredInTheDatabaseCorrectly,
		`^tag ID is not existed in database$`:            s.tagIDIsNotExistedInDatabase,
		`^tag Name is existed in database$`:              s.tagNameIsExistedInDatabase,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *TagCreateSuite) tagIDIsNotExistedInDatabase(ctx context.Context) (context.Context, error) {
	s.validTag = &entities.Tag{
		TagID:   database.Text(idutil.ULIDNow()),
		TagName: database.Text(idutil.ULIDNow()),
	}

	return ctx, nil
}

func (s *TagCreateSuite) adminUpsertTag(ctx context.Context) (context.Context, error) {
	upsertRequest := &npb.UpsertTagRequest{
		TagId: s.validTag.TagID.String,
		Name:  s.validTag.TagName.String,
	}
	_, err := npb.NewTagMgmtModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertTag(
		ctx,
		upsertRequest,
	)
	if err != nil {
		if errors.Is(err, status.Error(codes.InvalidArgument, "TagName is exist")) {
			s.existError = err
			return ctx, nil
		}
		return ctx, fmt.Errorf("failed upsert tag: %v", err)
	}
	return ctx, nil
}

func (s *TagCreateSuite) tagDataIsStoredInTheDatabaseCorrectly(ctx context.Context) (context.Context, error) {
	query := `SELECT COUNT(*) FROM tags WHERE tag_id=$1 AND tag_name=$2 AND deleted_at IS NULL`
	row := s.BobDBConn.QueryRow(ctx, query, s.validTag.TagID, s.validTag.TagName)
	var count int
	if err := row.Scan(&count); err != nil {
		return ctx, err
	}
	expectedCount := 1
	if count > 1 || count == 0 {
		return ctx, fmt.Errorf("expected tag count: %v, got: %v", expectedCount, count)
	}
	return ctx, nil
}

func (s *TagCreateSuite) adminUpsertTagWithUpdatedData(ctx context.Context) (context.Context, error) {
	s.validTag.TagName = database.Text(idutil.ULIDNow())
	return s.adminUpsertTag(ctx)
}

func (s *TagCreateSuite) tagNameIsExistedInDatabase(ctx context.Context) (context.Context, error) {
	s.validTag = &entities.Tag{
		TagName: database.Text(idutil.ULIDNow()),
	}

	return s.adminUpsertTag(ctx)
}

func (s *TagCreateSuite) returnErrorTagNameExisted(ctx context.Context) (context.Context, error) {
	if s.existError == nil {
		return ctx, fmt.Errorf("returnErrorTagNameExisted: expected %v, got %v", status.Error(codes.InvalidArgument, "TagName is exist"), nil)
	}
	return ctx, nil
}
