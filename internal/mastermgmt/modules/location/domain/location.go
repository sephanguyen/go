package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type Location struct {
	LocationID              string
	Name                    string
	PartnerInternalID       string
	PartnerInternalParentID string
	LocationType            string
	ParentLocationID        string
	IsArchived              bool
	AccessPath              string
	UpdatedAt               time.Time
	CreatedAt               time.Time
	DeletedAt               *time.Time
	ResourcePath            string
	IsUnauthorized          bool

	// internal state
	Persisted         bool // true: location already exists in db
	Repo              LocationRepo
	TypeRepo          LocationTypeRepo
	LocationTypeLevel int
	CSVPosition       int
}

type UpsertErrors struct {
	RowNumber int32
	Error     string
}
type LocationBuilder struct {
	location *Location
}

func NewLocation() *LocationBuilder {
	return &LocationBuilder{
		location: &Location{},
	}
}

func (l *LocationBuilder) Build(ctx context.Context, db database.Ext, locationRoot string, locations []*Location) (*Location, error) {
	if err := l.location.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid location: %w", err)
	}
	if err := l.location.PrepareData(ctx, db, locationRoot, locations); err != nil {
		return nil, fmt.Errorf("can prepare data: %w", err)
	}

	return l.location, nil
}

func (l *Location) PrepareData(ctx context.Context, db database.Ext, locationRoot string, locations []*Location) error {
	// partner_internal_id
	location, err := l.Repo.GetLocationByPartnerInternalID(ctx, db, l.PartnerInternalID)
	if err != nil {
		return fmt.Errorf("getLocationByPartnerInternalID err %s: %w", l.PartnerInternalID, err)
	}
	if location != nil {
		l.Persisted = true
		l.LocationID = location.LocationID
	} else {
		l.Persisted = false
		l.LocationID = idutil.ULIDNow()
	}

	// parent internal id
	l.ParentLocationID = locationRoot
	parentType := "org"
	if len(l.PartnerInternalParentID) > 0 {
		parent, err := l.Repo.GetLocationByPartnerInternalID(ctx, db, l.PartnerInternalParentID)
		if err != nil {
			return fmt.Errorf("can not get parent of location %s: %w", l.PartnerInternalParentID, err)
		}
		// if not exist in DB, we will get by ID just generate of this parent
		if parent != nil {
			l.ParentLocationID = parent.LocationID
		} else {
			parent = getParentByInternalID(l.PartnerInternalParentID, locations)
			if parent == nil {
				return fmt.Errorf("miss parent location: %w", err)
			}
			l.ParentLocationID = parent.LocationID
		}
		typeParent, err := l.TypeRepo.GetLocationTypeByID(ctx, db, parent.LocationType)
		if err != nil {
			return fmt.Errorf("cant not get location type %s: %w", parent.LocationType, err)
		}
		parentType = typeParent.Name
	}
	// check relationship between 2 location is satisfy rule of location type
	_, err = l.TypeRepo.GetLocationTypeByNameAndParent(ctx, db, l.LocationType, parentType)
	if err != nil {
		return fmt.Errorf("location type invalid: %w", err)
	}
	// location_type
	locationType, err := l.TypeRepo.GetLocationTypeByName(ctx, db, l.LocationType, false)
	if err != nil {
		return fmt.Errorf("getLocationTypeByName err: %w", err)
	}
	l.LocationType = locationType.LocationTypeID
	// check duplicate
	isDuplicate := checkDuplicateLocation(l.PartnerInternalID, locations)
	if isDuplicate {
		return fmt.Errorf("location %s is duplicated", l.PartnerInternalID)
	}
	// HACK bypass AC to update access path
	l.AccessPath = locationRoot
	return nil
}

func checkDuplicateLocation(id string, locations []*Location) bool {
	for _, location := range locations {
		if strings.EqualFold(location.PartnerInternalID, id) {
			return true
		}
	}
	return false
}

func getParentByInternalID(id string, locations []*Location) *Location {
	for _, location := range locations {
		if strings.EqualFold(location.PartnerInternalID, id) {
			return location
		}
	}
	return nil
}

