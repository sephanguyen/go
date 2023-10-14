package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type MasterDataPort interface {
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*domain.Location, error)
	GetCourseByID(ctx context.Context, db database.Ext, id string) (*domain.Course, error)
	GetCourseTeachingTimeByIDs(ctx context.Context, db database.Ext, ids []string) (map[string]*domain.Course, error)
	GetClassByID(ctx context.Context, db database.Ext, id string) (*domain.Class, error)
	GetLowestLocationsByPartnerInternalIDs(ctx context.Context, db database.Ext, ids []string) (map[string]*domain.Location, error)
	FindPermissionByNamesAndUserID(ctx context.Context, db database.QueryExecer, permissionName []string, userID string) (*domain.UserPermissions, error)
	CheckLocationByIDs(ctx context.Context, db database.Ext, ids []string, locationName map[string]string) error
}
