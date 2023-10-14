package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/discount/repositories"
	discountSvc "github.com/manabie-com/backend/internal/discount/services/domain_service"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genDiscountRepo(_ *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"bill_item":                       &repositories.BillItemRepo{},
		"discount":                        &repositories.DiscountRepo{},
		"product_group":                   &repositories.ProductGroupRepo{},
		"product_group_mapping":           &repositories.ProductGroupMappingRepo{},
		"student_discount_tracker":        &repositories.StudentDiscountTrackerRepo{},
		"student_product":                 &repositories.StudentProductRepo{},
		"user_discount_tag":               &repositories.UserDiscountTagRepo{},
		"student_parent":                  &repositories.StudentParentRepo{},
		"discount_tag":                    &repositories.DiscountTagRepo{},
		"package_discount_setting":        &repositories.PackageDiscountSettingRepo{},
		"package_discount_course_mapping": &repositories.PackageDiscountCourseMappingRepo{},
		"user":                            &repositories.UserRepo{},
		"order_item":                      &repositories.OrderItemRepo{},
		"product":                         &repositories.ProductRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "discount", repos)

	structs := map[string][]interface{}{
		"internal/discount/services/domain_service/discount_event":   {&discountSvc.DiscountEventService{}},
		"internal/discount/services/domain_service/discount_tag":     {&discountSvc.DiscountTagService{}},
		"internal/discount/services/domain_service/discount_tracker": {&discountSvc.DiscountTrackerService{}},
		"internal/discount/services/domain_service/product_group":    {&discountSvc.ProductGroupService{}},
		"internal/discount/services/domain_service/student_product":  {&discountSvc.StudentProductService{}},
		"internal/discount/services/domain_service/student_sibling":  {&discountSvc.StudentSiblingService{}},
	}
	return tools.GenMockStructs(structs)
}

func newGenDiscountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "discount [../../mock/discount]",
		Short: "generate discount repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genDiscountRepo,
	}
}
