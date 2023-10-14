package interceptors

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"

	"google.golang.org/grpc"
)

type LocationRestricted struct {
	methods      map[string]struct{}
	db           database.Ext
	locationRepo infrastructure.LocationRepo
}

type LocationIDRequest interface {
	GetLocationIds() []string
}

func NewLocationRestricted(methods map[string]struct{}, db database.Ext, locationRepo infrastructure.LocationRepo) *LocationRestricted {
	return &LocationRestricted{
		methods:      methods,
		db:           db,
		locationRepo: locationRepo,
	}
}

func (l *LocationRestricted) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if _, exist := l.methods[info.FullMethod]; !exist {
		return handler(ctx, req)
	}
	locationReq, ok := req.(LocationIDRequest)
	if !ok {
		return nil, fmt.Errorf("request does not have LocationIds")
	}

	locationIDs := locationReq.GetLocationIds()

	// If not send locationIDs. Will check org permission.
	// Else: Will check does use have permission with all locations params?
	if len(locationIDs) == 0 {
		rootLocation, err := l.locationRepo.GetRootLocation(ctx, l.db)

		if err != nil {
			return nil, fmt.Errorf("can not get root location: %w", err)
		}
		if len(rootLocation) == 0 {
			return nil, fmt.Errorf("permission denied: user is not granted org level")
		}
	} else {
		locations, err := l.locationRepo.GetLocationsByLocationIDs(ctx, l.db, database.TextArray(locationIDs), false)
		if err != nil {
			return nil, fmt.Errorf("can not get locations by ids: %w", err)
		}
		if len(locations) != len(locationIDs) {
			return nil, fmt.Errorf("permission denied: some locations of %v are not granted for user", locationIDs)
		}
	}

	return handler(ctx, req)
}
