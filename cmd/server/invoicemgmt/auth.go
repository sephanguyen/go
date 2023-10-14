package invoicemgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/invoicemgmt.v1.InternalService/InvoiceScheduleChecker",
	"/invoicemgmt.v1.InternalService/RetrieveStudentPaymentMethod",
	"/grpc.health.v1.Health/Check",
}

var fakeJwtCtxEndpoint = []string{
	"/invoicemgmt.v1.InternalService/InvoiceScheduleChecker",
	"/invoicemgmt.v1.InternalService/RetrieveStudentPaymentMethod",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch":                                        nil,
	"/invoicemgmt.v1.InvoiceService/IssueInvoice":                         {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/GenerateInvoices":                     {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/RetrieveInvoiceRecords":               {constant.RoleParent},
	"/invoicemgmt.v1.InvoiceService/RetrieveInvoiceInfo":                  {constant.RoleParent},
	"/invoicemgmt.v1.InvoiceService/VoidInvoice":                          {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/ApproveInvoicePayment":                {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/CancelInvoicePayment":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.ImportMasterDataService/ImportInvoiceSchedule":       {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.ImportMasterDataService/ImportPartnerBank":           {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/BulkIssueInvoice":                     {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/CreatePaymentRequest":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/DownloadPaymentFile":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/DownloadBulkPaymentValidationsDetail": {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/CreateBulkPaymentValidation":          {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.EditPaymentDetailService/UpsertStudentPaymentInfo":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/invoicemgmt.v1.EditPaymentDetailService/UpdateStudentPaymentMethod": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/invoicemgmt.v1.ExportMasterDataService/ExportInvoiceSchedule":       {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.ExportMasterDataService/ExportBank":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.ExportMasterDataService/ExportBankBranch":            {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.ExportMasterDataService/ExportBankMapping":           {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/CreateInvoiceFromOrder":               {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/UpsertInvoiceAdjustments":             {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	// v2 endpoints
	"/invoicemgmt.v1.InvoiceService/IssueInvoiceV2":     {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/VoidInvoiceV2":      {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/BulkIssueInvoiceV2": {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/RefundInvoice":      {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	// v2 endpoints on Payment Service
	"/invoicemgmt.v1.PaymentService/CancelInvoicePaymentV2":  {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.PaymentService/AddInvoicePayment":       {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.PaymentService/ApproveInvoicePaymentV2": {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	// order squad related endpoints on Payment Service
	"/invoicemgmt.v1.PaymentService/RetrieveStudentPaymentMethod":     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/invoicemgmt.v1.PaymentService/RetrieveBulkStudentPaymentMethod": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},

	//  migration data endpoint
	"/invoicemgmt.v1.DataMigrationService/ImportDataMigration": {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	// bulk add payment on Payment Service
	"/invoicemgmt.v1.PaymentService/BulkAddPayment": {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	// bulk cancel payment on Payment Service
	"/invoicemgmt.v1.PaymentService/BulkCancelPayment": {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	// Retrieve Invoice Data
	"/invoicemgmt.v1.InvoiceService/RetrieveInvoiceData":        {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/invoicemgmt.v1.InvoiceService/RetrieveInvoiceStatusCount": {constant.RoleSchoolAdmin, constant.RoleHQStaff},
}

func fakeSchoolAdminJwtInterceptor() *gl_interceptors.FakeJwtContext {
	endpoints := map[string]struct{}{}
	for _, endpoint := range fakeJwtCtxEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	return gl_interceptors.NewFakeJwtContext(endpoints, constant.UserGroupSchoolAdmin)
}

func authInterceptor(c *configurations.Config, l *zap.Logger, db database.QueryExecer) *interceptors.Auth {
	groupDecider := &interceptors.GroupDecider{
		GroupFetcher: func(ctx context.Context, userID string) ([]string, error) {
			userRepo := &repository.UserRepo{}
			return interceptors.RetrieveUserRoles(ctx, userRepo, db)
		},
		AllowedGroups: rbacDecider,
	}

	auth, err := interceptors.NewAuth(
		ignoreAuthEndpoint,
		groupDecider,
		c.Issuers,
	)
	if err != nil {
		l.Panic("err init authInterceptor", zap.Error(err))
	}

	return auth
}
