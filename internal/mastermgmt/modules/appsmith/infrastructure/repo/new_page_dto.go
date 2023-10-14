package repo

import (
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (n *NewPage) ColletionName() string {
	return "newPage"
}

type defaultResources struct {
	ApplicationID string `bson:"applicationId"`
	PageID        string `bson:"pageId"`
	BranchName    string `bson:"branchName,omitempty"`
}
type NewPage struct {
	ID               primitive.ObjectID `bson:"_id"`
	DefaultResources defaultResources   `bson:"defaultResources"`
	Deleted          bool               `bson:"deleted"`
}

func (n *NewPage) ToNewPageEntity() *domain.NewPage {
	newPage := &domain.NewPage{
		ID:            n.DefaultResources.PageID,
		ApplicationID: n.DefaultResources.ApplicationID,
		Deleted:       n.Deleted,
	}
	return newPage
}
