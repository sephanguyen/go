package communication

import (
	"bytes"
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/services/mappers"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type TagExportCsvSuite struct {
	*common.NotificationSuite
	csvFile []byte
}

func (c *SuiteConstructor) InitTagExportCsv(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &TagExportCsvSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin create some tags named "([^"]*)"$`: s.CreatesTagsWithNames,
		`^school admin export tags$`:                      s.schoolAdminExportTags,
		`^csv data is correctly exported$`:                s.csvDataIsCorrectlyExported,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *TagExportCsvSuite) schoolAdminExportTags(ctx context.Context) (context.Context, error) {
	res, err := npb.NewTagMgmtReaderServiceClient(s.NotificationMgmtGRPCConn).ExportTags(
		ctx,
		&npb.ExportTagsRequest{},
	)
	if err != nil {
		return ctx, fmt.Errorf("failed ExportTags: %v", err)
	}
	if len(res.GetData()) == 0 {
		return ctx, fmt.Errorf("expected CSV file not empty")
	}
	s.csvFile = res.GetData()

	// just to check if export structure is valid
	_, err = npb.NewTagMgmtModifierServiceClient(s.NotificationMgmtGRPCConn).ImportTags(
		ctx,
		&npb.ImportTagsRequest{
			Payload: s.csvFile,
		},
	)
	if err != nil {
		return ctx, fmt.Errorf("failed import data from exported csv: %v", err)
	}
	return ctx, nil
}

func (s *TagExportCsvSuite) csvDataIsCorrectlyExported(ctx context.Context) (context.Context, error) {
	sc := scanner.NewCSVScanner(bytes.NewReader(s.csvFile))
	tagIDs := []string{}
	tagNames := []string{}
	isArchives := []bool{}
	for sc.Scan() {
		tag, err := mappers.CSVRowToTag(sc)
		if err != nil {
			return ctx, fmt.Errorf("failed CSVRowToTag: %v", err)
		}
		tagIDs = append(tagIDs, tag.TagID.String)
		tagNames = append(tagNames, tag.TagName.String)
		isArchives = append(isArchives, tag.IsArchived.Bool)
	}

	query := `
		WITH tmp AS (
			SELECT unnest($1::TEXT[]) AS tag_id
		)
		SELECT *
		FROM tmp
		WHERE tmp.tag_id NOT IN (
			SELECT t.tag_id
			FROM tags t
			WHERE t.deleted_at IS NULL
			AND t.tag_id = any($1::TEXT[])
			AND t.tag_name = any($2::TEXT[])
			AND t.is_archived = any($3::BOOL[])
		);
	`
	rows, err := s.BobDBConn.Query(ctx, query,
		database.TextArray(tagIDs),
		database.TextArray(tagNames),
		database.BoolArray(isArchives))
	if err != nil {
		return ctx, fmt.Errorf("failed query: %v", err)
	}
	defer rows.Close()

	errTagIDs := []string{}
	for rows.Next() {
		var tagID string
		err = rows.Scan(&tagID)
		if err != nil {
			return ctx, fmt.Errorf("error scan: %v", err)
		}
		errTagIDs = append(errTagIDs, tagID)
	}

	if len(errTagIDs) > 0 {
		return ctx, fmt.Errorf("tags %v are not correct or not exist", errTagIDs)
	}

	return ctx, nil
}
