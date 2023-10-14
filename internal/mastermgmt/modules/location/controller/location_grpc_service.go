package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	location_commands "github.com/manabie-com/backend/internal/mastermgmt/modules/location/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/application/producers"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LocationManagementGRPCService struct {
	DB                         database.Ext
	LocationProducer           producers.LocationProducer
	LocationCommandHandler     location_commands.LocationCommandHandler
	LocationTypeCommandHandler location_commands.LocationTypeCommandHandler
	LocationRepo               infrastructure.LocationRepo
	LocationTypeRepo           infrastructure.LocationTypeRepo
	ImportLogRepo              infrastructure.ImportLogRepo
	GetLocationQueryHandler    queries.GetLocationQueryHandler
	UnleashClientIns           unleashclient.ClientInstance
	Env                        string
}

func NewLocationManagementGRPCService(
	db database.Ext,
	jsm nats.JetStreamManagement,
	locationRepo infrastructure.LocationRepo,
	locationTypeRepo infrastructure.LocationTypeRepo,
	importLogRepo infrastructure.ImportLogRepo,
	unleashClientIns unleashclient.ClientInstance,
	env string,
) *LocationManagementGRPCService {
	return &LocationManagementGRPCService{
		DB: db,
		LocationProducer: producers.LocationProducer{
			JSM: jsm,
		},
		LocationRepo:     locationRepo,
		LocationTypeRepo: locationTypeRepo,
		LocationCommandHandler: location_commands.LocationCommandHandler{
			DB:               db,
			LocationRepo:     locationRepo,
			LocationTypeRepo: locationTypeRepo,
		},
		LocationTypeCommandHandler: location_commands.LocationTypeCommandHandler{
			DB:               db,
			LocationTypeRepo: locationTypeRepo,
		},
		GetLocationQueryHandler: queries.GetLocationQueryHandler{
			DB:               db,
			LocationRepo:     locationRepo,
			UnleashClientIns: unleashClientIns,
			Env:              env,
		},
		ImportLogRepo:    importLogRepo,
		UnleashClientIns: unleashClientIns,
		Env:              env,
	}
}

func (l *LocationManagementGRPCService) ImportLocation(ctx context.Context, req *mpb.ImportLocationRequest) (*mpb.ImportLocationResponse, error) {
	v2Req := &mpb.ImportLocationV2Request{
		Payload: req.GetPayload(),
	}
	_, err := l.ImportLocationV2(ctx, v2Req)
	if err != nil {
		return nil, status.Error(codes.Internal, "入力した情報に誤りがあります。データを確認してください!")
	}
	return &mpb.ImportLocationResponse{}, err
}

func (l *LocationManagementGRPCService) ImportLocationType(ctx context.Context, req *mpb.ImportLocationTypeRequest) (*mpb.ImportLocationTypeResponse, error) {
	v2Req := &mpb.ImportLocationTypeV2Request{
		Payload: req.GetPayload(),
	}
	_, err := l.ImportLocationTypeV2(ctx, v2Req)
	if err != nil {
		return nil, status.Error(codes.Internal, "入力した情報に誤りがあります。データを確認してください!")
	}
	return &mpb.ImportLocationTypeResponse{}, err
}

type payloadParams struct {
	userId     string
	importType string
	payload    interface{}
}

