package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	commands "github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WorkingHoursService struct {
	masterDB database.Ext
	bobDB    database.Ext

	WorkingHoursCommandHandler commands.WorkingHoursCommandHandler

	WorkingHoursRepo infrastructure.WorkingHoursRepo
	LocationRepo     infrastructure.LocationRepo
}

func NewWorkingHoursService(
	masterDB database.Ext,
	bobDB database.Ext,
	workingHoursRepo infrastructure.WorkingHoursRepo,
	locationRepo infrastructure.LocationRepo,
) *WorkingHoursService {
	return &WorkingHoursService{
		masterDB: masterDB,
		bobDB:    bobDB,
		WorkingHoursCommandHandler: commands.WorkingHoursCommandHandler{
			DB:               masterDB,
			WorkingHoursRepo: workingHoursRepo,
		},
		WorkingHoursRepo: workingHoursRepo,
		LocationRepo:     locationRepo,
	}
}

func (w *WorkingHoursService) ImportWorkingHours(ctx context.Context, req *mpb.ImportWorkingHoursRequest) (*mpb.ImportWorkingHoursResponse, error) {
	locations, errGetLocs := w.LocationRepo.GetChildLocations(ctx, w.bobDB, req.LocationId)

	if errGetLocs != nil {
		return nil, status.Error(codes.Internal, errGetLocs.Error())
	}

	locationIDs := sliceutils.Map(locations, func(l *location_domain.Location) string {
		return l.LocationID
	})

	if len(locationIDs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "location not found")
	}

	config := validators.CSVImportConfig[commands.ImportWorkingHoursCsvFields]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "day",
				Required: true,
			},
			{
				Column:   "opening_time",
				Required: true,
			},
			{
				Column:   "closing_time",
				Required: true,
			},
		},
		Transform: transformCSVFileToData,
	}

	csvFile, csvErr := validators.ReadAndValidateCSV(req.Payload, config)

	if csvErr != nil {
		return nil, status.Error(codes.InvalidArgument, csvErr.Error())
	}

	rowErrors := sliceutils.MapSkip(csvFile, validators.GetErrorFromCSVValue[commands.ImportWorkingHoursCsvFields], validators.HasCSVErr[commands.ImportWorkingHoursCsvFields])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	workingHoursList := []*domain.WorkingHours{}
	errs := []*mpb.ImportWorkingHoursResponse_ImportWorkingHoursError{}
	now := time.Now()

	for index, row := range csvFile {
		rowNumber := index + 2

		workingHoursBuilder := domain.NewWorkingHoursBuilder().
			WithWorkingHoursRepo(w.WorkingHoursRepo).
			WithDay(row.Value.Day).
			WithOpeningTime(row.Value.OpeningTime).
			WithClosingTime(row.Value.ClosingTime).
			WithModificationTime(now, now)

		workingHours, buildWeekErr := workingHoursBuilder.BuildWithoutPKCheck()

		if buildWeekErr != nil {
			icErr := &mpb.ImportWorkingHoursResponse_ImportWorkingHoursError{
				RowNumber: int32(rowNumber),
				Error:     buildWeekErr.Error(),
			}
			errs = append(errs, icErr)
			continue
		}

		workingHoursList = append(workingHoursList, workingHours)
	}

	if len(workingHoursList) == 0 && len(errs) == 0 {
		return &mpb.ImportWorkingHoursResponse{}, status.Error(codes.InvalidArgument, "no data in csv file")
	}

	payload := commands.ImportWorkingHoursPayload{WorkingHours: workingHoursList}
	err := w.WorkingHoursCommandHandler.ImportWorkingHoursTx(ctx, payload, locationIDs)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mpb.ImportWorkingHoursResponse{}, nil
}

func compareTimeString(time1, time2 string) (int, error) {
	errs := []error{}

	time1 = strings.ReplaceAll(time1, ":", "")
	time2 = strings.ReplaceAll(time2, ":", "")
	time1Int, err1 := strconv.Atoi(time1)
	time2Int, err2 := strconv.Atoi(time2)

	if err1 != nil || time1 == "" {
		errs = append(errs, fmt.Errorf("%s", "Invalid format opening time"))
	}

	if err2 != nil || time2 == "" {
		errs = append(errs, fmt.Errorf("%s", "Invalid format closing time"))
	}

	if len(errs) > 0 {
		return 0, errs[0]
	}

	if time1Int > time2Int {
		return 1, nil
	} else if time1Int < time2Int {
		return -1, nil
	}
	return 0, nil
}

func transformCSVFileToData(s []string) (*commands.ImportWorkingHoursCsvFields, error) {
	result := &commands.ImportWorkingHoursCsvFields{}
	errs := []error{}

	const (
		Day = iota
		OpeningTime
		ClosingTime
	)

	if !utf8.ValidString(s[Day]) {
		errs = append(errs, fmt.Errorf("%s", "day is not a valid UTF8 string"))
	}

	result.Day = s[Day]

	compareValue, errTime := compareTimeString(s[OpeningTime], s[ClosingTime])

	if errTime != nil {
		errs = append(errs, fmt.Errorf("%v", errTime))
	}

	if compareValue > 0 {
		errs = append(errs, fmt.Errorf("%s", "Invalid format opening time is not less than closing time"))
	}

	result.OpeningTime = s[OpeningTime]
	result.ClosingTime = s[ClosingTime]

	if len(errs) > 0 {
		return result, errs[0]
	}

	return result, nil
}
