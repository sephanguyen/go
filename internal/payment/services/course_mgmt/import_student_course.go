package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *CourseMgMt) ImportStudentCourses(ctx context.Context, req *pb.ImportStudentCoursesRequest) (res *pb.ImportStudentCoursesResponse, err error) {
	var (
		errors                                       []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError
		eventMessages                                []*npb.EventStudentPackage
		mapStudentCourseRowsWithStudentID            map[string][]utils.ImportedStudentCourseRow
		studentIDs                                   []string
		courseIDs                                    []string
		mapLocationAccessWithStudentID               map[string]interface{}
		mapLocationAccessWithCourseID                map[string]interface{}
		mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath
	)

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	header := lines[0]
	headerTitles := []string{
		"student_id",
		"course_id",
		"location_id",
		"start_date",
		"end_date",
	}

	err = utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}
	studentIDs, courseIDs, mapStudentCourseRowsWithStudentID, errors = toMapStudentWithStudentPackage(lines)
	if len(errors) > 0 {
		res = &pb.ImportStudentCoursesResponse{}
		res.Errors = errors
		return
	}

	mapLocationAccessWithStudentID, err = s.StudentService.GetMapLocationAccessStudentByStudentIDs(ctx, s.DB, studentIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get access path from student ids with error %v", err.Error())
		return
	}

	mapLocationAccessWithCourseID, err = s.CourseService.GetMapLocationAccessCourseForCourseIDs(ctx, s.DB, courseIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get access path from course ids with error %v", err.Error())
		return
	}

	mapStudentCourseWithStudentPackageAccessPath, err = s.StudentPackage.GetMapStudentCourseWithStudentPackageIDByIDs(ctx, s.DB, studentIDs)
	if err != nil {
		return
	}

	res = &pb.ImportStudentCoursesResponse{}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		eventMessages, errors = s.validateStudentIDAndUpsertStudentPackage(
			ctx,
			tx,
			mapLocationAccessWithStudentID,
			mapLocationAccessWithCourseID,
			mapStudentCourseWithStudentPackageAccessPath,
			mapStudentCourseRowsWithStudentID)
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error when importing student course: %s", err.Error())
		res.Errors = errors
		return res, nil
	}

	err = s.SubscriptionService.PublishStudentPackage(ctx, eventMessages)
	if err != nil {
		log.Printf("Error when publish student package: %s", err.Error())
	}

	return res, nil
}

func toMapStudentWithStudentPackage(rows [][]string) (
	studentIDs []string,
	courseIDs []string,
	mapStudentCourseRowsWithStudentID map[string][]utils.ImportedStudentCourseRow,
	errors []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError,
) {
	mapStudentCourseRowsWithStudentID = make(map[string][]utils.ImportedStudentCourseRow)
	mapStudentCourseUnique := make(map[string]bool)
	mapStudentIDUnique := make(map[string]bool)
	mapCourseIDUnique := make(map[string]bool)
	headerTitles := []string{
		"student_id",
		"course_id",
		"location_id",
		"start_date",
		"end_date",
	}
	for i, line := range rows[1:] {
		rowIndex := int32(i) + 2
		var (
			studentPackage           *entities.StudentPackages
			studentPackageAccessPath *entities.StudentPackageAccessPath
			eventMessage             *npb.EventStudentPackage
			courseID                 string
			err                      error
		)
		studentPackage, studentPackageAccessPath, eventMessage, courseID, err = toStudentPackageAndStudentPackageAccessPathsEntityFromCsv(line, headerTitles)
		if err != nil {
			errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
				RowNumber: rowIndex, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf(constant.UnableToParseStudentCourse, err),
			})
			continue
		}
		if _, ok := mapCourseIDUnique[courseID]; !ok {
			courseIDs = append(courseIDs, courseID)
			mapCourseIDUnique[courseID] = true
		}

		if _, ok := mapStudentIDUnique[studentPackage.StudentID.String]; !ok {
			studentIDs = append(studentIDs, studentPackage.StudentID.String)
			mapStudentIDUnique[studentPackage.StudentID.String] = true
		}

		keyStudentCourse := fmt.Sprintf("%s_%s_%s", studentPackage.StudentID.String, courseID, studentPackageAccessPath.LocationID.String)
		if ok := mapStudentCourseUnique[keyStudentCourse]; ok {
			errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
				RowNumber: rowIndex, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("duplicate student course with student id %s and course id %s", studentPackage.StudentID.String, courseID),
			})
			continue
		}
		mapStudentCourseUnique[keyStudentCourse] = true

		if importedStudentCourseRows, ok := mapStudentCourseRowsWithStudentID[studentPackage.StudentID.String]; ok {
			mapStudentCourseRowsWithStudentID[studentPackage.StudentID.String] = append(importedStudentCourseRows, utils.ImportedStudentCourseRow{
				Row:                      rowIndex,
				StudentPackage:           studentPackage,
				StudentPackageAccessPath: studentPackageAccessPath,
				StudentPackageEvent:      eventMessage,
			})
		} else {
			mapStudentCourseRowsWithStudentID[studentPackage.StudentID.String] = []utils.ImportedStudentCourseRow{
				{
					Row:                      rowIndex,
					StudentPackage:           studentPackage,
					StudentPackageAccessPath: studentPackageAccessPath,
					StudentPackageEvent:      eventMessage,
				},
			}
		}
	}
	return
}