func (l *LocationManagementGRPCService) ImportLocationV2(ctx context.Context, req *mpb.ImportLocationV2Request) (res *mpb.ImportLocationV2Response, err error) {
	config := validators.CSVImportConfig[domain.Location]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "partner_internal_id",
				Required: true,
			},
			{
				Column:   "name",
				Required: true,
			},
			{
				Column:   "location_type",
				Required: true,
			},
			{
				Column:   "partner_internal_parent_id",
				Required: false,
			},
		},
		Transform: transformCSVLineToLocation,
	}
	csvLocations, err := validators.ReadAndValidateCSV(req.Payload, config)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	markDuplicatedPartnerID(csvLocations)
	// collect error lines only
	rowErrors := sliceutils.MapSkip(csvLocations, validators.GetErrorFromCSVValue[domain.Location], validators.HasCSVErr[domain.Location])
	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}
	violate, err := l.hasLocationHierarchyViolation(ctx, csvLocations)

	if err != nil {
		if err.Error() == "mustImportAllExistData" {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("resources.masters.message.%s", err.Error()).Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if violate {
		rowErrors = sliceutils.MapSkip(csvLocations, validators.GetErrorFromCSVValue[domain.Location], validators.HasCSVErr[domain.Location])

		if len(rowErrors) > 0 {
			return nil, utils.GetValidationError(rowErrors)
		}
	}

	locations := make([]*domain.Location, len(csvLocations))
	for i, c := range csvLocations {
		locations[i] = mapLocationCSVtoLocation(i)(c)
	}
	// Keep the higher level of location type to be inserted first
	slices.SortStableFunc(locations, func(l1, l2 *domain.Location) bool {
		if l1.LocationTypeLevel < l2.LocationTypeLevel {
			return true
		} else if l1.LocationTypeLevel == l2.LocationTypeLevel {
			return l1.CSVPosition < l2.CSVPosition
		}
		return false
	})

	payload := location_commands.UpsertLocation{
		Locations: locations,
	}
	err = l.LocationCommandHandler.ImportLocationV2(ctx, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mpb.ImportLocationV2Response{}, nil
}

func (l *LocationManagementGRPCService) ImportLocationTypeV2(ctx context.Context, req *mpb.ImportLocationTypeV2Request) (res *mpb.ImportLocationTypeV2Response, err error) {
	config := validators.CSVImportConfig[domain.LocationType]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "name",
				Required: true,
			},
			{
				Column:   "display_name",
				Required: true,
			},
			{
				Column:   "level",
				Required: true,
			},
		},
		Transform: transformCSVLineToLocationType,
	}
	csvLocationTypes, err := validators.ReadAndValidateCSV(req.Payload, config)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	csvLocationTypes, _ = checkUnique(csvLocationTypes)
	csvLocationTypes, _ = checkLevelOrder(csvLocationTypes)

	// collect error lines only
	rowErrors := sliceutils.MapSkip(csvLocationTypes, validators.GetErrorFromCSVValue[domain.LocationType], validators.HasCSVErr[domain.LocationType])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	locationTypes := sliceutils.Map(csvLocationTypes, mapLocTypeCSVtoLocationType)
	payload := location_commands.ImportLocationTypeV2Payload{
		LocationTypes: locationTypes,
	}

	bErr := l.LocationTypeCommandHandler.ImportLocationTypes(ctx, payload)
	if bErr != nil {
		if bErr.Is("levelAlreadyExisted") || bErr.Is("levelSwapped") || bErr.Is("mustImportAllExistData") || bErr.Is("canNotUpdateLowestType") {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("resources.masters.message.%s", bErr.Name).Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mpb.ImportLocationTypeV2Response{}, nil
}

func mapLocTypeCSVtoLocationType(c *validators.CSVLineValue[domain.LocationType]) *domain.LocationType {
	return &domain.LocationType{
		Name:        c.Value.Name,
		Level:       c.Value.Level,
		DisplayName: c.Value.DisplayName,
		IsArchived:  c.Value.IsArchived,
	}
}

func transformCSVLineToLocationType(s []string) (*domain.LocationType, error) {
	lt := &domain.LocationType{}
	const (
		Name = iota
		DisplayName
		Level
	)

	errs := []error{}

	name := s[Name]
	if len(name) < 1 {
		errs = append(errs, fmt.Errorf("%s", "name can not be empty"))
	}
	if !utf8.ValidString(s[Name]) {
		errs = append(errs, fmt.Errorf("%s", "name is not a valid UTF8 string"))
	}
	if strings.ToLower(strings.TrimSpace(name)) == "org" {
		errs = append(errs, fmt.Errorf("%s", "can not import org"))
	} else {
		lt.Name = name
	}

	displayName := s[DisplayName]
	if len(displayName) < 1 {
		errs = append(errs, fmt.Errorf("%s", "display name can not be empty"))
	}
	if !utf8.ValidString(s[DisplayName]) {
		errs = append(errs, fmt.Errorf("%s", "display name is not a valid UTF8 string"))
	} else {
		lt.DisplayName = displayName
	}

	lv, err := strconv.Atoi(strings.TrimSpace(s[Level]))
	if err != nil {
		errs = append(errs, fmt.Errorf("level is not a number: %s", s[Level]))
	} else {
		lt.Level = lv
	}

	if len(errs) > 0 {
		return lt, errs[0]
	}

	return lt, nil
}

func checkUnique(locTypes []*validators.CSVLineValue[domain.LocationType]) ([]*validators.CSVLineValue[domain.LocationType], bool) {
	nameMap := make(map[string]*validators.CSVLineValue[domain.LocationType], len(locTypes))
	levelMap := make(map[int]*validators.CSVLineValue[domain.LocationType], len(locTypes))
	hasDuplication := false

	for i, g := range locTypes {
		v, ok := nameMap[g.Value.Name]
		if ok {
			if g.Error == nil {
				g.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     fmt.Sprintf("name %s is duplicated", v.Value.Name),
				}
				hasDuplication = true
			}
		} else {
			nameMap[g.Value.Name] = g
		}
		v, ok = levelMap[g.Value.Level]
		if ok {
			if g.Error == nil {
				g.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     fmt.Sprintf("level %d is duplicated", v.Value.Level),
				}
				hasDuplication = true
			}
		} else {
			levelMap[g.Value.Level] = g
		}
	}
	return locTypes, hasDuplication
}

