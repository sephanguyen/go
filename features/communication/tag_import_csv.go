package communication

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	bddEntities "github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/consts"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"k8s.io/utils/strings/slices"
)

type TagImportCsvSuite struct {
	*common.NotificationSuite
	csvFile    []byte
	mapTagName map[string]*bddEntities.Tag
	mapTagID   map[string]*bddEntities.Tag
}

func (c *SuiteConstructor) InitTagImportCsv(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &TagImportCsvSuite{
		NotificationSuite: dep.notiCommonSuite,
		mapTagName:        make(map[string]*bddEntities.Tag),
		mapTagID:          make(map[string]*bddEntities.Tag),
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^admin create "([^"]*)" tag with "([^"]*)" keywords$`:                                                                      s.CreatesNumberOfTags,
		`^a valid csv file$`:                                            s.aValidCsvFile,
		`^admin import csv tag file$`:                                   s.adminImportCsvTagFile,
		`^csv data is correctly stored in database$`:                    s.csvDataIsCorrectlyStoredInDatabase,
		`^returns "([^"]*)" status code and error message have "(.*)"$`: s.CheckReturnStatusCodeAndContainMsg,
		`^a invalid csv file with "([^"]*)"$`:                           s.aInvalidCsvFileWith,
		`^admin update csv file$`:                                       s.adminUpdateCsvFile,
		`^returns "([^"]*)" status code$`:                               s.CheckReturnStatusCode,
		`^admin update csv to "([^"]*)" tag "([^"]*)"$`:                 s.adminUpdateCsvToTag,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *TagImportCsvSuite) aInvalidCsvFileWith(ctx context.Context, invalidType string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	var strPayload string
	switch invalidType {
	case "wrong header":
		strPayload = "tag_order"
	case "missing header":
		headers := strings.Split(consts.AllowTagCSVHeaders, "|")
		strPayload = strings.Join(headers[:len(headers)-1], ",")
	}
	s.csvFile = []byte(strPayload)

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *TagImportCsvSuite) aValidCsvFile(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	strPayload := strings.ReplaceAll(consts.AllowTagCSVHeaders, "|", ",") + "\n"
	// case DB is empty
	if len(commonState.Tags) == 0 {
		for i := 0; i < 5; i++ {
			tag := &bddEntities.Tag{
				ID:         "",
				Name:       idutil.ULIDNow() + " case empty db",
				IsArchived: false,
			}
			if i%2 == 0 {
				tag.IsArchived = true
			}
			commonState.Tags = append(commonState.Tags, tag)
			s.mapTagName[tag.Name] = tag
		}
	}

	for _, tag := range commonState.Tags {
		strIsArchived := "0"
		if tag.IsArchived {
			strIsArchived = "1"
		}
		strPayload += tag.ID + "," + tag.Name + "," + strIsArchived + "\n"
		s.mapTagID[tag.ID] = tag
	}
	s.csvFile = []byte(strPayload)
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *TagImportCsvSuite) adminUpdateCsvFile(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	strPayload := strings.ReplaceAll(consts.AllowTagCSVHeaders, "|", ",") + "\n"
	for _, tag := range commonState.Tags {
		tag.Name = idutil.ULIDNow()
		tag.IsArchived = !tag.IsArchived
		strIsArchived := "0"
		if tag.IsArchived {
			strIsArchived = "1"
		}
		strPayload += tag.ID + "," + tag.Name + "," + strIsArchived + "\n"
		s.mapTagID[tag.ID] = tag
	}

	// insert new tags
	for i := 0; i < 3; i++ {
		tagName := idutil.ULIDNow() + " insert"
		tag := &bddEntities.Tag{
			ID:         "",
			Name:       tagName,
			IsArchived: false,
		}
		commonState.Tags = append(commonState.Tags, tag)
		s.mapTagName[tagName] = tag
		strPayload += "," + tagName + "," + "0" + "\n"
	}
	s.csvFile = []byte(strPayload)
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *TagImportCsvSuite) adminUpdateCsvToTag(ctx context.Context, errType, tagField string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	toBeDuplicatedIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(commonState.Tags))))
	if err != nil {
		return ctx, fmt.Errorf("failed to random: %v", err)
	}
	switch errType {
	case "duplicate":
		for idx, tag := range commonState.Tags {
			if idx != int(toBeDuplicatedIdx.Int64()) {
				switch tagField {
				case "name":
					tag.Name = commonState.Tags[toBeDuplicatedIdx.Int64()].Name
				case "id":
					tag.ID = commonState.Tags[toBeDuplicatedIdx.Int64()].ID
				}
				break
			}
		}
	case "not exist":
		for idx, tag := range commonState.Tags {
			if idx != int(toBeDuplicatedIdx.Int64()) {
				tag.ID = idutil.ULIDNow()
				break
			}
		}
	}

	s.csvFile = s.toCSVPayload(commonState.Tags, false)
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *TagImportCsvSuite) adminImportCsvTagFile(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	_, commonState.ResponseErr = npb.NewTagMgmtModifierServiceClient(s.NotificationMgmtGRPCConn).ImportTags(
		ctx,
		&npb.ImportTagsRequest{
			Payload: s.csvFile,
		},
	)
	return ctx, nil
}