func (l *LocationBuilder) BuildDefault(ctx context.Context, db database.Ext) (*Location, error) {
	if err := l.location.IsValidDefault(); err != nil {
		return nil, fmt.Errorf("invalid location: %w", err)
	}
	return l.location, nil
}

func (l *LocationBuilder) WithLocationRepo(repo LocationRepo) *LocationBuilder {
	l.location.Repo = repo
	return l
}

func (l *LocationBuilder) WithLocationTypeRepo(typeRepo LocationTypeRepo) *LocationBuilder {
	l.location.TypeRepo = typeRepo
	return l
}

func (l *LocationBuilder) WithPartnerInternalID(id string) *LocationBuilder {
	l.location.PartnerInternalID = id
	return l
}

func (l *LocationBuilder) WithName(name string) *LocationBuilder {
	l.location.Name = name
	return l
}

func (l *LocationBuilder) WithPartnerInternalParentID(id string) *LocationBuilder {
	l.location.PartnerInternalParentID = id
	return l
}

func (l *LocationBuilder) WithLocationType(locationType string) *LocationBuilder {
	l.location.LocationType = locationType
	return l
}

func (l *LocationBuilder) WithLocationTypeIDOfDefault(ctx context.Context, db database.Ext, locationTypeID string) (*LocationBuilder, error) {
	locations, err := l.location.Repo.GetLocationByLocationTypeID(ctx, db, locationTypeID)
	if err != nil {
		return l, fmt.Errorf("locationRepo.GetLocationByLocationTypeID err: %w", err)
	}
	// just get first location from locationType default org
	if len(locations) > 0 {
		l.location.LocationID = locations[0].LocationID
		l.location.AccessPath = locations[0].LocationID
		l.location.LocationType = locations[0].LocationType
	} else {
		l.location.LocationType = locationTypeID
	}
	return l, nil
}

func (l *LocationBuilder) WithModificationTime(createdAt, updatedAt time.Time) *LocationBuilder {
	l.location.CreatedAt = createdAt
	l.location.UpdatedAt = updatedAt
	return l
}

func (l *LocationBuilder) WithIsArchived(isArchived bool) *LocationBuilder {
	l.location.IsArchived = isArchived
	return l
}

func (l *LocationBuilder) WithLocationID(locationID string) *LocationBuilder {
	l.location.LocationID = locationID
	return l
}

func (l *LocationBuilder) WithAccessPath(accessPath string) *LocationBuilder {
	l.location.AccessPath = accessPath
	return l
}

func (l *LocationBuilder) GetLocation() *Location {
	return l.location
}

func (l *Location) IsValid() error {
	if len(l.PartnerInternalID) == 0 {
		return fmt.Errorf("location.PartnerInternalID cannot be empty")
	}

	if len(l.Name) == 0 {
		return fmt.Errorf("location.Name cannot be empty")
	}

	if l.PartnerInternalID == l.PartnerInternalParentID {
		return fmt.Errorf("location.PartnerInternalID cannot same Location.PartnerInternalParentID")
	}

	if l.UpdatedAt.Before(l.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}

	if len(l.LocationType) == 0 {
		return fmt.Errorf("location.LocationType cannot be empty")
	}

	if strings.EqualFold(l.LocationType, DefaultLocationType) {
		return fmt.Errorf("can not import default location type")
	}

	if !utf8.ValidString(l.Name) {
		return fmt.Errorf("location name is not in a valid utf8 format")
	}

	return nil
}

func (l *Location) IsValidDefault() error {
	if len(l.LocationID) == 0 {
		return fmt.Errorf("location.LocationID cannot be empty")
	}

	if len(l.Name) == 0 {
		return fmt.Errorf("location.Name cannot be empty")
	}

	if l.UpdatedAt.Before(l.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}

	if len(l.LocationType) == 0 {
		return fmt.Errorf("location.LocationType cannot be empty")
	}

	return nil
}

type FilterLocation struct {
	IncludeIsArchived bool
	UserID            string
}

func (l *Location) String() string {
	return fmt.Sprintf("id: %s, name: %s, type:%s, partner id: %s, archived: %v;", l.LocationID, l.Name, l.LocationType, l.PartnerInternalID, l.IsArchived)
}
