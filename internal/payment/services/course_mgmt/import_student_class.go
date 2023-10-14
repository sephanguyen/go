package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CourseMgMt) ImportStudentClasses(ctx context.Context, req *pb.ImportStudentClassesRequest) (res *pb.ImportStudentClassesResponse, err error) {
	var (
		errors                                []*pb.ImportStudentClassesResponse_ImportStudentClassesError
		studentClassRows                      []utils.ImportedStudentClassRow
		classIDs                              []string
		studentIDs                            []string
		mapStudentCourseWithPackageAccessPath map[string]entities.StudentPackageAccessPath
		mapClassWithClassID                   map[string]entities.Class
		eventStudentClasses                   []*npb.EventStudentPackageV2
	)
	res = &pb.ImportStudentClassesResponse{}
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
		"class_id",
	}

	err = utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}
	studentClassRows, studentIDs, classIDs, errors = toMapStudentWithStudentClass(lines)
	if len(errors) > 0 {
		res.Errors = errors
		return
	}

	mapStudentCourseWithPackageAccessPath, err = s.StudentPackage.GetMapStudentCourseWithStudentPackageIDByIDs(ctx, s.DB, studentIDs)
	if err != nil {
		return nil, err
	}

	mapClassWithClassID, err = s.ClassService.GetMapClassWithLocationByClassIDs(ctx, s.DB, classIDs)
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		if req.IsAddClass {
			eventStudentClasses, errors = s.StudentPackage.UpsertStudentClass(ctx, tx, mapStudentCourseWithPackageAccessPath, mapClassWithClassID, studentClassRows)
		} else {
			eventStudentClasses, errors = s.StudentPackage.DeleteStudentClass(ctx, tx, mapStudentCourseWithPackageAccessPath, mapClassWithClassID, studentClassRows)
		}
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

	err = s.SubscriptionService.PublishStudentClass(ctx, eventStudentClasses)
	if err != nil {
		log.Printf("Error when publish student package: %s", err.Error())
	}

	return res, nil
}

func toMapStudentWithStudentClass(rows [][]string) (
	studentClassRows []utils.ImportedStudentClassRow,
	studentIDs []string,
	classIDs []string,
	errors []*pb.ImportStudentClassesResponse_ImportStudentClassesError,
) {
	var mapValue interface{}
	mapStudentCourseUnique := make(map[string]interface{})
	mapStudentClassUnique := make(map[string]interface{})
	mapStudentUnique := make(map[string]interface{})
	mapClassUnique := make(map[string]interface{})
	headerTitles := []string{
		"student_id",
		"course_id",
		"class_id",
	}
	for i, line := range rows[1:] {
		rowIndex := int32(i) + 2
		var (
			studentClassRow utils.ImportedStudentClassRow
			studentID       string
			err             error
		)
		studentClassRow, studentID, err = toStudentClassPackageFromCsv(line, headerTitles)
		if err != nil {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: rowIndex, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf(constant.UnableToParseStudentCourse, err),
			})
			continue
		}
		studentClassRow.Row = rowIndex

		studentClassRows = append(studentClassRows, studentClassRow)
		keyStudentCourse := fmt.Sprintf("%s_%s", studentID, studentClassRow.CourseID)
		if _, ok := mapStudentCourseUnique[keyStudentCourse]; ok {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: rowIndex, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("duplicate student class with student id %s and course id %s", studentID, studentClassRow.CourseID),
			})
			continue
		}
		mapStudentCourseUnique[keyStudentCourse] = mapValue

		keyStudentClass := fmt.Sprintf("%s_%s", studentID, studentClassRow.ClassID)
		if _, ok := mapStudentClassUnique[keyStudentClass]; ok {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: rowIndex, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("duplicate student class with student id %s and class id %s", studentID, studentClassRow.ClassID),
			})
			continue
		}
		mapStudentClassUnique[keyStudentClass] = mapValue

		if _, ok := mapStudentUnique[studentID]; !ok {
			studentIDs = append(studentIDs, studentID)
			mapStudentUnique[studentID] = mapValue
		}
		if _, ok := mapClassUnique[studentClassRow.ClassID]; !ok {
			classIDs = append(classIDs, studentClassRow.ClassID)
			mapClassUnique[studentClassRow.ClassID] = mapValue
		}
	}
	return
}

func toStudentClassPackageFromCsv(line []string, columnNames []string) (
	studentClassRow utils.ImportedStudentClassRow,
	studentID string,
	err error,
) {
	const (
		StudentID = iota
		CourseID
		ClassID
	)

	mandatory := []int{StudentID, CourseID, ClassID}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		err = fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
		return
	}
	studentClassRow.ClassID = line[ClassID]
	studentClassRow.CourseID = line[CourseID]
	studentClassRow.StudentID = line[StudentID]
	studentID = line[StudentID]
	return
}
