package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/service/validator"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ICourseLocationScheduleService interface {
	ImportCourseLocationSchedule(ctx context.Context, req *lpb.ImportCourseLocationScheduleRequest) (res *lpb.ImportCourseLocationScheduleResponse, err error)
	ExportCourseLocationSchedule(ctx context.Context) (*lpb.ExportCourseLocationScheduleResponse, error)
}

type CourseLocationScheduleService struct {
	wrapperConnection                 *support.WrapperDBConnection
	courseLocationScheduleServiceRepo infrastructure.CourseLocationScheduleRepo
}

func NewCourseLocationScheduleService(wrapperConnection *support.WrapperDBConnection, courseLocationScheduleServiceRepo infrastructure.CourseLocationScheduleRepo) *CourseLocationScheduleService {
	return &CourseLocationScheduleService{
		wrapperConnection:                 wrapperConnection,
		courseLocationScheduleServiceRepo: courseLocationScheduleServiceRepo,
	}
}

func ConvertErrToImportCSVErr(errors map[int]error) []*lpb.ImportCourseLocationScheduleResponse_ImportError {
	errorCSVs := []*lpb.ImportCourseLocationScheduleResponse_ImportError{}
	for line, err := range errors {
		errorCSVs = append(errorCSVs, &lpb.ImportCourseLocationScheduleResponse_ImportError{
			RowNumber: int32(line),
			Error:     fmt.Sprintf("unable to parse this item: %s", err),
		})
	}
	return errorCSVs
}

func (z *CourseLocationScheduleService) buildImportCourseLocationScheduleArgs(ctx context.Context, data []byte) ([]*domain.CourseLocationSchedule, *lpb.ImportCourseLocationScheduleResponse, error) {
	sc := scanner.NewCSVScanner(bytes.NewReader(data))
	columnsIndex := map[string]int{
		domain.IDLabel:                  0,
		domain.CourseIDLabel:            1,
		domain.LocationIDLabel:          2,
		domain.AcademicWeekLabel:        3,
		domain.ProductTypeScheduleLabel: 4,
		domain.FrequencyLabel:           5,
		domain.TotalNoLessonLabel:       6,
	}

	mapErrors := ValidateImportFileHeader(sc, columnsIndex)
	if len(mapErrors) > 0 {
		return nil, &lpb.ImportCourseLocationScheduleResponse{Errors: ConvertErrToImportCSVErr(mapErrors)}, nil
	}
	errors := make([]*lpb.ImportCourseLocationScheduleResponse_ImportError, 0, len(data)+1)

	dataImport := []*domain.CourseLocationSchedule{}
	locationIds := make([]string, 0)
	for sc.Scan() {
		academicWeeks, err := support.ConvertAcademicWeeks(sc.Text(domain.AcademicWeekLabel))
		arrError := make([]string, 0, 8)
		if len(academicWeeks) == 0 || err != nil {
			arrError = append(arrError, "academic_week is required")
		}
		freq, err := support.ConvertStringToPointInt(sc.Text(domain.FrequencyLabel))
		if err != nil {
			arrError = append(arrError, err.Error())
		}
		totalNoLesson, err := support.ConvertStringToPointInt(sc.Text(domain.TotalNoLessonLabel))
		if err != nil {
			arrError = append(arrError, err.Error())
		}
		productTypeSchedule, ok := domain.ProductTypeScheduleMap[sc.Text(domain.ProductTypeScheduleLabel)]
		if !ok {
			arrError = append(arrError, "Product Type Schedule not valid")
		} else {
			v, err := validator.GetValidator(&validator.ParamValidationProductTypeSchedule{ProductType: productTypeSchedule, TotalNoLesson: totalNoLesson, Freq: freq})
			if err != nil {
				arrError = append(arrError, err.Error())
			}
			strErr := v.Validation()

			if strErr != "" {
				arrError = append(arrError, strErr)
			}
			now := time.Now()
			locationID := sc.Text(domain.LocationIDLabel)
			ID := sc.Text(domain.IDLabel)
			courseID := sc.Text(domain.CourseIDLabel)
			courseLocationSchedule, err := domain.NewCourseLocationScheduleBuilder().
				WithID(ID).
				WithCourseID(courseID).
				WithLocationID(locationID).
				WithAcademicWeek(academicWeeks).
				WithProductTypeSchedule(productTypeSchedule).
				WithFrequency(freq).
				WithTotalNoLesson(totalNoLesson).
				WithCreateAt(&now).
				WithUpdatedAt(&now).
				Build()
			if err != nil {
				arrError = append(arrError, err.Error())
			}
			if ID != "" {
				locationIds = append(locationIds, locationID)
			}
			dataImport = append(dataImport, courseLocationSchedule)
		}

		if len(arrError) > 0 {
			icErr := &lpb.ImportCourseLocationScheduleResponse_ImportError{
				RowNumber: int32(sc.GetCurRow()) - 1,
				Error:     strings.Join(arrError, ", "),
			}
			errors = append(errors, icErr)
			continue
		}
	}
	if len(errors) == 0 && len(locationIds) > 0 {
		conn, err := z.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
		if err != nil {
			return nil, nil, err
		}
		mapLocationWeekValid, err := z.courseLocationScheduleServiceRepo.GetAcademicWeekValid(ctx, conn, locationIds, time.Now())
		if err != nil {
			return nil, nil, err
		}
		for index, data := range dataImport {
			if data.Persisted {
				for _, week := range data.AcademicWeeks {
					key := fmt.Sprintf("%s-%s", data.LocationID, week)
					_, ok := mapLocationWeekValid[key]
					if !ok {
						icErr := &lpb.ImportCourseLocationScheduleResponse_ImportError{
							RowNumber: int32(index + 1),
							Error:     fmt.Sprintf("LocationId %s - Week order %s Invalid", data.LocationID, week),
						}
						errors = append(errors, icErr)
						break
					}
				}
			}
		}
	}

	return dataImport, &lpb.ImportCourseLocationScheduleResponse{Errors: errors}, nil
}

