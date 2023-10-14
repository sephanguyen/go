package controller

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	commands "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/application/commands"
	queries "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AcademicYearService struct {
	masterDB database.Ext
	bobDB    database.Ext

	AcademicCalendarCommandHandler commands.AcademicCalendarCommandHandler
	AcademicCalendarQueryHandler   queries.AcademicCalendarQueryHandler

	AcademicYearRepo      infrastructure.AcademicYearRepo
	AcademicWeekRepo      infrastructure.AcademicWeekRepo
	AcademicClosedDayRepo infrastructure.AcademicClosedDayRepo
	LocationRepo          infrastructure.LocationRepo
}

func NewAcademicYearService(
	masterDB database.Ext,
	bobDB database.Ext,
	academicYearRepo infrastructure.AcademicYearRepo,
	academicWeekRepo infrastructure.AcademicWeekRepo,
	academicClosedDayRepo infrastructure.AcademicClosedDayRepo,
	locationRepo infrastructure.LocationRepo,
	locationTypeRepo infrastructure.LocationTypeRepo,
	configRepo infrastructure.ConfigRepo,
) *AcademicYearService {
	return &AcademicYearService{
		masterDB: masterDB,
		bobDB:    bobDB,
		AcademicCalendarCommandHandler: commands.AcademicCalendarCommandHandler{
			DB:                    masterDB,
			AcademicWeekRepo:      academicWeekRepo,
			AcademicClosedDayRepo: academicClosedDayRepo,
		},
		AcademicCalendarQueryHandler: queries.AcademicCalendarQueryHandler{
			MasterDB:              masterDB,
			BobDB:                 bobDB,
			AcademicWeekRepo:      academicWeekRepo,
			LocationRepo:          locationRepo,
			LocationTypeRepo:      locationTypeRepo,
			ConfigRepo:            configRepo,
			AcademicClosedDayRepo: academicClosedDayRepo,
			AcademicYearRepo:      academicYearRepo,
		},
		AcademicYearRepo:      academicYearRepo,
		AcademicWeekRepo:      academicWeekRepo,
		AcademicClosedDayRepo: academicClosedDayRepo,
		LocationRepo:          locationRepo,
	}
}

