package entities

import (
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntity(t *testing.T) {
	t.Parallel()
	sv, err := database.NewSchemaVerifier("invoicemgmt")
	require.NoError(t, err)

	ents := []database.Entity{
		&InvoiceBillItem{},
		&Invoice{},
		&BillItem{},
		&User{},
		&UserAccessPaths{},
		&Location{},
		&Student{},
		&Payment{},
		&InvoiceActionLog{},
		&InvoiceCreditNote{},
		&InvoiceSchedule{},
		&InvoiceScheduleHistory{},
		&InvoiceScheduleStudent{},
		&Organization{},
		&Bank{},
		&BulkPaymentRequest{},
		&BulkPaymentRequestFile{},
		&Discount{},
		&NewCustomerCodeHistory{},
		&BulkPaymentRequestFilePayment{},
		&PartnerConvenienceStore{},
		&PartnerBank{},
		&BankMapping{},
		&BulkPaymentValidations{},
		&BulkPaymentValidationsDetail{},
		&StudentPaymentDetail{},
		&Prefecture{},
		&BillingAddress{},
		&BankAccount{},
		&BankBranch{},
		&Order{},
		&UserBasicInfo{},
		&InvoiceAdjustment{},
		&BulkPayment{},
		&CompanyDetail{},
		&StudentPaymentDetailActionLog{},
	}

	assert := assert.New(t)
	dir, err := os.Getwd()
	assert.NoError(err)

	count, err := database.CheckEntity(dir)
	assert.NoError(err)
	assert.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assert.NoError(database.CheckEntityDefinition(e))
		assert.NoError(sv.Verify(e))
	}
}

func TestEntities(t *testing.T) {
	t.Parallel()
	ents := []database.Entities{&InvoiceBillItems{}}

	assert := assert.New(t)
	dir, err := os.Getwd()
	assert.NoError(err)

	count, err := database.CheckEntities(dir)
	assert.NoError(err)
	assert.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assert.NoError(database.CheckEntitiesDefinition(e))
	}
}
