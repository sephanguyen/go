package service

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Group func for manual upsert student course

func (s *StudentPackageService) UpsertStudentPackageForManualFlow(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	studentCourse *pb.StudentCourseData,
) (event *npb.EventStudentPackage, err error) {
	err = s.StudentPackageAccessPathRepo.CheckExistStudentPackageAccessPath(ctx, db, studentID, studentCourse.CourseId)
	if err != nil {
		return
	}
	courseIDs := []string{studentCourse.CourseId}
	property := &entities.PackageProperties{
		CanDoQuiz:         courseIDs,
		CanViewStudyGuide: courseIDs,
		CanWatchVideo:     courseIDs,
	}
	studentPackage := &entities.StudentPackages{
		ID:        pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
		StudentID: pgtype.Text{String: studentID, Status: pgtype.Present},
		StartAt:   pgtype.Timestamptz{Status: pgtype.Present, Time: studentCourse.StartDate.AsTime()},
		EndAt:     pgtype.Timestamptz{Status: pgtype.Present, Time: studentCourse.EndDate.AsTime()},
		IsActive:  pgtype.Bool{Bool: true, Status: pgtype.Present},
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}
	_ = studentPackage.LocationIDs.Set([]string{studentCourse.LocationId})
	_ = studentPackage.Properties.Set(property)
	_ = studentPackage.CreatedAt.Set(nil)
	_ = studentPackage.UpdatedAt.Set(nil)
	_ = studentPackage.PackageID.Set(nil)

	err = s.StudentPackageRepo.Insert(ctx, db, studentPackage)
	if err != nil {
		err = status.Errorf(codes.Internal, "upsert student package have error %v", err.Error())
		return
	}
	action := pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_INSERT.String()
	flow := "Manual Flow"
	studentPackageAccessPath := entities.StudentPackageAccessPath{
		StudentPackageID: studentPackage.ID,
		StudentID:        studentPackage.StudentID,
		CourseID:         pgtype.Text{Status: pgtype.Present, String: studentCourse.CourseId},
		LocationID:       pgtype.Text{Status: pgtype.Present, String: studentCourse.LocationId},
		AccessPath:       pgtype.Text{Status: pgtype.Null},
		DeletedAt:        pgtype.Timestamptz{Status: pgtype.Null},
		CreatedAt:        pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now()},
		UpdatedAt:        pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now()},
	}
	err = s.StudentPackageAccessPathRepo.Insert(ctx, db, &studentPackageAccessPath)
	if err != nil {
		err = status.Errorf(codes.Internal, "upsert student package access path have error %v", err.Error())
		return
	}
	if err = s.writeStudentPackageLog(ctx, db, studentPackage, studentCourse.CourseId, action, flow); err != nil {
		return
	}
	event = &npb.EventStudentPackage{
		LocationIds: []string{
			studentCourse.LocationId,
		},
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: studentID,
			IsActive:  true,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        courseIDs,
				StartDate:        timestamppb.New(studentCourse.StartDate.AsTime()),
				EndDate:          timestamppb.New(studentCourse.EndDate.AsTime()),
				LocationIds:      []string{studentCourse.LocationId},
				StudentPackageId: studentPackage.ID.String,
			},
		},
	}
	return
}

func (s *StudentPackageService) UpdateTimeStudentPackageForManualFlow(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	studentCourse *pb.StudentCourseData,
) (event *npb.EventStudentPackage, err error) {
	courseIDs := []string{studentCourse.CourseId}
	property := &entities.PackageProperties{
		CanDoQuiz:         courseIDs,
		CanViewStudyGuide: courseIDs,
		CanWatchVideo:     courseIDs,
	}
	studentPackage := entities.StudentPackages{
		ID:        pgtype.Text{String: studentCourse.StudentPackageId.Value, Status: pgtype.Present},
		StudentID: pgtype.Text{String: studentID, Status: pgtype.Present},
		PackageID: pgtype.Text{Status: pgtype.Null},
		StartAt:   pgtype.Timestamptz{Time: studentCourse.StartDate.AsTime(), Status: pgtype.Present},
		EndAt:     pgtype.Timestamptz{Time: studentCourse.EndDate.AsTime(), Status: pgtype.Present},
		IsActive:  pgtype.Bool{Bool: true, Status: pgtype.Present},
		CreatedAt: pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}
	_ = studentPackage.LocationIDs.Set([]string{studentCourse.LocationId})
	_ = studentPackage.Properties.Set(property)

	err = s.StudentPackageRepo.Update(ctx, db, &studentPackage)
	if err != nil {
		err = status.Errorf(codes.Internal, "upsert student package have error %v", err.Error())
		return
	}
	action := pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_UPDATE.String()
	flow := "Manual flow"
	if err = s.writeStudentPackageLog(ctx, db, &studentPackage, studentCourse.CourseId, action, flow); err != nil {
		return
	}
	event = &npb.EventStudentPackage{
		LocationIds: []string{
			studentCourse.LocationId,
		},
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: studentID,
			IsActive:  true,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        courseIDs,
				StartDate:        timestamppb.New(studentCourse.StartDate.AsTime()),
				EndDate:          timestamppb.New(studentCourse.EndDate.AsTime()),
				LocationIds:      []string{studentCourse.LocationId},
				StudentPackageId: studentPackage.ID.String,
			},
		},
	}
	return
}