func ValidateImportFileHeader(sc scanner.CSVScanner, columnsIndex map[string]int) map[int]error {
	totalRows := len(sc.GetRow())
	errors := make(map[int]error)
	currentRow := 1

	if totalRows == 0 {
		errors[currentRow] = fmt.Errorf("request payload empty")
	}
	if totalRows < len(columnsIndex) {
		errors[currentRow] = fmt.Errorf("invalid format: number of column should be greater than or equal %d", len(columnsIndex))
	}
	for colName, colIndex := range columnsIndex {
		if i, ok := sc.Head[colName]; !ok || i != colIndex && colIndex != -1 {
			errors[currentRow] = fmt.Errorf("invalid format: the column have index %d (toLowerCase) should be '%s'", colIndex, colName)
		}
	}
	return errors
}

func (z *CourseLocationScheduleService) ImportCourseLocationSchedule(ctx context.Context, req *lpb.ImportCourseLocationScheduleRequest) (res *lpb.ImportCourseLocationScheduleResponse, err error) {
	dataImport, dataError, err := z.buildImportCourseLocationScheduleArgs(ctx, req.Payload)

	if err != nil {
		return nil, err
	}

	if len(dataError.Errors) == 0 {
		conn, err := z.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
		if err != nil {
			return nil, err
		}
		errImport := z.courseLocationScheduleServiceRepo.UpsertMultiCourseLocationSchedule(ctx, conn, dataImport)
		if errImport != nil {
			e := errImport.Err
			msgErr := e.Error()
			dataIndexErr := errImport.Index
			dataError := dataImport[dataIndexErr]
			if errors.Is(e, domain.ErrUniqCourseLocationSchedule) {
				msgErr = fmt.Sprintf("Duplicate CourseID: %s - LocationID: %s", dataError.CourseID, dataError.LocationID)
			}
			if errors.Is(e, domain.ErrNotExistsFKCourseLocationSchedule) {
				msgErr = fmt.Sprintf("Not Exists CourseID: %s - LocationID: %s in course_access_path table", dataError.CourseID, dataError.LocationID)
			}
			return &lpb.ImportCourseLocationScheduleResponse{Errors: []*lpb.ImportCourseLocationScheduleResponse_ImportError{
				{
					RowNumber: int32(dataIndexErr + 1),
					Error:     msgErr,
				},
			}}, nil
		}
	}
	return dataError, nil
}

func (z *CourseLocationScheduleService) ExportCourseLocationSchedule(ctx context.Context) (*lpb.ExportCourseLocationScheduleResponse, error) {
	conn, err := z.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	data, err := z.courseLocationScheduleServiceRepo.ExportCourseLocationSchedule(ctx, conn)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	dataExport := [][]string{}

	for _, courseLocationSchedule := range data {
		freqStr := ""
		totalNoLesson := ""
		if courseLocationSchedule.Frequency != nil {
			freqP := courseLocationSchedule.Frequency
			if *freqP != 0 {
				freqStr = fmt.Sprint(*freqP)
			}
		}
		if courseLocationSchedule.TotalNoLesson != nil {
			totalNoLessonP := courseLocationSchedule.TotalNoLesson
			if *totalNoLessonP != 0 {
				totalNoLesson = fmt.Sprint(*totalNoLessonP)
			}
		}
		line := []string{
			courseLocationSchedule.ID, courseLocationSchedule.CourseID, courseLocationSchedule.LocationID,
			strings.Join(courseLocationSchedule.AcademicWeeks, "_"), domain.MapStringToProductTypeScheduleNumber[string(courseLocationSchedule.ProductTypeSchedule)],
			freqStr, totalNoLesson,
		}
		dataExport = append(dataExport, line)
	}

	title := []string{domain.IDLabel, domain.CourseIDLabel, domain.LocationIDLabel, domain.AcademicWeekLabel, domain.ProductTypeScheduleLabel, domain.FrequencyLabel, domain.TotalNoLessonLabel}
	csvData := append([][]string{title}, dataExport...)

	return &lpb.ExportCourseLocationScheduleResponse{
		Data: exporter.ToCSV(csvData),
	}, nil
}
