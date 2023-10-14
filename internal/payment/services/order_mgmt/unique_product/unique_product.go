package unique_product

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	packageService "github.com/manabie-com/backend/internal/payment/services/domain_service/package"
	studentProduct "github.com/manabie-com/backend/internal/payment/services/domain_service/student_product"
)

type IStudentProductServiceForUniqueProduct interface {
	GetUniqueProductsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (studentProductOfUniqueProducts []*entities.StudentProduct, err error)
	EndDateOfUniqueRecurringProduct(ctx context.Context, db database.QueryExecer, productID string, endTime time.Time) (endTimeOfUniqueProduct time.Time, err error)
	GetUniqueProductsByStudentIDs(ctx context.Context, db database.QueryExecer, studentID []string) (mapStudentIDAndStudentProducts map[string][]*entities.StudentProduct, err error)
}

type IPackageServiceForUniqueProduct interface {
	GetByIDForUniqueProduct(ctx context.Context, db database.Ext, packageID string) (packageEntities entities.Package, err error)
}

type UniqueProduct struct {
	DB database.Ext

	StudentProductService IStudentProductServiceForUniqueProduct
	PackageService        IPackageServiceForUniqueProduct
}

func NewUniqueProduct(db database.Ext) (uniqueProduct *UniqueProduct) {
	return &UniqueProduct{
		DB:                    db,
		StudentProductService: studentProduct.NewStudentProductService(),
		PackageService:        packageService.NewPackageService(),
	}
}
