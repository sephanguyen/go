package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StudentPackageService) DeleteStudentClass(
	ctx context.Context,
	db database.QueryExecer,
	mapStudentCourse map[string]entities.StudentPackageAccessPath,
	mapClass map[string]entities.Class,
	importedStudentClass []utils.ImportedStudentClassRow,
) (
	events []*npb.EventStudentPackageV2,
	errors []*pb.ImportStudentClassesResponse_ImportStudentClassesError,
) {
	errors = make([]*pb.ImportStudentClassesResponse_ImportStudentClassesError, 0, len(importedStudentClass))
	events = make([]*npb.EventStudentPackageV2, 0, len(importedStudentClass))
	for _, class := range importedStudentClass {
		packageAccessPath, err := validateStudentCourseLocationWithClassLocation(class, mapStudentCourse, mapClass)
		if err != nil {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: class.Row,
				Error:     err.Error(),
			})
			continue
		}
		studentPackages, err := s.StudentPackageRepo.GetByID(ctx, db, packageAccessPath.StudentPackageID.String)
		if err != nil {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: class.Row,
				Error:     fmt.Sprintf("can't get student package by id %v with err: %v", packageAccessPath.StudentPackageID.String, err.Error()),
			})
			continue
		}
		studentPackageClass := entities.StudentPackageClass{}
		_ = multierr.Combine(
			studentPackageClass.StudentPackageID.Set(studentPackages.ID.String),
			studentPackageClass.StudentID.Set(studentPackages.StudentID.String),
			studentPackageClass.ClassID.Set(class.ClassID),
			studentPackageClass.CourseID.Set(class.CourseID),
			studentPackageClass.LocationID.Set(packageAccessPath.LocationID.String),
			studentPackageClass.UpdatedAt.Set(time.Now()),
			studentPackageClass.DeletedAt.Set(time.Now()),
		)
		err = s.StudentPackageClassRepo.Delete(ctx, db, &studentPackageClass)
		if err != nil {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: class.Row,
				Error:     fmt.Sprintf("can't insert student package class with err: %v", err.Error()),
			})
			continue
		}
		event := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: class.StudentID,
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   class.CourseID,
					LocationId: packageAccessPath.LocationID.String,
					StartDate:  timestamppb.New(studentPackages.StartAt.Time),
					EndDate:    timestamppb.New(studentPackages.EndAt.Time),
				},
				IsActive: true,
			},
		}
		events = append(events, event)
	}
	return
}

func (s *StudentPackageService) UpsertStudentClass(
	ctx context.Context,
	db database.QueryExecer,
	mapStudentCourse map[string]entities.StudentPackageAccessPath,
	mapClass map[string]entities.Class,
	importedStudentClass []utils.ImportedStudentClassRow,
) (
	events []*npb.EventStudentPackageV2,
	errors []*pb.ImportStudentClassesResponse_ImportStudentClassesError,
) {
	errors = make([]*pb.ImportStudentClassesResponse_ImportStudentClassesError, 0, len(importedStudentClass))
	events = make([]*npb.EventStudentPackageV2, 0, len(importedStudentClass))
	for _, class := range importedStudentClass {
		packageAccessPath, err := validateStudentCourseLocationWithClassLocation(class, mapStudentCourse, mapClass)
		if err != nil {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: class.Row,
				Error:     err.Error(),
			})
			continue
		}
		studentPackages, err := s.StudentPackageRepo.GetByID(ctx, db, packageAccessPath.StudentPackageID.String)
		if err != nil {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: class.Row,
				Error:     fmt.Sprintf("can't get student package by id %v with err: %v", packageAccessPath.StudentPackageID.String, err.Error()),
			})
			continue
		}
		studentPackageClass := entities.StudentPackageClass{}
		_ = multierr.Combine(
			studentPackageClass.StudentPackageID.Set(studentPackages.ID.String),
			studentPackageClass.StudentID.Set(studentPackages.StudentID.String),
			studentPackageClass.ClassID.Set(class.ClassID),
			studentPackageClass.CourseID.Set(class.CourseID),
			studentPackageClass.LocationID.Set(packageAccessPath.LocationID.String),
			studentPackageClass.CreatedAt.Set(time.Now()),
			studentPackageClass.UpdatedAt.Set(time.Now()),
			studentPackageClass.DeletedAt.Set(nil),
		)
		err = s.StudentPackageClassRepo.Upsert(ctx, db, &studentPackageClass)
		if err != nil {
			errors = append(errors, &pb.ImportStudentClassesResponse_ImportStudentClassesError{
				RowNumber: class.Row,
				Error:     fmt.Sprintf("can't insert student package class with err: %v", err.Error()),
			})
			continue
		}
		event := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: class.StudentID,
				Package: &npb.EventStudentPackageV2_PackageV2{
					ClassId:    class.ClassID,
					CourseId:   class.CourseID,
					LocationId: packageAccessPath.LocationID.String,
					StartDate:  timestamppb.New(studentPackages.StartAt.Time),
					EndDate:    timestamppb.New(studentPackages.EndAt.Time),
				},
				IsActive: true,
			},
		}
		events = append(events, event)
	}
	return
}