func (a *AcademicYearService) ImportAcademicCalendar(ctx context.Context, req *mpb.ImportAcademicCalendarRequest) (*mpb.ImportAcademicCalendarResponse, error) {
	locations, errGetLocs := a.LocationRepo.GetChildLocations(ctx, a.bobDB, req.LocationId)

	if errGetLocs != nil {
		return nil, status.Error(codes.Internal, errGetLocs.Error())
	}

	locationIDs := sliceutils.Map(locations, func(l *location_domain.Location) string {
		return l.LocationID
	})

	academicYearID := req.AcademicYearId
	academicClosedDayArrayReq := req.AcademicClosedDays

	errValidateAcademicYear := a.validateAcademicYear(ctx, a.masterDB, academicYearID)
	if errValidateAcademicYear != nil {
		return nil, status.Error(codes.InvalidArgument, errValidateAcademicYear.Error())
	}

	config := validators.CSVImportConfig[commands.ImportAcademicCalendarCsvFields]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "order",
				Required: true,
			},
			{
				Column:   "academic_week",
				Required: true,
			},
			{
				Column:   "start_date",
				Required: true,
			},
			{
				Column:   "end_date",
				Required: true,
			},
			{
				Column:   "period",
				Required: true,
			},
			{
				Column:   "academic_closed_day",
				Required: false,
			},
		},
		Transform: transformAcademicCalendarCSVLine,
	}

	csvAcademicCalendar, errCsv := validators.ReadAndValidateCSV(req.Payload, config)

	if errCsv != nil {
		return nil, status.Error(codes.InvalidArgument, errCsv.Error())
	}

	rowErrors := sliceutils.MapSkip(csvAcademicCalendar, validators.GetErrorFromCSVValue[commands.ImportAcademicCalendarCsvFields], validators.HasCSVErr[commands.ImportAcademicCalendarCsvFields])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	weeks := []*domain.AcademicWeek{}
	academicClosedDays := []*domain.AcademicClosedDay{}
	errs := []*mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{}
	now := time.Now()

	for index, row := range csvAcademicCalendar {
		rowNumber := index + 2
		weekStartDate := row.Value.AcademicWeekStartDate
		weekEndDate := row.Value.AcademicWeekEndDate

		for _, locationID := range locationIDs {
			academicWeekBuilder := domain.NewAcademicWeekBuilder().
				WithAcademicWeekRepo(a.AcademicWeekRepo).
				WithAcademicYearID(academicYearID).
				WithLocationID(locationID).
				WithWeekOrder(row.Value.Order).
				WithName(row.Value.AcademicWeekName).
				WithStartDate(weekStartDate).
				WithEndDate(weekEndDate).
				WithPeriod(row.Value.Period).
				WithModificationTime(now, now)

			week, buildWeekErr := academicWeekBuilder.Build()

			if buildWeekErr != nil {
				icErr := &mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{
					RowNumber: int32(rowNumber),
					Error:     buildWeekErr.Error(),
				}
				errs = append(errs, icErr)
				break
			}

			isBreak := false
			// import closed day from academic_closed_day column in csv file
			for _, academicClosedDay := range row.Value.AcademicClosedDays {
				if academicClosedDay.Before(weekStartDate) || academicClosedDay.After(weekEndDate) {
					icErr := &mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{
						RowNumber: int32(rowNumber),
						Error:     "Closed day of week must be day of week",
					}
					errs = append(errs, icErr)
					isBreak = true
					break
				}

				academicClosedDayBuilder := domain.NewAcademicClosedDayBuilder().
					WithAcademicClosedDayRepo(a.AcademicClosedDayRepo).
					WithAcademicWeekID(week.AcademicWeekID).
					WithAcademicYearID(academicYearID).
					WithLocationID(locationID).
					WithDate(academicClosedDay).
					WithModificationTime(now, now)

				academicClosedDay, buildAcademicClosedDayErr := academicClosedDayBuilder.Build(true)

				if buildAcademicClosedDayErr != nil {
					icErr := &mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{
						RowNumber: int32(rowNumber),
						Error:     buildAcademicClosedDayErr.Error(),
					}
					errs = append(errs, icErr)
					isBreak = true
					break
				}
				academicClosedDays = append(academicClosedDays, academicClosedDay)
			}

			if isBreak {
				break
			}

			weeks = append(weeks, week)
		}
	}

	// import closed day from request, it's gap of weeks, be filtered and validated from FE
	for _, academicClosedDayStrReq := range academicClosedDayArrayReq {
		for _, locationID := range locationIDs {
			if academicClosedDayStrReq == "" {
				break
			}
			academicClosedDayTime, errAcademicClosedDay := time.Parse("2006-01-02", academicClosedDayStrReq)

			if errAcademicClosedDay != nil {
				icErr := &mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{
					RowNumber: 0,
					Error:     errAcademicClosedDay.Error(),
				}
				errs = append(errs, icErr)
				break
			}

			academicClosedDayBuilder := domain.NewAcademicClosedDayBuilder().
				WithAcademicClosedDayRepo(a.AcademicClosedDayRepo).
				WithAcademicYearID(academicYearID).
				WithLocationID(locationID).
				WithDate(academicClosedDayTime).
				WithModificationTime(now, now)

			academicClosedDay, buildAcademicClosedDayErr := academicClosedDayBuilder.Build(false)

			if buildAcademicClosedDayErr != nil {
				icErr := &mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{
					RowNumber: 0,
					Error:     buildAcademicClosedDayErr.Error(),
				}
				errs = append(errs, icErr)
				break
			}
			academicClosedDays = append(academicClosedDays, academicClosedDay)
		}
	}

	if len(weeks) == 0 && len(errs) == 0 {
		return &mpb.ImportAcademicCalendarResponse{}, status.Error(codes.InvalidArgument, "no data in csv file")
	}

	if len(errs) > 0 {
		return &mpb.ImportAcademicCalendarResponse{
			Errors: errs,
		}, status.Errorf(codes.InvalidArgument, "InvalidArgument")
	}

	var err error
	if len(weeks) > 0 {
		payload := commands.ImportAcademicCalendarPayload{AcademicWeeks: weeks, AcademicClosedDays: academicClosedDays}
		err = a.AcademicCalendarCommandHandler.ImportAcademicCalendarTx(ctx, payload)
	}

	if err != nil {
		return &mpb.ImportAcademicCalendarResponse{
			Errors: errs,
		}, status.Errorf(codes.Internal, err.Error())
	}

	resp := &mpb.ImportAcademicCalendarResponse{
		Errors: nil,
	}

	return resp, nil
}

