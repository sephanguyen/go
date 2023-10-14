package services

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/jackc/pgtype"
)

type ImportMasterDataService struct {
	pb.UnimplementedImportMasterDataServiceServer
	DB              database.Ext
	DiscountTagRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.DiscountTag) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.DiscountTag) error
	}
	ProductGroupRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.ProductGroup) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.ProductGroup) error
		GetByID(ctx context.Context, db database.QueryExecer, productGroupID string) (entities.ProductGroup, error)
	}
	ProductGroupMappingRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, productGroupID pgtype.Text, e []*entities.ProductGroupMapping) error
	}
	PackageDiscountSettingRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, e []*entities.PackageDiscountSetting) error
	}
	PackageDiscountCourseMappingRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, e []*entities.PackageDiscountCourseMapping) error
	}
}

func checkMandatoryColumnAndGetIndex(column []string, positions []int) (bool, int) {
	for _, position := range positions {
		if strings.TrimSpace(column[position]) == "" {
			return false, position
		}
	}
	return true, 0
}