func validateStudentCourseLocationWithClassLocation(
	studentClassRow utils.ImportedStudentClassRow,
	mapStudentCourse map[string]entities.StudentPackageAccessPath,
	mapClass map[string]entities.Class,
) (packageAccessPath entities.StudentPackageAccessPath, err error) {
	keyStudentCourse := fmt.Sprintf("%v_%v", studentClassRow.StudentID, studentClassRow.CourseID)
	packageAccessPath, ok := mapStudentCourse[keyStudentCourse]
	if !ok {
		err = fmt.Errorf("this course %v didn't register for student %v so we can't register for this class", studentClassRow.CourseID, studentClassRow.StudentID)
		return
	}

	class, ok := mapClass[studentClassRow.ClassID]
	if !ok {
		err = fmt.Errorf("this class %v didn't exist in database", studentClassRow.ClassID)
		return
	}

	if packageAccessPath.LocationID.String != class.LocationID.String {
		err = fmt.Errorf("this class %v difference location with course so we can't register for this class", studentClassRow.ClassID)
		return
	}

	return
}

func validateAccessLocationForStudentID(locationID string, studentID string, mapAccessLocationForStudent map[string]interface{}) (err error) {
	key := fmt.Sprintf("%v_%v", locationID, studentID)
	if _, ok := mapAccessLocationForStudent[key]; !ok {
		err = fmt.Errorf("student id %v can't access location id %v", studentID, locationID)
	}
	return
}

func validateAccessLocationForCourseID(locationID string, courseID string, mapAccessLocationForCourse map[string]interface{}) (err error) {
	key := fmt.Sprintf("%v_%v", locationID, courseID)
	if _, ok := mapAccessLocationForCourse[key]; !ok {
		err = fmt.Errorf("course id %v can't access location id %v", courseID, locationID)
	}
	return
}

