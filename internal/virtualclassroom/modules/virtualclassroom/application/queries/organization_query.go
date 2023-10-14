package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type OrganizationQuery struct {
	WrapperDBConnection *support.WrapperDBConnection
	OrganizationRepo    infrastructure.OrganizationRepo
}

const OrderNumber = "%03d" // 001

func (o *OrganizationQuery) GetOrganizationMap(ctx context.Context) (map[string]string, error) {
	conn, err := o.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	mOrg := make(map[string]string, 0)
	orgIds, err := o.OrganizationRepo.GetIDs(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("error in OrganizationRepo.GetIDs: %w", err)
	}
	for i, v := range orgIds {
		mOrg[fmt.Sprintf(OrderNumber, i)] = v
	}
	return mOrg, nil
}