// Level must be in sorted order
func checkLevelOrder(locTypes []*validators.CSVLineValue[domain.LocationType]) ([]*validators.CSVLineValue[domain.LocationType], bool) {
	wrongOrder := false
	minLevel := 0
	for i, v := range locTypes {
		if v.Value.Level < 1 && v.Error == nil {
			v.Error = &dto.UpsertError{
				RowNumber: int32(i + 2),
				Error:     "level must be greater than 0",
			}
			continue
		}

		if minLevel+1 == v.Value.Level {
			minLevel = v.Value.Level
		} else {
			wrongOrder = true
			if v.Error == nil {
				v.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     "level must be in sequential order",
				}
			}
		}
	}
	return locTypes, wrongOrder
}

func mapLocationCSVtoLocation(row int) func(c *validators.CSVLineValue[domain.Location]) *domain.Location {
	return func(c *validators.CSVLineValue[domain.Location]) *domain.Location {
		return &domain.Location{
			Name:                    c.Value.Name,
			PartnerInternalID:       c.Value.PartnerInternalID,
			PartnerInternalParentID: c.Value.PartnerInternalParentID,
			LocationType:            c.Value.LocationType,
			IsArchived:              c.Value.IsArchived,
			LocationID:              c.Value.LocationID,
			ParentLocationID:        c.Value.ParentLocationID,
			AccessPath:              c.Value.AccessPath,

			// For sorting, ensures insert order
			CSVPosition: row,
		}
	}
}

func transformCSVLineToLocation(s []string) (*domain.Location, error) {
	l := &domain.Location{}
	const (
		PartnerInternalID = iota
		Name
		LocationType
		PartnerInternalParentID // optional
	)

	errs := []error{}

	pID := s[PartnerInternalID]
	l.PartnerInternalID = pID

	locName := s[Name]
	if !utf8.ValidString(s[Name]) {
		errs = append(errs, fmt.Errorf("%s", "name is not a valid UTF8 string"))
	} else {
		l.Name = locName
	}

	locType := s[LocationType]
	l.LocationType = locType

	pParentID := s[PartnerInternalParentID]
	l.PartnerInternalParentID = pParentID

	if len(errs) > 0 {
		return l, errs[0]
	}

	return l, nil
}

