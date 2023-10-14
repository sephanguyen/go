package communication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TagDeleteSuite struct {
	*common.NotificationSuite
	validTag          *entities.Tag
	validTagDeleteReq *npb.DeleteTagRequest
	existError        error
}

func (c *SuiteConstructor) InitTagDelete(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &TagDeleteSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^a valid delete tag request$`:     s.aValidDeleteTagRequest,
		`^tag is exist in database$`:       s.tagIsExistInDatabase,
		`^tag is soft delete in database$`: s.tagIsSoftDeleteInDatabase,
		`^admin delete tag$`:               s.adminDeleteTag,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *TagDeleteSuite) aValidDeleteTagRequest(ctx context.Context) (context.Context, error) {
	s.validTagDeleteReq = &npb.DeleteTagRequest{
		TagId: idutil.ULIDNow(),
	}

	now := time.Now()
	e := &entities.Tag{}
	err := multierr.Combine(
		e.TagID.Set(s.validTagDeleteReq.TagId),
		e.TagName.Set(idutil.ULIDNow()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.DeletedAt.Set(nil),
	)
	if err != nil {
		return ctx, fmt.Errorf("aValidDeleteTagRequest: %w", err)
	}
	s.validTag = e
	return ctx, nil
}

func (s *TagDeleteSuite) adminDeleteTag(ctx context.Context) (context.Context, error) {
	_, err := npb.NewTagMgmtModifierServiceClient(s.NotificationMgmtGRPCConn).DeleteTag(
		ctx,
		s.validTagDeleteReq)
	if err != nil {
		if errors.Is(err, status.Error(codes.InvalidArgument, "TagName is exist")) {
			s.existError = err
			return ctx, nil
		}
		return ctx, err
	}
	return ctx, nil
}

func (s *TagDeleteSuite) tagIsExistInDatabase(ctx context.Context) (context.Context, error) {
	if s.validTag == nil {
		return ctx, fmt.Errorf("tag entity is not initialized")
	}

	// insert a sample Tag to DB
	res, err := npb.NewTagMgmtModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertTag(
		ctx,
		&npb.UpsertTagRequest{
			TagId: s.validTag.TagID.String,
			Name:  s.validTag.TagName.String,
		},
	)

	if err != nil {
		return ctx, fmt.Errorf("tagIsExistInDatabase: %w", err)
	}

	expcUpsertResponse := &npb.UpsertTagResponse{
		TagId: s.validTag.TagID.String,
	}

	if !protoEqual(res, expcUpsertResponse) {
		return ctx, fmt.Errorf("tagIsExistInDatabase. UpsertTagReponse not match. %s", protoDiff(res, expcUpsertResponse))
	}
	return ctx, nil
}

func (s *TagDeleteSuite) tagIsSoftDeleteInDatabase(ctx context.Context) (context.Context, error) {
	if s.validTag == nil {
		return ctx, fmt.Errorf("tag entity is not initialized")
	}

	query := `
		SELECT COUNT(*)
		FROM tags
		WHERE tag_id=$1 AND tag_name=$2 AND deleted_at IS NOT NULL
	`
	row := s.BobDBConn.QueryRow(ctx, query, s.validTag.TagID, s.validTag.TagName)
	var count int
	if err := row.Scan(&count); err != nil {
		return ctx, err
	}
	if count != 1 {
		return ctx, fmt.Errorf("expected tag soft deleted: %v, got %v", true, count == 1)
	}
	return ctx, nil
}
