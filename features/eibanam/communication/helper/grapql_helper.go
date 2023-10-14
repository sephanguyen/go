package helper

import (
	"context"

	"github.com/manabie-com/backend/features/eibanam"
	"github.com/manabie-com/backend/features/eibanam/communication/entity"
)

func (h CommunicationHelper) GrantPermissionToQueryGraphql(admin *entity.Admin, tableNames ...string) error {
	if err := eibanam.CreateSelectPermissionForHasuraQuery(
		h.hasuraAdminUrl,
		admin.UserGroup,
		tableNames...,
	); err != nil {
		return err
	}
	return nil
}
func (h CommunicationHelper) QueryHasura(schoolAdmin *entity.Admin, query interface{}, variables map[string]interface{}) ([]byte, error) {
	ctx := eibanam.ContextWithToken(context.Background(), schoolAdmin.Token)
	res, err := eibanam.QueryRawHasura(ctx, h.hasuraAdminUrl, query, variables)
	if err != nil {
		return nil, err
	}
	return res.MarshalJSON()
}
