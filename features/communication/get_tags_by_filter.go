package communication

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type GetTagsByFilterSuite struct {
	*common.NotificationSuite
	createdFilter  repositories.FindTagFilter
	createdKeyword string
}

func (c *SuiteConstructor) InitGetTagsByFilter(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &GetTagsByFilterSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" tag with "([^"]*)" name$`:                                                                  s.schoolAdminCreatesTagWithName,
		`^response have correct "([^"]*)" and "([^"]*)"$`:                                                                           s.responseHaveCorrectAnd,
		`^school admin delete "([^"]*)" of those tag$`:                                                                              s.schoolAdminDeleteOfThoseTag,
		`^school admin search with filter of "([^"]*)" result at position "([^"]*)"$`:                                               s.schoolAdminSearchWithFilterOfResultAtPosition,
		`^school admin see "([^"]*)" in total of "([^"]*)"$`:                                                                        s.schoolAdminSeeInTotalOf,
		`^school admin archived those tag$`:                                                                                         s.schoolAdminArchivedThoseTag,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *GetTagsByFilterSuite) schoolAdminCreatesTagWithName(ctx context.Context, num, keyword string) (context.Context, error) {
	if keyword != "random" {
		s.createdKeyword = keyword
	}
	return s.CreatesNumberOfTags(ctx, num, keyword)
}

func (s *GetTagsByFilterSuite) schoolAdminSearchWithFilterOfResultAtPosition(ctx context.Context, limit, offset int) (context.Context, error) {
	filter := repositories.NewFindTagFilter()
	filter.Limit = database.Int8(int64(limit))
	filter.Offset = database.Int8(int64(offset))

	if s.createdKeyword != "" {
		_ = filter.Keyword.Set(s.createdKeyword)
	}

	s.createdFilter = filter
	return ctx, nil
}

func (s *GetTagsByFilterSuite) responseHaveCorrectAnd(ctx context.Context, prevOffset, nextOffset int) (context.Context, error) {
	res, err := npb.NewTagMgmtReaderServiceClient(s.NotificationMgmtGRPCConn).GetTagsByFilter(
		ctx,
		&npb.GetTagsByFilterRequest{
			Keyword: s.createdFilter.Keyword.String,
			Paging: &cpb.Paging{
				Limit:  uint32(s.createdFilter.Limit.Int),
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: s.createdFilter.Offset.Int},
			},
		},
	)
	if err != nil {
		return ctx, fmt.Errorf("GetTagsByFilter: %v", err)
	}

	if res.NextPage.GetOffsetInteger() != int64(nextOffset) {
		return ctx, fmt.Errorf("incorrect next offset, expected %d, got %d", nextOffset, res.NextPage.GetOffsetInteger())
	}

	if res.PreviousPage.GetOffsetInteger() != int64(prevOffset) {
		return ctx, fmt.Errorf("incorrect prev offset, expected %d, got %d", prevOffset, res.PreviousPage.GetOffsetInteger())
	}
	return ctx, nil
}

func (s *GetTagsByFilterSuite) schoolAdminDeleteOfThoseTag(ctx context.Context, numDelete int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	tagIDsWithKeyword := []string{}
	for _, tag := range commonState.Tags {
		if strings.Contains(tag.Name, s.createdKeyword) {
			tagIDsWithKeyword = append(tagIDsWithKeyword, tag.ID)
		}
	}
	i := 0
	for i < numDelete && i < len(tagIDsWithKeyword) {
		_, err := npb.NewTagMgmtModifierServiceClient(s.NotificationMgmtGRPCConn).DeleteTag(
			ctx,
			&npb.DeleteTagRequest{
				TagId: tagIDsWithKeyword[i],
			},
		)
		if err != nil {
			return nil, fmt.Errorf("DeleteTag: %v", err)
		}
		i++
	}
	return ctx, nil
}

func (s *GetTagsByFilterSuite) schoolAdminSeeInTotalOf(ctx context.Context, numResult, numTotal int) (context.Context, error) {
	res, err := npb.NewTagMgmtReaderServiceClient(s.NotificationMgmtGRPCConn).GetTagsByFilter(
		ctx,
		&npb.GetTagsByFilterRequest{
			Keyword: s.createdFilter.Keyword.String,
			Paging: &cpb.Paging{
				Limit:  uint32(s.createdFilter.Limit.Int),
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: s.createdFilter.Offset.Int},
			},
		},
	)
	if err != nil {
		return ctx, fmt.Errorf("GetTagsByFilter: %v", err)
	}

	if res.TotalItems != uint32(numTotal) {
		return ctx, fmt.Errorf("expected total item is %d, got %d", numTotal, res.TotalItems)
	}

	if len(res.Tags) != numResult {
		return ctx, fmt.Errorf("expected number result is %d, got %d", numResult, len(res.Tags))
	}
	for _, tag := range res.Tags {
		if !strings.Contains(tag.GetName(), s.createdKeyword) {
			return ctx, fmt.Errorf("result not contain expected keyword %s, got %s", s.createdKeyword, tag.GetName())
		}
	}
	return ctx, nil
}

func (s *GetTagsByFilterSuite) schoolAdminArchivedThoseTag(ctx context.Context) (context.Context, error) {
	query := `
		UPDATE tags SET is_archived = TRUE
		WHERE tag_name LIKE '%archived%';
	`
	_, err := s.BobDBConn.Exec(ctx, query)
	if err != nil {
		return ctx, fmt.Errorf("failed set archive: %v", err)
	}
	return ctx, nil
}
