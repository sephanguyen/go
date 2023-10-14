package queries

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure/repo"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AcademicCalendarQueryHandler struct {
	MasterDB              database.Ext
	BobDB                 database.Ext
	AcademicYearRepo      infrastructure.AcademicYearRepo
	AcademicWeekRepo      infrastructure.AcademicWeekRepo
	AcademicClosedDayRepo infrastructure.AcademicClosedDayRepo
	LocationRepo          infrastructure.LocationRepo
	LocationTypeRepo      infrastructure.LocationTypeRepo
	ConfigRepo            infrastructure.ConfigRepo
}

func (a *AcademicCalendarQueryHandler) GetLocationsByLocationTypeLevelConfig(ctx context.Context) ([]*location_domain.Location, error) {
	const locationTypeLevelKey = "mastermgmt.academic_calendar.location_type_level"
	configInfo, err := a.ConfigRepo.GetByKey(ctx, a.MasterDB, locationTypeLevelKey)

	if err != nil {
		return nil, fmt.Errorf("GetByKey: %s %w", locationTypeLevelKey, err)
	}

	locTypes, err := a.LocationTypeRepo.GetLocationTypesByLevel(ctx, a.BobDB, configInfo.ConfigValue)

	if err != nil {
		return nil, fmt.Errorf("GetLocationTypesByLevel: %w", err)
	}

	locTypeIDs := sliceutils.Map(locTypes, func(l *location_domain.LocationType) string {
		return l.LocationTypeID
	})

	locations, err := a.LocationRepo.GetLocationByLocationTypeIDs(ctx, a.BobDB, locTypeIDs)

	if err != nil {
		return nil, fmt.Errorf("GetLocationByLocationTypeIDs: %w", err)
	}

	return locations, nil
}

func (a *AcademicCalendarQueryHandler) GetLocationsForAcademic(ctx context.Context, academicYearID string) ([]*location_domain.Location, error) {
	locations, err := a.GetLocationsByLocationTypeLevelConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("GetLocationsByLocationTypeLevelConfig: %w", err)
	}

	exLocs, err := a.AcademicWeekRepo.GetLocationsByAcademicWeekID(ctx, a.MasterDB, academicYearID)

	if err != nil {
		return nil, fmt.Errorf("GetLocationsByAcademicWeekID: %w", err)
	}

	result := []*location_domain.Location{}

	for _, location := range locations {
		if !slices.Contains(exLocs, location.LocationID) {
			result = append(result, location)
		}
	}

	return result, nil
}

func (a *AcademicCalendarQueryHandler) ExportAcademicCalendar(ctx context.Context, academicYearID string, locationID string) (data []byte, err error) {
	allLocations, err := a.LocationRepo.GetChildLocations(ctx, a.BobDB, locationID)
	if err != nil {
		return nil, fmt.Errorf("a.LocationRepo.GetChildLocations failed: %w", err)
	}

	academicYear, err := a.AcademicYearRepo.GetAcademicYearByID(ctx, a.MasterDB, academicYearID)

	if err != nil {
		return nil, fmt.Errorf("a.AcademicYearRepo.GetAcademicYearByID failed: %w", err)
	}

	mapLocations := make(map[string]string, 0)
	locationIDs := make([]string, 0, len(allLocations))

	for _, location := range allLocations {
		mapLocations[location.LocationID] = location.Name
		locationIDs = append(locationIDs, location.LocationID)
	}

	allLocationIDs := golibs.Uniq(locationIDs)

	weeks, err := a.AcademicWeekRepo.GetAcademicWeeksByYearAndLocationIDs(ctx, a.MasterDB, academicYearID, allLocationIDs)
	if err != nil {
		return nil, fmt.Errorf("a.AcademicWeekRepo.GetAcademicWeeksByYearAndLocationIDs failed: %w", err)
	}

	weekIDs := golibs.Uniq(sliceutils.Map(weeks, func(w *repo.AcademicWeek) string {
		return w.AcademicWeekID.String
	}))

	closedDays, err := a.AcademicClosedDayRepo.GetAcademicClosedDayByWeeks(ctx, a.MasterDB, weekIDs)
	if err != nil {
		return nil, fmt.Errorf("a.AcademicWeekRepo.GetAcademicClosedDayByWeeks failed: %w", err)
	}
	academicCalendar := a.getAcademicCalendar(weeks, closedDays, academicYear.Name, mapLocations)

	exportableAC := sliceutils.Map(academicCalendar, func(a *repo.AcademicCalendar) database.Entity {
		return a
	})

	ec := []exporter.ExportColumnMap{
		{
			DBColumn:  "academic_week_id",
			CSVColumn: "academic_week_id",
		},
		{
			DBColumn:  "week_order",
			CSVColumn: "order",
		},
		{
			DBColumn:  "name",
			CSVColumn: "name",
		},
		{
			DBColumn:  "start_date",
			CSVColumn: "start_date",
		},
		{
			DBColumn:  "end_date",
			CSVColumn: "end_date",
		},
		{
			DBColumn:  "period",
			CSVColumn: "period",
		},
		{
			DBColumn:  "academic_closed_day",
			CSVColumn: "academic_closed_day",
		},
		{
			DBColumn:  "academic_year",
			CSVColumn: "academic_year",
		},
		{
			DBColumn:  "location",
			CSVColumn: "location",
		},
	}

	str, err := exporter.ExportBatch(exportableAC, ec)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return exporter.ToCSV(str), nil
}

func (a *AcademicCalendarQueryHandler) getAcademicCalendar(weeks []*repo.AcademicWeek, closedDays []*repo.AcademicClosedDay, academicYear string, mapLocations map[string]string) []*repo.AcademicCalendar {
	list := make([]*repo.AcademicCalendar, 0, len(weeks))

	for _, w := range weeks {
		ac := &repo.AcademicCalendar{}
		ac.AcademicYear = database.Text(academicYear)
		ac.AcademicWeekID = w.AcademicWeekID
		ac.Location = database.Text(mapLocations[w.LocationID.String])
		ac.WeekOrder = database.Text(strconv.Itoa(int(w.WeekOrder.Int)))
		ac.Name = w.Name
		ac.StartDate = w.StartDate
		ac.EndDate = w.EndDate
		ac.Period = w.Period

		var arrClosedDaysStr []string

		for _, c := range closedDays {
			if c.AcademicWeekID == w.AcademicWeekID {
				arrClosedDaysStr = append(arrClosedDaysStr, c.Date.Time.Format("2006-01-02"))
			}
		}
		ac.AcademicClosedDays = database.Text(strings.Join(arrClosedDaysStr, ";"))

		list = append(list, ac)
	}

	return list
}
