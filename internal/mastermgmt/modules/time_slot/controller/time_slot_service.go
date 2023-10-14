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
	commands "github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TimeSlotService struct {
	masterDB database.Ext
	bobDB    database.Ext

	TimeSlotCommandHandler commands.TimeSlotCommandHandler

	LocationRepo infrastructure.LocationRepo
}

func NewTimeSlotService(
	masterDB database.Ext,
	bobDB database.Ext,
	timeSlotRepo infrastructure.TimeSlotRepo,
	locationRepo infrastructure.LocationRepo,
) *TimeSlotService {
	return &TimeSlotService{
		masterDB: masterDB,
		bobDB:    bobDB,
		TimeSlotCommandHandler: commands.TimeSlotCommandHandler{
			DB:           masterDB,
			TimeSlotRepo: timeSlotRepo,
		},
		LocationRepo: locationRepo,
	}
}

func (tss *TimeSlotService) ImportTimeSlots(ctx context.Context, req *mpb.ImportTimeSlotRequest) (*mpb.ImportTimeSlotResponse, error) {
	locations, errGetLocs := tss.LocationRepo.GetChildLocations(ctx, tss.bobDB, req.LocationId)
	if errGetLocs != nil {
		return nil, status.Error(codes.Internal, errGetLocs.Error())
	}

	locationIDs := sliceutils.Map(locations, func(l *location_domain.Location) string {
		return l.LocationID
	})

	config := validators.CSVImportConfig[commands.ImportTimeSlotCsvFields]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "time_slot_internal_id",
				Required: true,
			},
			{
				Column:   "start_time",
				Required: true,
			},
			{
				Column:   "end_time",
				Required: true,
			},
		},
		Transform: transformTimeSlotCSVLine,
	}

	csvImportTimeSlot, errCsv := validators.ReadAndValidateCSV(req.Payload, config)

	if errCsv != nil {
		return nil, status.Error(codes.InvalidArgument, errCsv.Error())
	}

	rowErrors := sliceutils.MapSkip(csvImportTimeSlot, validators.GetErrorFromCSVValue[commands.ImportTimeSlotCsvFields], validators.HasCSVErr[commands.ImportTimeSlotCsvFields])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	timeSlots := []*domain.TimeSlot{}
	errs := []*mpb.ImportTimeSlotResponse_ImportTimeSlotError{}
	now := time.Now()

	for index, row := range csvImportTimeSlot {
		rowNumber := index + 2

		timeSlotBuilder := domain.NewTimeSlotBuilder().
			WithTimeSlotInternalID(row.Value.TimeSlotInternalID).
			WithStartTime(row.Value.StartTime).
			WithEndTime(row.Value.EndTime).
			WithModificationTime(now, now)

		timeSlot, buildTimeSlotErr := timeSlotBuilder.BuildWithoutPKCheck()

		if buildTimeSlotErr != nil {
			icErr := &mpb.ImportTimeSlotResponse_ImportTimeSlotError{
				RowNumber: int32(rowNumber),
				Error:     buildTimeSlotErr.Error(),
			}
			errs = append(errs, icErr)
			continue
		}

		timeSlots = append(timeSlots, timeSlot)
	}

	if len(timeSlots) == 0 && len(errs) == 0 {
		return &mpb.ImportTimeSlotResponse{}, status.Error(codes.InvalidArgument, "no data in csv file")
	}

	if len(errs) > 0 {
		return &mpb.ImportTimeSlotResponse{
			Errors: errs,
		}, status.Errorf(codes.InvalidArgument, "InvalidArgument")
	}

	var err error
	if len(timeSlots) > 0 {
		payload := commands.ImportTimeSlotPayload{TimeSlots: timeSlots}
		err = tss.TimeSlotCommandHandler.ImportTimeSlotTx(ctx, payload, locationIDs)
	}

	if err != nil {
		return &mpb.ImportTimeSlotResponse{
			Errors: errs,
		}, status.Errorf(codes.Internal, err.Error())
	}

	resp := &mpb.ImportTimeSlotResponse{
		Errors: errs,
	}

	return resp, nil
}

func compareTimeString(time1, time2 string) (int, error) {
	errs := []error{}

	time1 = strings.ReplaceAll(time1, ":", "")
	time2 = strings.ReplaceAll(time2, ":", "")
	time1Int, err1 := strconv.Atoi(time1)
	time2Int, err2 := strconv.Atoi(time2)

	if err1 != nil || time1 == "" {
		errs = append(errs, fmt.Errorf("%s", "Invalid format start time"))
	}

	if err2 != nil || time2 == "" {
		errs = append(errs, fmt.Errorf("%s", "Invalid format end time"))
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

func transformTimeSlotCSVLine(s []string) (*commands.ImportTimeSlotCsvFields, error) {
	result := &commands.ImportTimeSlotCsvFields{}
	const (
		TimeSlotInternalID = iota
		StartTime
		EndTime
	)

	errs := []error{}

	if !utf8.ValidString(s[TimeSlotInternalID]) {
		errs = append(errs, fmt.Errorf("%s", "TimeSlotInternalID is not a valid UTF8 string"))
	}
	result.TimeSlotInternalID = s[TimeSlotInternalID]

	compareValue, errTime := compareTimeString(s[StartTime], s[EndTime])

	if errTime != nil {
		errs = append(errs, fmt.Errorf("%v", errTime))
	}

	if compareValue > 0 {
		errs = append(errs, fmt.Errorf("%s", "Invalid format start time is not less than end time"))
	}

	result.StartTime = s[StartTime]
	result.EndTime = s[EndTime]

	if len(errs) > 0 {
		return result, errs[0]
	}

	return result, nil
}
