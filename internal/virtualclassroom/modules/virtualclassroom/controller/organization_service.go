package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
)

type OrganizationService struct {
	Query queries.OrganizationQuery
}

func (g *OrganizationService) GetOrgs(ctx context.Context) (map[string]string, error) {
	return g.Query.GetOrganizationMap(ctx)
}
