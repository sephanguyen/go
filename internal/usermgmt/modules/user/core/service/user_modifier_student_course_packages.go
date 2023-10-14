package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserModifierService) UpsertStudentCoursePackage(ctx context.Context, req *pb.UpsertStudentCoursePackageRequest) (*pb.UpsertStudentCoursePackageResponse, error) {
	if err := s.validRequest(ctx, req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var studentPackageProfilePBs []*pb.UpsertStudentCoursePackageResponse_StudentPackageProfile

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		student, err := s.StudentRepo.Find(ctx, s.DB, pgtype.Text{String: req.StudentId, Status: pgtype.Present})
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("StudentRepo.Find, studentID %s: %v", req.StudentId, err))
		}

		if student == nil {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot find student with id: %s", req.StudentId))
		}

		for _, studentPackageProfile := range req.StudentPackageProfiles {
			studentPackageID := ""
			startAt := studentPackageProfile.StartTime
			endAt := studentPackageProfile.EndTime

			switch studentPackageProfile.Id.(type) {
			case *pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId:
				studentPackageID, err = s.addStudentCoursePackage(ctx, req.StudentId, studentPackageProfile)
				if err != nil {
					return err
				}

			case *pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId:
				studentPackageID, err = s.editStudentCoursePackage(ctx, studentPackageProfile)
				if err != nil {
					return err
				}
			}

			studentPackage := &pb.UpsertStudentCoursePackageResponse_StudentPackageProfile{
				StudentCoursePackageId: studentPackageID,
				CourseId:               studentPackageProfile.GetCourseId(),
				StartTime:              startAt,
				EndTime:                endAt,
				LocationIds:            studentPackageProfile.LocationIds,
				StudentPackageExtra:    studentPackageProfile.GetStudentPackageExtra(),
			}
			studentPackageProfilePBs = append(studentPackageProfilePBs, studentPackage)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &pb.UpsertStudentCoursePackageResponse{
		StudentId:              req.StudentId,
		StudentPackageProfiles: studentPackageProfilePBs,
	}, nil
}

func (s *UserModifierService) validRequest(ctx context.Context, req *pb.UpsertStudentCoursePackageRequest) error {
	if req.StudentId == "" {
		return fmt.Errorf("UpsertStudentCoursePackage.validRequest: studentID cannot be empty")
	}

	for _, profile := range req.StudentPackageProfiles {
		switch {
		case profile.Id == nil:
			return fmt.Errorf("UpsertStudentCoursePackage.validRequest: package profile id cannot be empty")
		case profile.StartTime.AsTime().After(profile.EndTime.AsTime()):
			return fmt.Errorf("UpsertStudentCoursePackage.validRequest: package profile start date must before end date")
		}

		// if len(profile.LocationIds) == 0 {
		//	return fmt.Errorf("UpsertStudentCoursePackage.validRequest: locationIDs cannot be empty")
		//}

		_, err := s.GetLocations(ctx, profile.LocationIds)
		if err != nil {
			return fmt.Errorf("s.getLocations: %w", err)
		}
	}

	return nil
}

func (s *UserModifierService) addStudentCoursePackage(ctx context.Context, studentID string, studentPackageProfile *pb.UpsertStudentCoursePackageRequest_StudentPackageProfile) (string, error) {
	startAt := studentPackageProfile.StartTime
	endAt := studentPackageProfile.EndTime
	var addStudentPackageCourseReq *fpb.AddStudentPackageCourseRequest
	if studentPackageProfile.StudentPackageExtra != nil {
		addStudentPackageCourseReq = &fpb.AddStudentPackageCourseRequest{
			CourseIds:           []string{studentPackageProfile.GetCourseId()},
			StudentId:           studentID,
			StartAt:             startAt,
			EndAt:               endAt,
			LocationIds:         golibs.Uniq(getLocationIDsFromStudentPackageExtras(studentPackageProfile.StudentPackageExtra)),
			StudentPackageExtra: mapFromStudentPackageProfileToAddStudentPackageCourseExtra(studentPackageProfile.GetStudentPackageExtra(), studentPackageProfile.GetCourseId()),
		}
	} else {
		addStudentPackageCourseReq = &fpb.AddStudentPackageCourseRequest{
			StudentId:   studentID,
			CourseIds:   []string{studentPackageProfile.GetCourseId()},
			StartAt:     startAt,
			EndAt:       endAt,
			LocationIds: studentPackageProfile.GetLocationIds(),
		}
	}
	resp, err := s.FatimaClient.AddStudentPackageCourse(signCtx(ctx), addStudentPackageCourseReq)
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Errorf("s.FatimaClient.AddStudentPackageCourse: %w", err).Error())
	}

	return resp.StudentPackageId, nil
}

func mapFromStudentPackageProfileToAddStudentPackageCourseExtra(studentPackagesExtra []*pb.StudentPackageExtra, courseId string) []*fpb.AddStudentPackageCourseRequest_AddStudentPackageExtra {
	result := make([]*fpb.AddStudentPackageCourseRequest_AddStudentPackageExtra, 0)
	for _, value := range studentPackagesExtra {
		result = append(result, &fpb.AddStudentPackageCourseRequest_AddStudentPackageExtra{
			LocationId: value.LocationId,
			ClassId:    value.ClassId,
			CourseId:   courseId,
		})
	}
	return result
}

func (s *UserModifierService) editStudentCoursePackage(ctx context.Context, studentPackageProfile *pb.UpsertStudentCoursePackageRequest_StudentPackageProfile) (string, error) {
	startAt := studentPackageProfile.StartTime
	endAt := studentPackageProfile.EndTime

	var editStudentPackageCourseReq *fpb.EditTimeStudentPackageRequest
	if studentPackageProfile.StudentPackageExtra != nil {
		editStudentPackageCourseReq = &fpb.EditTimeStudentPackageRequest{
			StudentPackageId:    studentPackageProfile.GetStudentPackageId(),
			StartAt:             startAt,
			EndAt:               endAt,
			StudentPackageExtra: mapFromStudentPackageProfileToEditTimeStudentPackageCourseExtra(studentPackageProfile.GetStudentPackageExtra()),
			LocationIds:         golibs.Uniq(getLocationIDsFromStudentPackageExtras(studentPackageProfile.StudentPackageExtra)),
		}
	} else {
		editStudentPackageCourseReq = &fpb.EditTimeStudentPackageRequest{
			StudentPackageId: studentPackageProfile.GetStudentPackageId(),
			StartAt:          startAt,
			EndAt:            endAt,
			LocationIds:      studentPackageProfile.GetLocationIds(),
		}
	}

	resp, err := s.FatimaClient.EditTimeStudentPackage(signCtx(ctx), editStudentPackageCourseReq)
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Errorf("s.FatimaClient.EditTimeStudentPackage: %w", err).Error())
	}

	return resp.StudentPackageId, nil
}

func getLocationIDsFromStudentPackageExtras(studentPackageExtras []*pb.StudentPackageExtra) []string {
	locationIDs := make([]string, 0)
	for _, studentPackageExtra := range studentPackageExtras {
		locationIDs = append(locationIDs, studentPackageExtra.LocationId)
	}
	return locationIDs
}

func mapFromStudentPackageProfileToEditTimeStudentPackageCourseExtra(studentPackagesExtra []*pb.StudentPackageExtra) []*fpb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra {
	result := make([]*fpb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra, 0)
	for _, value := range studentPackagesExtra {
		result = append(result, &fpb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra{
			LocationId: value.LocationId,
			ClassId:    value.ClassId,
		})
	}
	return result
}
