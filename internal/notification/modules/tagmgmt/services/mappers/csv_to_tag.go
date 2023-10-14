package mappers

import (
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"

	"go.uber.org/multierr"
	"k8s.io/utils/strings/slices"
)

func CSVRowToTag(sc scanner.CSVScanner) (*entities.Tag, error) {
	tag := &entities.Tag{}
	database.AllNullEntity(tag)
	allowedHeaders := strings.Split(consts.AllowTagCSVHeaders, "|")

	var (
		tagID         string
		tagName       string
		strIsArchived string
	)

	// allow to input empty tag id in csv
	tagID = sc.Text(allowedHeaders[0])

	if tagName = sc.Text(allowedHeaders[1]); tagName == "" {
		return nil, fmt.Errorf("value of column %s is empty", allowedHeaders[1])
	}

	strIsArchived = sc.Text(allowedHeaders[2])
	if !slices.Contains([]string{"0", "1"}, strIsArchived) {
		return nil, fmt.Errorf("value of column %s is invalid", allowedHeaders[2])
	}

	isArchived := sc.Text(allowedHeaders[2]) == "1"
	now := time.Now()
	err := multierr.Combine(
		tag.TagID.Set(tagID),
		tag.TagName.Set(tagName),
		tag.CreatedAt.Set(now),
		tag.UpdatedAt.Set(now),
		tag.IsArchived.Set(isArchived),
	)

	if err != nil {
		return nil, fmt.Errorf("failed combine: %v", err)
	}

	return tag, nil
}
