package payment

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/payment.v1.InternalService/UpdateBillItemStatus",
	"/payment.v1.InternalService/UpdateStudentProductStatus",
	"/payment.v1.InternalService/GenerateBillingItems",
	"/payment.v1.InternalService/UpdateStudentCourse",
	"/payment.v1.InternalService/UpdateStudentPackage",

	"/grpc.health.v1.Health/Check",
}

var fakeJwtCtxEndpoint = []string{
	"/payment.v1.InternalService/UpdateBillItemStatus",
	"/payment.v1.InternalService/UpdateStudentProductStatus",
	"/payment.v1.InternalService/GenerateBillingItems",
	"/payment.v1.InternalService/UpdateStudentCourse",
	"/payment.v1.InternalService/UpdateStudentPackage",
}

var rbacDecider = map[string][]string{
	"/payment.v1.EchoService/Echo":                                            {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataForTestService/ImportAllForTest":             {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportAccountingCategory":            {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportBillingSchedule":               {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportBillingSchedulePeriod":         {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportDiscount":                      {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportProduct":                       {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportTax":                           {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportProductPrice":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportProductAssociatedData":         {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportBillingRatioType":              {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportLeavingReason":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportBillingRatio":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportPackageQuantityTypeMapping":    {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportProductSetting":                {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportAssociatedProducts":            {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportProductGroup":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportProductGroupMapping":           {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.ImportMasterDataService/ImportNotificationDate":              {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.OrderService/CreateOrder":                                    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/CreateBulkOrder":                                {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/CreateCustomBilling":                            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/RetrieveListOfOrders":                           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.OrderService/RetrieveListOfBillItems":                        {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.OrderService/RetrieveListOfOrderItems":                       {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.OrderService/RetrieveListOfOrderProducts":                    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.OrderService/RetrieveBillingOfOrderDetails":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.OrderService/RetrieveListOfOrderDetailProducts":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.OrderService/UpdateBillItemStatus":                           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/UpdateOrderStatus":                              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/VoidOrder":                                      {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.OrderService/RetrieveRecurringProductForWithdrawal":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/UpdateOrderReviewedFlag":                        {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.OrderService/RetrieveListOfUniqueProductIDs":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.OrderService/GetLocationsForCreatingOrder":                   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/RetrieveListOfUniqueProductIDForBulkOrder":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.ExportService/ExportStudentBilling":                          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.ExportService/ExportMasterData":                              {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.FileService/UploadFile":                                      {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.FileService/GetEnrollmentFile":                               {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.CourseService/ImportStudentCourses":                          {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.OrderService/RetrieveListOfOrderAssociatedProductOfPackages": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/payment.v1.CourseService/ImportStudentClasses":                          {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.OrderService/RetrieveListOfProducts":                         {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/payment.v1.OrderService/RetrieveStudentEnrollmentStatusByLocation":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/RetrieveStudentEnrolledLocations":               {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/RetrieveRecurringProductsOfStudentInLocation":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.OrderService/GetOrgLevelStudentStatus":                       {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/payment.v1.CourseService/ManualUpsertStudentCourse":                     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
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

	authInterceptor, err := interceptors.NewAuth(
		ignoreAuthEndpoint,
		groupDecider,
		c.Issuers,
	)
	if err != nil {
		l.Panic("err init authInterceptor", zap.Error(err))
	}

	return authInterceptor
}
