package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"

	"golang.org/x/exp/slices"
)

type GetLocationQueryHandler struct {
	DB               database.Ext
	LocationRepo     infrastructure.LocationRepo
	LocationTypeRepo infrastructure.LocationTypeRepo
	UnleashClientIns unleashclient.ClientInstance
	Env              string
}

func (g *GetLocationQueryHandler) GetLocationsByQuery(ctx context.Context, payload *GetLocations) ([]*domain.Location, error) {
	locations, err := g.LocationRepo.RetrieveLocations(ctx, g.DB, payload.FilterLocation)
	if err != nil {
		return nil, fmt.Errorf("LocationRepo.RetrieveLocations: %w", err)
	}
	return locations, nil
}

func (g *GetLocationQueryHandler) GetDefaultLocation(ctx context.Context) ([]*domain.Location, error) {
	return g.LocationRepo.GetLocationByLocationTypeName(ctx, g.DB, domain.DefaultLocationType)
}

func (g *GetLocationQueryHandler) GetBaseLocationsByQuery(ctx context.Context, payload *GetLocations) ([]*domain.Location, error) {
	locations, err := g.LocationRepo.RetrieveLocations(ctx, g.DB, payload.FilterLocation)
	if err != nil {
		return nil, fmt.Errorf("LocationRepo.RetrieveLocations: %w", err)
	}
	locations, err = g.generateUnauthorizedLocationV2(locations)
	if err != nil {
		return nil, fmt.Errorf(`generateUnauthorizedLocationV2: %v`, err)
	}
	return locations, nil
}

func (g *GetLocationQueryHandler) GetLocationsTree(ctx context.Context, payload *GetLocations) (string, error) {
	locations, err := g.GetBaseLocationsByQuery(ctx, payload)
	if err != nil {
		return "", fmt.Errorf(`GetBaseLocationsByQuery: %v`, err)
	}

	if len(locations) == 0 {
		userID := interceptors.UserIDFromContext(ctx)
		return "", fmt.Errorf(`GetLocationsTree: User ID [%s] does not have access to any location`, userID)
	}

	uniqueTypes := make(map[string]struct{})
	for _, loc := range locations {
		if len(loc.LocationType) > 0 {
			uniqueTypes[loc.LocationType] = struct{}{}
		}
	}
	uniqueTypeSlice := make([]string, 0, len(uniqueTypes))
	for locType := range uniqueTypes {
		uniqueTypeSlice = append(uniqueTypeSlice, locType)
	}
	locationTypes, err := g.LocationTypeRepo.GetLocationTypeByIDs(ctx, g.DB, database.TextArray(uniqueTypeSlice), false)
	if err != nil {
		return "", fmt.Errorf(`locationTypeRepo.RetrieveLocationTypes: %v`, err)
	}
	if len(locationTypes) < 1 {
		return "", fmt.Errorf("missing location type")
	}

	// sort the slice of location types
	sort.Slice(locationTypes, func(i, j int) bool {
		return locationTypes[i].Level < locationTypes[j].Level
	})
	lowestLocationType := locationTypes[len(locationTypes)-1]

	locTypeMap := make(map[string]*domain.LocationType, len(locationTypes))
	for _, locType := range locationTypes {
		locTypeMap[locType.LocationTypeID] = locType
	}

	jsonTree, err := g.buildLocationTree(locations, locTypeMap, lowestLocationType)
	if err != nil {
		return "", fmt.Errorf(`buildLocationTree: %v`, err)
	}

	return jsonTree, nil
}

// O(n log n)
func (g *GetLocationQueryHandler) generateUnauthorizedLocationV2(authorizedLocations []*domain.Location) ([]*domain.Location, error) {
	if len(authorizedLocations) == 0 {
		return authorizedLocations, nil
	}
	// add un-authorized locations
	visitedMap := make(map[string]*domain.Location)

	for _, l := range authorizedLocations {
		currentLocation := l

		path := strings.Split(currentLocation.AccessPath, "/")
		prev := 1
		// If location is not visited, or visited but just find a new Authorized one
		for visitedMap[currentLocation.LocationID] == nil || (visitedMap[currentLocation.LocationID].IsUnauthorized && prev == 1) {
			visitedMap[currentLocation.LocationID] = currentLocation
			parentLocation := visitedMap[currentLocation.ParentLocationID]
			// check whether parent is added (different branches):
			if parentLocation != nil {
				break
			}

			index := slices.Index(path, currentLocation.LocationID)
			if index < 0 {
				return nil, fmt.Errorf("could not find id: %s in access path", currentLocation.LocationID)
			}
			if index == 0 || len(path) == 1 {
				// meet root
				break
			}

			accessPath := strings.Join(path[:len(path)-prev], "/")
			grandID := ""
			// [Org, Brand] --> no Grandpa
			if index > 1 {
				grandID = path[index-2]
			}
			parentLocation = &domain.Location{
				LocationID:       currentLocation.ParentLocationID,
				Name:             "UnAuthorized",
				AccessPath:       accessPath,
				IsUnauthorized:   true,
				ParentLocationID: grandID,
			}
			currentLocation = parentLocation
			prev++
		}
	}

	allLocations := sliceutils.MapValuesToSlice(visitedMap)
	return allLocations, nil
}

// O(n log n)
func (g *GetLocationQueryHandler) buildLocationTree(locations []*domain.Location, locTypes map[string]*domain.LocationType, lowestLocType *domain.LocationType) (string, error) {
	// Map to keep track of each location's children
	childrenMap := make(map[string][]*domain.TreeLocation)

	var rootLocation *domain.TreeLocation
	// Add each location to its parent's list of children
	for _, l := range locations {
		isLowest := false

		// unauthorized location will not have location type
		if l.LocationType != "" {
			locType, ok := locTypes[l.LocationType]
			if !ok {
				return "", fmt.Errorf("could not find location type id: %s", l.LocationType)
			}
			if locType.Level == lowestLocType.Level {
				isLowest = true
			}
		}

		treeLoc := &domain.TreeLocation{
			LocationID:        l.LocationID,
			Name:              l.Name,
			LocationType:      l.LocationType,
			ParentLocationID:  l.ParentLocationID,
			PartnerInternalID: l.PartnerInternalID,
			AccessPath:        l.AccessPath,
			IsArchived:        l.IsArchived,
			IsUnauthorized:    l.IsUnauthorized,
			UpdatedAt:         l.UpdatedAt,
			CreatedAt:         l.CreatedAt,
			IsLowestLevel:     isLowest,
			Children:          []*domain.TreeLocation{},
		}
		// meet root
		path := strings.Split(l.AccessPath, "/")
		if l.ParentLocationID == "" || len(path) == 1 {
			rootLocation = treeLoc
			continue
		}

		childrenMap[l.ParentLocationID] = append(childrenMap[l.ParentLocationID], treeLoc)
	}

	// Build the tree by iterating over the locations and adding their children
	queue := []*domain.TreeLocation{rootLocation}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		children := childrenMap[node.LocationID]

		// sort by updated_at, created_at asc
		slices.SortFunc(children, func(l1, l2 *domain.TreeLocation) bool {
			if l1.UpdatedAt.Equal(l2.UpdatedAt) {
				return l1.CreatedAt.Before(l2.CreatedAt)
			}
			return l1.UpdatedAt.Before(l2.UpdatedAt)
		})

		node.Children = children
		queue = append(queue, node.Children...)
	}

	json, err := json.Marshal(rootLocation)
	if err != nil {
		return "", fmt.Errorf(`could not marshal location tree: %v`, err)
	}

	return string(json), nil
}