func (s *TagImportCsvSuite) csvDataIsCorrectlyStoredInDatabase(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	if len(s.mapTagName) > 0 {
		// these tags won't have ID
		tagNames := []string{}
		for _, tag := range commonState.Tags {
			tagNames = append(tagNames, tag.Name)
		}

		query := `
			SELECT tag_id, tag_name, is_archived
			FROM tags
			WHERE tag_name = ANY($1::TEXT[]) AND resource_path = $2 AND deleted_at IS NULL
		`
		rows, err := s.BobDBConn.Query(ctx, query, database.TextArray(tagNames), fmt.Sprint(commonState.CurrentOrganicationID))
		if err != nil {
			return ctx, fmt.Errorf("failed query: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				tagID      string
				tagName    string
				isArchived bool
			)

			err = rows.Scan(&tagID, &tagName, &isArchived)
			if err != nil {
				return ctx, fmt.Errorf("failed scan: %v", err)
			}

			if !slices.Contains(tagNames, tagName) {
				return ctx, fmt.Errorf("unexpected tag name %s", tagName)
			}

			if tag := s.mapTagName[tagName]; tag != nil {
				if tag.IsArchived != isArchived {
					return ctx, fmt.Errorf("expected isArchived to be %t, found %t", tag.IsArchived, isArchived)
				}
			}
		}
	}

	if len(s.mapTagID) > 0 {
		tagIDs := []string{}
		for _, tag := range s.mapTagID {
			tagIDs = append(tagIDs, tag.ID)
		}

		query := `
			SELECT tag_id, tag_name, is_archived
			FROM tags
			WHERE tag_id = ANY($1::TEXT[]) AND resource_path = $2 AND deleted_at IS NULL
		`
		rows, err := s.BobDBConn.Query(ctx, query, database.TextArray(tagIDs), fmt.Sprint(commonState.CurrentOrganicationID))
		if err != nil {
			return ctx, fmt.Errorf("failed query: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				tagID      string
				tagName    string
				isArchived bool
			)

			err = rows.Scan(&tagID, &tagName, &isArchived)
			if err != nil {
				return ctx, fmt.Errorf("failed scan: %v", err)
			}

			if !slices.Contains(tagIDs, tagID) {
				return ctx, fmt.Errorf("unexpected tag id %s", tagID)
			}

			if tag := s.mapTagID[tagID]; tag != nil {
				if tag.Name != tagName {
					return ctx, fmt.Errorf("expected tagName to be %s, found %s", tag.Name, tagName)
				}
				if tag.IsArchived != isArchived {
					return ctx, fmt.Errorf("expected isArchived to be %t, found %t", tag.IsArchived, isArchived)
				}
			}
		}
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *TagImportCsvSuite) toCSVPayload(tags []*bddEntities.Tag, emptyID bool) []byte {
	strPayload := strings.ReplaceAll(consts.AllowTagCSVHeaders, "|", ",") + "\n"
	switch emptyID {
	case true:
		for _, tag := range tags {
			strIsArchived := "0"
			if tag.IsArchived {
				strIsArchived = "1"
			}
			strPayload += "," + tag.Name + "," + strIsArchived + "\n"
			s.mapTagName[tag.Name] = tag
		}
	case false:
		for _, tag := range tags {
			strIsArchived := "0"
			if tag.IsArchived {
				strIsArchived = "1"
			}
			strPayload += tag.ID + "," + tag.Name + "," + strIsArchived + "\n"
			s.mapTagID[tag.ID] = tag
		}
	}

	return []byte(strPayload)
}