func (s *StudentPackageService) UpsertStudentPackage(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	mapLocationAccessWithStudentID map[string]interface{},
	mapLocationAccessWithCourseID map[string]interface{},
	mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
	importedStudentCourseRows []utils.ImportedStudentCourseRow,
) (
	events []*npb.EventStudentPackage,
	errors []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError,
) {
	for _, row := range importedStudentCourseRows {
		tmpStudentPackageAccessPath := row.StudentPackageAccessPath

		err := utils.GroupErrorFunc(
			validateAccessLocationForStudentID(tmpStudentPackageAccessPath.LocationID.String, tmpStudentPackageAccessPath.StudentID.String, mapLocationAccessWithStudentID),
			validateAccessLocationForCourseID(tmpStudentPackageAccessPath.LocationID.String, tmpStudentPackageAccessPath.CourseID.String, mapLocationAccessWithCourseID),
		)
		if err != nil {
			errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
				RowNumber: row.Row,
				Error:     err.Error(),
			})
			continue
		}
		key := fmt.Sprintf("%v_%v", studentID, tmpStudentPackageAccessPath.CourseID.String)
		if packageAccessPath, ok := mapStudentCourseWithStudentPackageAccessPath[key]; ok {
			tmpStudentPack := row.StudentPackage
			_ = tmpStudentPack.ID.Set(packageAccessPath.StudentPackageID.String)
			_ = tmpStudentPack.DeletedAt.Set(nil)
			err = s.StudentPackageRepo.Update(ctx, db, tmpStudentPack)
			if err != nil {
				errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					RowNumber: row.Row,
					Error:     fmt.Sprintf("update student package by student id %s and course id %s have error %s", studentID, tmpStudentPackageAccessPath.CourseID.String, err.Error()),
				})
				continue
			}
			tmpEvent := row.StudentPackageEvent
			tmpEvent.StudentPackage.Package.StudentPackageId = packageAccessPath.StudentPackageID.String
			_ = tmpStudentPackageAccessPath.StudentPackageID.Set(packageAccessPath.StudentPackageID.String)
			err = s.StudentPackageAccessPathRepo.Update(ctx, db, tmpStudentPackageAccessPath)
			if err != nil {
				errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					RowNumber: row.Row,
					Error:     fmt.Sprintf("update student package access path by student id %s and course id %s have error %s", studentID, tmpStudentPackageAccessPath.CourseID.String, err.Error()),
				})
				continue
			}
			events = append(events, tmpEvent)
		} else {
			err = s.StudentPackageRepo.Insert(ctx, db, row.StudentPackage)
			if err != nil {
				errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					RowNumber: row.Row,
					Error:     fmt.Sprintf("insert student package by student id %s and course id %s have error %s", studentID, tmpStudentPackageAccessPath.CourseID.String, err.Error()),
				})
				continue
			}
			err = s.StudentPackageAccessPathRepo.Insert(ctx, db, tmpStudentPackageAccessPath)
			if err != nil {
				errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					RowNumber: row.Row,
					Error:     fmt.Sprintf("insert student package access path by student id %s and course id %s have error %s", studentID, tmpStudentPackageAccessPath.CourseID.String, err.Error()),
				})
				continue
			}
			err = s.StudentCourseRepo.UpsertStudentCourse(ctx, db, entities.StudentCourse{
				StudentPackageID:  row.StudentPackage.ID,
				StudentID:         row.StudentPackage.StudentID,
				CourseID:          row.StudentPackageAccessPath.CourseID,
				LocationID:        row.StudentPackageAccessPath.LocationID,
				StudentStartDate:  row.StudentPackage.StartAt,
				StudentEndDate:    row.StudentPackage.EndAt,
				CourseSlot:        pgtype.Int4{Status: pgtype.Null},
				CourseSlotPerWeek: pgtype.Int4{Status: pgtype.Null},
				Weight:            pgtype.Int4{Status: pgtype.Null},
				PackageType:       pgtype.Text{Status: pgtype.Null},
				DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			})
			if err != nil {
				errors = append(errors, &pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					RowNumber: row.Row,
					Error:     fmt.Sprintf("insert student course by student id %s and course id %s have error %s", studentID, tmpStudentPackageAccessPath.CourseID.String, err.Error()),
				})
				continue
			}
			events = append(events, row.StudentPackageEvent)
		}
	}
	return
}

func (s *StudentPackageService) GetMapStudentCourseWithStudentPackageIDByIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (mapStudentCourse map[string]entities.StudentPackageAccessPath, err error) {
	mapStudentCourse, err = s.StudentPackageAccessPathRepo.GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(ctx, db, studentIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get map student course with student package id by student ids with error: %v", err.Error())
	}
	return
}
