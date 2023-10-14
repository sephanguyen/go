package services

import (
	"context"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

// nolint:unused,structcheck
type TestCase struct {
	name               string
	ctx                context.Context
	req                interface{}
	expectedResp       interface{}
	expectedErr        error
	expectedErrorLines []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError
	setup              func(ctx context.Context)
	csvLine            [][]string
}
