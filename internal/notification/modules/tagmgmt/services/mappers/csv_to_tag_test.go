package mappers

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	"github.com/stretchr/testify/assert"
)

func Test_CSVRowToTag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name    string
		Payload []byte
		Tag     *entities.Tag
		Err     error
	}{
		{
			Name: "case missing tag name",
			Payload: []byte(`tag_id,tag_name,is_archived
			tag-id-1,,0`),
			Tag: nil,
			Err: fmt.Errorf("value of column tag_name is empty"),
		},
		{
			Name: "case missing tag id",
			Payload: []byte(`tag_id,tag_name,is_archived
			,tag name,0`),
			Tag: &entities.Tag{
				TagID:      database.Text(""),
				TagName:    database.Text("tag name"),
				IsArchived: database.Bool(false),
			},
			Err: nil,
		},
		{
			Name: "case missing tag id and is archive true",
			Payload: []byte(`tag_id,tag_name,is_archived
			,tag name,1`),
			Tag: &entities.Tag{
				TagID:      database.Text(""),
				TagName:    database.Text("tag name"),
				IsArchived: database.Bool(true),
			},
			Err: nil,
		},
		{
			Name: "case missing is archived",
			Payload: []byte(`tag_id,tag_name,is_archived
			tag-id-1,tag name`),
			Tag: nil,
			Err: fmt.Errorf("value of column is_archived is invalid"),
		},
		{
			Name: "case is archived invalid",
			Payload: []byte(`tag_id,tag_name,is_archived
			tag-id-1,tag name,abc`),
			Tag: nil,
			Err: fmt.Errorf("value of column is_archived is invalid"),
		},
		{
			Name:    "case all field empty",
			Payload: []byte(`tag_id,tag_name,is_archived`),
			Tag:     nil,
			Err:     fmt.Errorf("value of column is_archived is invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			sc := scanner.NewCSVScanner(bytes.NewReader(tc.Payload))
			for sc.Scan() {
				tag, err := CSVRowToTag(sc)
				if tc.Err != nil {
					assert.Equal(t, tc.Err.Error(), err.Error())
				} else {
					assert.NotEmpty(t, tc.Tag.TagID)
					assert.Equal(t, tc.Tag.TagName, tag.TagName)
					assert.Equal(t, tc.Tag.IsArchived, tag.IsArchived)
				}
			}
		})
	}
}
