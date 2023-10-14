package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type LocationType struct {
	LocationTypeID       string
	Name                 string
	DisplayName          string
	ParentName           string
	ParentLocationTypeID string
	Level                int
	IsArchived           bool
	UpdatedAt            time.Time
	CreatedAt            time.Time
	DeletedAt            *time.Time

	// internal state
	Persisted    bool // true: location already exists in db
	Repo         LocationTypeRepo
	LocationRepo LocationRepo
}

type LocationTypeBuilder struct {
	locationType *LocationType
}

const DefaultLocationType string = "org"

func NewLocationType() *LocationTypeBuilder {
	return &LocationTypeBuilder{
		locationType: &LocationType{},
	}
}

func (l *LocationTypeBuilder) Build(ctx context.Context, db database.Ext) (*LocationType, error) {
	if err := l.locationType.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid location type: %w", err)
	}
	return l.locationType, nil
}

func (l *LocationTypeBuilder) WithLocationTypeRepo(repo LocationTypeRepo) *LocationTypeBuilder {
	l.locationType.Repo = repo
	return l
}

func (l *LocationTypeBuilder) WithLocationRepo(repo LocationRepo) *LocationTypeBuilder {
	l.locationType.LocationRepo = repo
	return l
}

func (l *LocationTypeBuilder) WithIsArchived() *LocationTypeBuilder {
	l.locationType.IsArchived = false
	return l
}

func (l *LocationTypeBuilder) WithName(ctx context.Context, db database.Ext, name string) (*LocationTypeBuilder, error) {
	if name == "" {
		return l, fmt.Errorf("Name is null")
	}
	locationType, err := l.locationType.Repo.GetLocationTypeByName(ctx, db, name, true)
	if err != nil {
		return l, fmt.Errorf("GetLocationTypeByName: %w", err)
	}
	l.locationType.Name = name
	if locationType != nil {
		l.locationType.Persisted = true
		l.locationType.LocationTypeID = locationType.LocationTypeID
		l.locationType.ParentName = locationType.ParentName
	} else {
		l.locationType.Persisted = false
		l.locationType.LocationTypeID = idutil.ULIDNow()
		l.WithModificationTime(time.Now(), time.Now())
	}
	return l, nil
}

func (l *LocationTypeBuilder) WithValidParentName(ctx context.Context, db database.Ext, name, parentName string) (*LocationTypeBuilder, error) {
	if l.locationType.Persisted && l.locationType.ParentName != parentName {
		locations, err := l.locationType.LocationRepo.GetLocationByLocationTypeName(ctx, db, l.locationType.Name)
		if err != nil {
			return l, fmt.Errorf("LocationRepo.GetLocationByLocationTypeName: %w", err)
		}
		if len(locations) > 0 {
			return l, fmt.Errorf("locations with type %s exist", l.locationType.Name)
		}
	}
	locationType, err := l.locationType.Repo.GetLocationTypeByParentName(ctx, db, parentName)
	if err != nil {
		return l, fmt.Errorf("locationType.GetLocationTypeByParentName: %w", err)
	}
	if locationType != nil && locationType.Name != name {
		return l, fmt.Errorf("The child of parent name already exists")
	}
	return l, nil
}

func (l *LocationTypeBuilder) WithNameDefault(ctx context.Context, db database.Ext, name string) (*LocationTypeBuilder, error) {
	locationType, err := l.locationType.Repo.GetLocationTypeByName(ctx, db, name, true)
	if err != nil {
		return l, fmt.Errorf("LocationTypeRepo.getLocationTypeByName err: %w", err)
	}
	if locationType != nil {
		l.locationType.LocationTypeID = locationType.LocationTypeID
	}
	l.locationType.Name = name

	return l, nil
}

func (l *LocationTypeBuilder) WithDisplayName(name string) *LocationTypeBuilder {
	l.locationType.DisplayName = name
	return l
}

func (l *LocationTypeBuilder) WithLevel(level int) *LocationTypeBuilder {
	l.locationType.Level = level
	return l
}

func (l *LocationTypeBuilder) WithModificationTime(createdAt, updatedAt time.Time) *LocationTypeBuilder {
	l.locationType.CreatedAt = createdAt
	l.locationType.UpdatedAt = updatedAt
	return l
}

func (l *LocationTypeBuilder) WithLocationTypeID(locationTypeID string) *LocationTypeBuilder {
	l.locationType.LocationTypeID = locationTypeID
	return l
}

func (l *LocationTypeBuilder) GetLocationType() *LocationType {
	return l.locationType
}

func (l *LocationType) IsValid() error {
	if len(l.LocationTypeID) == 0 {
		return fmt.Errorf("LocationType.LocationID cannot be empty")
	}

	if len(l.Name) == 0 {
		return fmt.Errorf("LocationType.Name cannot be empty")
	}

	if len(l.DisplayName) == 0 {
		return fmt.Errorf("LocationType.DisplayName cannot be empty")
	}

	if l.UpdatedAt.Before(l.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}

	return nil
}

func (l *LocationType) String() string {
	return fmt.Sprintf("ID:%s, Name:%s, DisplayName:%s, Level:%d;", l.LocationTypeID, l.Name, l.DisplayName, l.Level)
}
