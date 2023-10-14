package support

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/app"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"google.golang.org/grpc"
)

type LocationConfigResolver struct {
	DB database.Ext

	LocationRepo interface {
		FindAccessPaths(ctx context.Context, db database.Ext, locationIDs []string) ([]string, error)
		FindRootIDs(ctx context.Context, db database.Ext) ([]string, error)
		FindLowestAccessPathByLocationIDs(ctx context.Context, db database.Ext, locationIDs []string) ([]string, map[string]string, error)
	}
	ExternalConfigurationService interface {
		GetConfigurationByKeysAndLocations(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error)
		GetConfigurationByKeysAndLocationsV2(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error)
	}
}

func (s *LocationConfigResolver) GetEnabledLocationConfigsByOrg(ctx context.Context, locationIDs []string) ([]tpb.ConversationType, []string, error) {
	var accessPaths []string
	var rootLocation string
	// Technical debt: This is a temporary fix (for restricted location and Chat Thread Launching P1)
	// Logic: bypass Security Filter to get Org LocationID
	// Brand, Center level
	if len(locationIDs) != 0 {
		accPathInDB, err := s.LocationRepo.FindAccessPaths(ctx, s.DB, locationIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("LocationRepo.FindAccessPaths %w", err)
		}
		if len(accPathInDB) != len(locationIDs) {
			return nil, nil, fmt.Errorf("finding access paths for locations %v only return %d items", locationIDs, len(accPathInDB))
		}
		accessPaths = accPathInDB

		firstAccessPath := accessPaths[0]
		separatedLocationsOfFirstAccessPath := strings.Split(firstAccessPath, "/")
		if len(separatedLocationsOfFirstAccessPath) == 0 {
			return nil, nil, fmt.Errorf("cannot get root location from access path")
		}

		rootLocation = separatedLocationsOfFirstAccessPath[0]
	} else {
		// Org level
		rootIDs, err := s.LocationRepo.FindRootIDs(ctx, s.DB)

		if err != nil {
			return nil, nil, fmt.Errorf("LocationRepo.FindRootIDs: %w", err)
		}
		if len(rootIDs) == 0 {
			return nil, nil, fmt.Errorf("cannot get root location from access path")
		}

		rootLocation = rootIDs[0]
	}

	studentConfigKey, parentConfigKey := app.GetLocationConfigKeys(false)

	excludeTypes := []tpb.ConversationType{}

	orgLocation := rootLocation

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return excludeTypes, nil, fmt.Errorf("GetOutgoingContext: %v", err)
	}
	res, err := s.ExternalConfigurationService.GetConfigurationByKeysAndLocations(mdCtx, &mpb.GetConfigurationByKeysAndLocationsRequest{
		Keys:         []string{studentConfigKey, parentConfigKey},
		LocationsIds: []string{orgLocation},
	})
	if err != nil {
		return excludeTypes, nil, fmt.Errorf("GetConfigurationByKeysAndLocations error: %v", err)
	}

	for _, config := range res.Configurations {
		enableChat, _ := strconv.ParseBool(config.ConfigValue)
		if !enableChat {
			if config.ConfigKey == studentConfigKey {
				excludeTypes = append(excludeTypes, tpb.ConversationType_CONVERSATION_STUDENT)
			} else if config.ConfigKey == parentConfigKey {
				excludeTypes = append(excludeTypes, tpb.ConversationType_CONVERSATION_PARENT)
			}
		}
	}

	return excludeTypes, accessPaths, nil
}

func (s *LocationConfigResolver) GetEnabledLocationConfigsByLocations(ctx context.Context, locationIDs []string, conversationTypes []tpb.ConversationType) (map[tpb.ConversationType][]string, error) {
	studentLocationConfigKey, parentLocationConfigKey := app.GetLocationConfigKeys(true)

	requestConfigKeys := []string{}
	for _, requestConvType := range conversationTypes {
		if requestConvType == tpb.ConversationType_CONVERSATION_STUDENT {
			requestConfigKeys = append(requestConfigKeys, studentLocationConfigKey)
		}
		if requestConvType == tpb.ConversationType_CONVERSATION_PARENT {
			requestConfigKeys = append(requestConfigKeys, parentLocationConfigKey)
		}
	}
	var (
		locationIDsToQuery []string
	)
	// admin request doesn't have locations
	if len(locationIDs) == 0 {
		rootIDs, err := s.LocationRepo.FindRootIDs(ctx, s.DB)
		if err != nil {
			return nil, fmt.Errorf("failed FindRootIDs: %+v", err)
		}

		if len(rootIDs) == 0 {
			return nil, fmt.Errorf("cannot find rootIDs")
		}

		locationIDsToQuery = rootIDs
	} else {
		// staff request always have locations
		locationIDsToQuery = locationIDs
	}

	lowestLocationIDs, lowestAccessPathsMap, err := s.LocationRepo.FindLowestAccessPathByLocationIDs(ctx, s.DB, locationIDsToQuery)
	if err != nil {
		return nil, fmt.Errorf("failed FindLowestAccessPathByLocationIDs: %+v", err)
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetOutgoingContext: %v", err)
	}

	res, err := s.ExternalConfigurationService.GetConfigurationByKeysAndLocationsV2(mdCtx, &mpb.GetConfigurationByKeysAndLocationsV2Request{
		Keys:        requestConfigKeys,
		LocationIds: lowestLocationIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("GetConfigurationByKeysAndLocationsV2 error: %v", err)
	}

	mapConversationTypeWithLocationEnabled := make(map[tpb.ConversationType][]string, len(res.GetConfigurations()))

	// filter enabled locations and it's access paths to be used in the ES
	// if location A doesn't have config exist, treat it as FALSE
	for _, locationConfig := range res.GetConfigurations() {
		locationID := locationConfig.GetLocationId()
		isEnabled, err := strconv.ParseBool(locationConfig.ConfigValue)
		if err != nil {
			return nil, fmt.Errorf("failed ParseBool: %+v", err)
		}
		if isEnabled {
			if accessPath, ok := lowestAccessPathsMap[locationID]; ok {
				var convType tpb.ConversationType
				if locationConfig.ConfigKey == studentLocationConfigKey {
					convType = tpb.ConversationType_CONVERSATION_STUDENT
				} else if locationConfig.ConfigKey == parentLocationConfigKey {
					convType = tpb.ConversationType_CONVERSATION_PARENT
				}
				mapConversationTypeWithLocationEnabled[convType] = append(mapConversationTypeWithLocationEnabled[convType], accessPath)
			} else {
				return nil, fmt.Errorf("locationID %s is not exist in mapped list", locationID)
			}
		}
	}

	return mapConversationTypeWithLocationEnabled, nil
}