func transformAcademicCalendarCSVLine(s []string) (*commands.ImportAcademicCalendarCsvFields, error) {
	result := &commands.ImportAcademicCalendarCsvFields{}
	const (
		Order = iota
		AcademicWeekName
		StartDate
		EndDate
		Period
		AcademicClosedDays
	)

	errs := []error{}

	weekOrder, err := strconv.Atoi(s[Order])

	if err != nil {
		errs = append(errs, fmt.Errorf("%s", "order is not a valid integer"))
	}

	result.Order = weekOrder

	if !utf8.ValidString(s[AcademicWeekName]) {
		errs = append(errs, fmt.Errorf("%s", "academic week name is not a valid UTF8 string"))
	}
	result.AcademicWeekName = s[AcademicWeekName]

	weekStartDate, errStareDate := time.Parse("2006-01-02", s[StartDate])
	weekEndDate, errEndDate := time.Parse("2006-01-02", s[EndDate])
	if errStareDate != nil || s[StartDate] == "" {
		errs = append(errs, fmt.Errorf("%s", "Invalid format start_date"))
	}

	if errEndDate != nil || s[EndDate] == "" {
		errs = append(errs, fmt.Errorf("%s", "Invalid format end_date"))
	}

	if weekEndDate.Before(weekStartDate) {
		errs = append(errs, fmt.Errorf("%s", "Invalid format end_date"))
	}

	result.AcademicWeekStartDate = weekStartDate
	result.AcademicWeekEndDate = weekEndDate

	if !utf8.ValidString(s[Period]) {
		errs = append(errs, fmt.Errorf("%s", "period is not a valid UTF8 string"))
	}
	result.Period = s[Period]

	academicClosedDayRes := []time.Time{}

	if s[AcademicClosedDays] != "" {
		academicClosedDayArr := strings.Split(s[AcademicClosedDays], ";")

		for _, academicClosedDayStr := range academicClosedDayArr {
			academicClosedDayTime, errAcademicClosedDay := time.Parse("2006-01-02", academicClosedDayStr)
			if errAcademicClosedDay != nil {
				errs = append(errs, fmt.Errorf("%s", "Invalid format academic_closed_day"))
				break
			}
			academicClosedDayRes = append(academicClosedDayRes, academicClosedDayTime)
		}
	}

	result.AcademicClosedDays = academicClosedDayRes

	if len(errs) > 0 {
		return result, errs[0]
	}

	return result, nil
}

func (a *AcademicYearService) validateAcademicYear(ctx context.Context, masterDB database.Ext, id string) error {
	_, err := a.AcademicYearRepo.GetAcademicYearByID(ctx, masterDB, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("invalid")
	} else if err != nil {
		return fmt.Errorf("failed to get academic year by id: %w", err)
	}
	return nil
}