func toStudentPackageAndStudentPackageAccessPathsEntityFromCsv(line []string, columnNames []string) (
	studentPackage *entities.StudentPackages,
	studentPackageAccessPath *entities.StudentPackageAccessPath,
	event *npb.EventStudentPackage,
	courseID string,
	err error,
) {
	const (
		StudentID = iota
		CourseID
		LocationID
		StartDate
		EndDate
	)

	mandatory := []int{StudentID, CourseID, LocationID, StartDate, EndDate}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		err = fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
		return
	}
	studentPackageID := idutil.ULIDNow()
	studentPackage = &entities.StudentPackages{}
	studentPackageAccessPath = &entities.StudentPackageAccessPath{}

	err = multierr.Combine(
		utils.StringToFormatString(columnNames[StudentID], line[StudentID], false, studentPackage.StudentID.Set),
		utils.StringToFormatString(columnNames[StudentID], line[StudentID], false, studentPackageAccessPath.StudentID.Set),
		utils.StringToFormatString(columnNames[CourseID], line[CourseID], false, studentPackageAccessPath.CourseID.Set),
		utils.StringToFormatString(columnNames[LocationID], line[LocationID], false, studentPackageAccessPath.LocationID.Set),
		studentPackageAccessPath.StudentPackageID.Set(studentPackageID),
		utils.StringToDate(columnNames[StartDate], line[StartDate], false, studentPackage.StartAt.Set),
		utils.StringToDate(columnNames[EndDate], line[EndDate], false, studentPackage.EndAt.Set),
		studentPackage.LocationIDs.Set([]string{line[LocationID]}),
		studentPackage.IsActive.Set(true),
		studentPackage.Properties.Set(&entities.PackageProperties{
			CanWatchVideo:     []string{line[CourseID]},
			CanDoQuiz:         []string{line[CourseID]},
			CanViewStudyGuide: []string{line[CourseID]},
		}),
		studentPackage.ID.Set(studentPackageID),
	)
	if err != nil {
		return
	}
	courseID = line[CourseID]
	event = &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: studentPackage.StudentID.String,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        []string{courseID},
				StartDate:        timestamppb.New(studentPackage.StartAt.Time),
				EndDate:          timestamppb.New(studentPackage.EndAt.Time),
				LocationIds:      []string{studentPackageAccessPath.LocationID.String},
				StudentPackageId: studentPackage.ID.String,
			},
			IsActive: true,
		},
		LocationIds: []string{studentPackageAccessPath.LocationID.String},
	}
	return
}

func (s *CourseMgMt) validateStudentIDAndUpsertStudentPackage(
	ctx context.Context,
	db database.QueryExecer,
	mapLocationAccessWithStudentID map[string]interface{},
	mapLocationAccessWithCourseID map[string]interface{},
	mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
	mapStudentCourseRowsWithStudentID map[string][]utils.ImportedStudentCourseRow,
) (
	events []*npb.EventStudentPackage,
	errors []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError,
) {
	for studentID, rows := range mapStudentCourseRowsWithStudentID {
		var (
			tmpEvents []*npb.EventStudentPackage
			tmpErrors []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError
		)
		tmpEvents, tmpErrors = s.StudentPackage.UpsertStudentPackage(
			ctx,
			db,
			studentID,
			mapLocationAccessWithStudentID,
			mapLocationAccessWithCourseID,
			mapStudentCourseWithStudentPackageAccessPath,
			rows,
		)
		events = append(events, tmpEvents...)
		errors = append(errors, tmpErrors...)
	}
	return
}

func checkMandatoryColumnAndGetIndex(column []string, positions []int) (bool, int) {
	for _, position := range positions {
		if strings.TrimSpace(column[position]) == "" {
			return false, position
		}
	}
	return true, 0
}
