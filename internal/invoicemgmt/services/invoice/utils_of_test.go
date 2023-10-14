package invoicesvc

import (
	"context"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

// nolint:unused,structcheck
type TestCase struct {
	name                string
	ctx                 context.Context
	req                 interface{}
	expectedResp        interface{}
	expectedErr         error
	setup               func(ctx context.Context)
	mockInvoiceEntities []*entities.Invoice
}
