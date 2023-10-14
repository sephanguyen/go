package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/payment/repositories"

	"github.com/spf13/cobra"
)

func genPaymentRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"accounting_category":               &repositories.AccountingCategoryRepo{},
		"billing_schedule":                  &repositories.BillingScheduleRepo{},
		"billing_schedule_period":           &repositories.BillingSchedulePeriodRepo{},
		"discount":                          &repositories.DiscountRepo{},
		"tax":                               &repositories.TaxRepo{},
		"product_accounting_category":       &repositories.ProductAccountingCategoryRepo{},
		"product_grade":                     &repositories.ProductGradeRepo{},
		"fee":                               &repositories.FeeRepo{},
		"material":                          &repositories.MaterialRepo{},
		"package":                           &repositories.PackageRepo{},
		"product_price":                     &repositories.ProductPriceRepo{},
		"package_course":                    &repositories.PackageCourseRepo{},
		"product_location":                  &repositories.ProductLocationRepo{},
		"billing_ratio":                     &repositories.BillingRatioRepo{},
		"leaving_reason":                    &repositories.LeavingReasonRepo{},
		"order":                             &repositories.OrderRepo{},
		"order_item":                        &repositories.OrderItemRepo{},
		"bill_item":                         &repositories.BillItemRepo{},
		"product":                           &repositories.ProductRepo{},
		"location":                          &repositories.LocationRepo{},
		"users":                             &repositories.UserRepo{},
		"order_action_log":                  &repositories.OrderActionLogRepo{},
		"students":                          &repositories.StudentRepo{},
		"bill_item_course":                  &repositories.BillItemCourseRepo{},
		"order_item_course":                 &repositories.OrderItemCourseRepo{},
		"package_quantity_type_mapping":     &repositories.PackageQuantityTypeMappingRepo{},
		"package_course_material":           &repositories.PackageCourseMaterialRepo{},
		"package_course_fee":                &repositories.PackageCourseFeeRepo{},
		"student_product":                   &repositories.StudentProductRepo{},
		"student_associated_product":        &repositories.StudentAssociatedProductRepo{},
		"product_setting":                   &repositories.ProductSettingRepo{},
		"product_discount":                  &repositories.ProductDiscountRepo{},
		"student_course":                    &repositories.StudentCourseRepo{},
		"file":                              &repositories.FileRepo{},
		"grade":                             &repositories.GradeRepo{},
		"student_package":                   &repositories.StudentPackageRepo{},
		"student_package_access_path":       &repositories.StudentPackageAccessPathRepo{},
		"course_access_path":                &repositories.CourseAccessPathRepo{},
		"user_access_path":                  &repositories.UserAccessPathRepo{},
		"student_package_class":             &repositories.StudentPackageClassRepo{},
		"class":                             &repositories.ClassRepo{},
		"student_enrollment_status_history": &repositories.StudentEnrollmentStatusHistoryRepo{},
		"bill_item_account_category":        &repositories.BillItemAccountCategoryRepo{},
		"upcoming_bill_item":                &repositories.UpcomingBillItemRepo{},
		"order_leaving_reason":              &repositories.OrderLeavingReasonRepo{},
		"notification_date":                 &repositories.NotificationDateRepo{},
		"student_package_log":               &repositories.StudentPackageLogRepo{},
		"student_package_order":             &repositories.StudentPackageOrderRepo{},
		"user_discount_tag":                 &repositories.UserDiscountTagRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "payment", repos)
	return nil
}

func newGenPaymentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "payment [../../mock/payment]",
		Short: "generate payment repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genPaymentRepo,
	}
}