func convertDomainToLocationForAcademicRes(locations []*location_domain.Location, importedLocs []string) (res []*mpb.RetrieveLocationsForAcademicResponse_Location) {
	// sort by updated_at, created_at asc
	slices.SortFunc(locations, func(l1, l2 *location_domain.Location) bool {
		if l1.UpdatedAt.Equal(l2.UpdatedAt) {
			return l1.CreatedAt.Before(l2.CreatedAt)
		}
		return l1.UpdatedAt.Before(l2.UpdatedAt)
	})

	locationsResp := sliceutils.Map(locations, func(l *location_domain.Location) *mpb.RetrieveLocationsForAcademicResponse_Location {
		return &mpb.RetrieveLocationsForAcademicResponse_Location{
			LocationId: l.LocationID,
			Name:       l.Name,
			IsImported: sliceutils.Contains(importedLocs, l.LocationID),
		}
	})

	return locationsResp
}

func (a *AcademicYearService) RetrieveLocationsForAcademic(ctx context.Context, req *mpb.RetrieveLocationsForAcademicRequest) (*mpb.RetrieveLocationsForAcademicResponse, error) {
	academicYearID := req.AcademicYearId

	errs := a.validateAcademicYear(ctx, a.masterDB, academicYearID)

	if errs != nil {
		return nil, status.Error(codes.InvalidArgument, errs.Error())
	}

	locations, err := a.AcademicCalendarQueryHandler.GetLocationsByLocationTypeLevelConfig(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	importedLocs, err := a.AcademicWeekRepo.GetLocationsByAcademicWeekID(ctx, a.masterDB, academicYearID)

	if err != nil {
		return nil, fmt.Errorf("GetLocationsByAcademicWeekID: %w", err)
	}

	resp := &mpb.RetrieveLocationsForAcademicResponse{
		Locations: convertDomainToLocationForAcademicRes(locations, importedLocs),
	}

	return resp, nil
}

func convertDomainToLocationByLevelConfigRes(locations []*location_domain.Location) (res []*mpb.RetrieveLocationsByLocationTypeLevelConfigResponse_Location) {
	// sort by updated_at, created_at asc
	slices.SortFunc(locations, func(l1, l2 *location_domain.Location) bool {
		if l1.UpdatedAt.Equal(l2.UpdatedAt) {
			return l1.CreatedAt.Before(l2.CreatedAt)
		}
		return l1.UpdatedAt.Before(l2.UpdatedAt)
	})

	locationsResp := sliceutils.Map(locations, func(l *location_domain.Location) *mpb.RetrieveLocationsByLocationTypeLevelConfigResponse_Location {
		return &mpb.RetrieveLocationsByLocationTypeLevelConfigResponse_Location{
			LocationId: l.LocationID,
			Name:       l.Name,
		}
	})

	return locationsResp
}

func (a *AcademicYearService) RetrieveLocationsByLocationTypeLevelConfig(ctx context.Context, _ *mpb.RetrieveLocationsByLocationTypeLevelConfigRequest) (*mpb.RetrieveLocationsByLocationTypeLevelConfigResponse, error) {
	locations, err := a.AcademicCalendarQueryHandler.GetLocationsByLocationTypeLevelConfig(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &mpb.RetrieveLocationsByLocationTypeLevelConfigResponse{
		Locations: convertDomainToLocationByLevelConfigRes(locations),
	}

	return resp, nil
}

func (a *AcademicYearService) ExportAcademicCalendar(ctx context.Context, req *mpb.ExportAcademicCalendarRequest) (res *mpb.ExportAcademicCalendarResponse, err error) {
	csv, err := a.AcademicCalendarQueryHandler.ExportAcademicCalendar(ctx, req.AcademicYearId, req.LocationId)
	if err != nil {
		return nil, err
	}
	return &mpb.ExportAcademicCalendarResponse{
		Data: csv,
	}, nil
}
