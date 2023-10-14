package repo

import (
	"testing"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/stretchr/testify/assert"
)

func TestToNewPageEntity(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name          string
		newPageDTO    *NewPage
		newPageDomain *domain.NewPage
	}{
		{
			name: "full fields",
			newPageDTO: &NewPage{
				ID: primitive.NewObjectID(),
				DefaultResources: defaultResources{
					ApplicationID: "application-id-1",
					PageID:        "page-1",
					BranchName:    "branchname-1",
				},
				Deleted: false,
			},
			newPageDomain: &domain.NewPage{
				ApplicationID: "application-id-1",
				ID:            "page-1",
				Deleted:       false,
			},
		},
		{
			name: "full fields with deleted",
			newPageDTO: &NewPage{
				ID: primitive.NewObjectID(),
				DefaultResources: defaultResources{
					ApplicationID: "application-id-1",
					PageID:        "page-1",
					BranchName:    "branchname-1",
				},
				Deleted: true,
			},
			newPageDomain: &domain.NewPage{
				ApplicationID: "application-id-1",
				ID:            "page-1",
				Deleted:       true,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			newPage := tc.newPageDTO
			actual := newPage.ToNewPageEntity()
			assert.Equal(t, tc.newPageDomain, actual)
		})
	}
}