// 3 rules:
// First still check existence of child/parent internal id
//
// higher location type levels can only be children of lower levels.
//
// Prevents circle relationship when input parent_location
func (l *LocationManagementGRPCService) hasLocationHierarchyViolation(ctx context.Context, locations []*validators.CSVLineValue[domain.Location]) (bool, error) {
	violate := false

	orgLocations, err := l.GetLocationQueryHandler.GetDefaultLocation(ctx)
	if err != nil || len(orgLocations) == 0 {
		return violate, fmt.Errorf("can not get default location: %s", err)
	}
	rootLocation := orgLocations[0].LocationID

	allLocTypes, err := l.LocationTypeRepo.RetrieveLocationTypes(ctx, l.DB)
	if err != nil {
		return violate, fmt.Errorf("can not get location types: %s", err.Error())
	}
	allLocations, err := l.LocationRepo.GetAllRawLocations(ctx, l.DB)

	if err != nil {
		return violate, fmt.Errorf("can not get locations: %s", err.Error())
	}

	locTypeNameMap := make(map[string]*domain.LocationType, len(allLocTypes))
	locTypeIDMap := make(map[string]*domain.LocationType, len(allLocTypes))
	locTypeMapByPartnerID := make(map[string]*domain.LocationType, len(allLocations))
	locationMapByPartnerID := make(map[string]*domain.Location, len(allLocations))
	newLocInternalMap := make(map[string]*validators.CSVLineValue[domain.Location], len(locations))

	for _, v := range allLocTypes {
		locTypeNameMap[v.Name] = v
		locTypeIDMap[v.LocationTypeID] = v
	}
	for _, v := range allLocations {
		locType, ok := locTypeIDMap[v.LocationType]
		if ok {
			locTypeMapByPartnerID[v.PartnerInternalID] = locType
		}
		locationMapByPartnerID[v.PartnerInternalID] = v
	}
	for _, v := range locations {
		newLocInternalMap[v.Value.PartnerInternalID] = v
	}
	idNumberChildMap := make(map[string]int)
	for i, v := range locations {
		idNumberChildMap[v.Value.PartnerInternalID] = countChildOfID(v.Value.PartnerInternalID, locations)
		locTypeName := v.Value.LocationType
		locInternalParentID := v.Value.PartnerInternalParentID
		locType, hasLocType := locTypeNameMap[locTypeName]
		if !hasLocType {
			v.Error = &dto.UpsertError{
				RowNumber: int32(i + 2),
				Error:     fmt.Sprintf("location type %s is not exist", locTypeName),
			}
			violate = true
			continue
		}

		var parentLocType *domain.LocationType
		if locInternalParentID == "" {
			// LocationType in Location is location_type_id, wrong naming
			parent, ok := locTypeIDMap[orgLocations[0].LocationType]
			if !ok {
				return violate, fmt.Errorf("default location does not have a location type")
			}
			parentLocType = parent
		} else {
			parentType, hasParent := locTypeMapByPartnerID[locInternalParentID]
			newParent, hasNewParent := newLocInternalMap[locInternalParentID]
			if !hasParent && !hasNewParent {
				v.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     fmt.Sprintf("partner internal parent id %s is not exist", locInternalParentID),
				}
				violate = true
				continue
			}
			// parent in DB
			if hasParent {
				parentLocType = parentType
			}
			// parent in CSV
			if hasNewParent {
				parentType, ok := locTypeNameMap[newParent.Value.LocationType]
				if ok {
					parentLocType = parentType
				}
				if !ok {
					v.Error = &dto.UpsertError{
						RowNumber: int32(i + 2),
						Error:     fmt.Sprintf("partner internal parent id %s is not exist", locInternalParentID),
					}
					violate = true
					continue
				}
			}
		}

		//  parent level could not be bigger than or equal child level
		if parentLocType.Level >= locType.Level {
			v.Error = &dto.UpsertError{
				RowNumber: int32(i + 2),
				Error: fmt.Sprintf("%s location level (%d) must be greater than parent.\n(parent internal id: %s, location type: %s, level: %d)",
					locTypeName, locType.Level, locInternalParentID, parentLocType.Name, parentLocType.Level),
			}
			violate = true
		}
	}

	if !violate {
		violate, err = checkNewRule(locations, allLocations, allLocTypes, idNumberChildMap)
		if err != nil {
			return violate, err
		}
	}

	if !violate {
		// bind location type id
		// because of wrong naming, but LocationType field is a foreign key (id)
		for _, v := range locations {
			locTypeName := v.Value.LocationType
			childType, hasLocType := locTypeNameMap[locTypeName]
			if hasLocType {
				v.Value.LocationType = childType.LocationTypeID
				v.Value.LocationTypeLevel = childType.Level
			}
			existingLoc, ok := locationMapByPartnerID[v.Value.PartnerInternalID]
			if ok {
				v.Value.LocationID = existingLoc.LocationID
			} else {
				v.Value.LocationID = idutil.ULIDNow()
			}
			// set parent id
			// case child of root
			if v.Value.PartnerInternalParentID == "" {
				v.Value.ParentLocationID = rootLocation
			} else {
				// other
				parentLoc, ok := locationMapByPartnerID[v.Value.PartnerInternalParentID]
				if ok {
					v.Value.ParentLocationID = parentLoc.LocationID
				} else {
					newParent, ok := newLocInternalMap[v.Value.PartnerInternalParentID]
					if ok {
						v.Value.ParentLocationID = newParent.Value.LocationID
					}
				}
			}
			// bypass access path, must update after importing in handler
			v.Value.AccessPath = rootLocation
		}
	}
	return violate, nil
}

