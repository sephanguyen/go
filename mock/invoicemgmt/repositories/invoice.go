// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

type MockInvoiceRepo struct {
	mock.Mock
}

func (r *MockInvoiceRepo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.Invoice) (pgtype.Text, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(pgtype.Text), args.Error(1)
}

func (r *MockInvoiceRepo) FindInvoicesFromInvoiceIDTempTable(arg1 context.Context, arg2 database.QueryExecer) ([]*entities.Invoice, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Invoice), args.Error(1)
}

func (r *MockInvoiceRepo) InsertInvoiceIDsTempTable(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockInvoiceRepo) RetrieveInvoiceByInvoiceID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*entities.Invoice, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Invoice), args.Error(1)
}

func (r *MockInvoiceRepo) RetrieveInvoiceByInvoiceReferenceID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*entities.Invoice, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Invoice), args.Error(1)
}

func (r *MockInvoiceRepo) RetrieveInvoiceData(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Int8, arg4 pgtype.Int8, arg5 string) ([]*entities.InvoicePaymentMap, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.InvoicePaymentMap), args.Error(1)
}

func (r *MockInvoiceRepo) RetrieveInvoiceStatusCount(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (map[string]int32, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(map[string]int32), args.Error(1)
}

func (r *MockInvoiceRepo) RetrieveRecordsByStudentID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 pgtype.Int8, arg5 pgtype.Int8) ([]*entities.Invoice, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Invoice), args.Error(1)
}

func (r *MockInvoiceRepo) RetrievedMigratedInvoices(arg1 context.Context, arg2 database.QueryExecer) ([]*entities.Invoice, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Invoice), args.Error(1)
}

func (r *MockInvoiceRepo) Update(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.Invoice) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockInvoiceRepo) UpdateIsExportedByInvoiceIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string, arg4 bool) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockInvoiceRepo) UpdateIsExportedByPaymentRequestFileID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 bool) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockInvoiceRepo) UpdateMultipleWithFields(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entities.Invoice, arg4 []string) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockInvoiceRepo) UpdateStatusFromInvoiceIDTempTable(arg1 context.Context, arg2 database.QueryExecer, arg3 string) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockInvoiceRepo) UpdateWithFields(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.Invoice, arg4 []string) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}