func checkNewRule(locations []*validators.CSVLineValue[domain.Location], allLocations []*domain.Location, allLocTypes []*domain.LocationType, idNumberChildMap map[string]int) (bool, error) {
	internalIDMap := make(map[string]bool)
	internalIDMapParent := make(map[string]string)
	for _, v := range locations {
		internalIDMap[v.Value.PartnerInternalID] = true
	}
	for _, v := range allLocations {
		internalIDMapParent[v.PartnerInternalID] = v.PartnerInternalParentID
		if _, exists := internalIDMap[v.PartnerInternalID]; !exists && len(v.PartnerInternalID) > 0 {
			return true, fmt.Errorf("mustImportAllExistData")
		}
	}
	for i, v := range locations {
		if v.Value.LocationType != allLocTypes[len(allLocTypes)-1].Name && idNumberChildMap[v.Value.PartnerInternalID] == 0 {
			v.Error = &dto.UpsertError{
				RowNumber: int32(i + 2),
				Error:     "cannot import location which is parent having no child",
			}
			return true, nil
		}
		if _, ok := internalIDMapParent[v.Value.PartnerInternalID]; ok && internalIDMapParent[v.Value.PartnerInternalID] != v.Value.PartnerInternalParentID {
			v.Error = &dto.UpsertError{
				RowNumber: int32(i + 2),
				Error:     "cannot change parent of the location",
			}
			return true, nil
		}
	}
	return false, nil
}

func countChildOfID(partnerInternalID string, locations []*validators.CSVLineValue[domain.Location]) int {
	count := 0
	for _, v := range locations {
		if partnerInternalID == v.Value.PartnerInternalParentID {
			count++
		}
	}
	return count
}

func markDuplicatedPartnerID(locations []*validators.CSVLineValue[domain.Location]) {
	newLocInternalMap := make(map[string]*validators.CSVLineValue[domain.Location], len(locations))
	for i, v := range locations {
		pID := v.Value.PartnerInternalID
		_, ok := newLocInternalMap[pID]
		if ok {
			if v.Error == nil {
				v.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     fmt.Sprintf("partner internal id %s is duplicated", pID),
				}
			}
		} else {
			newLocInternalMap[pID] = v
		}
	}
}
