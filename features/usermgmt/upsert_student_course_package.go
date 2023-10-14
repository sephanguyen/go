package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	magicUID1 = newID()
	magicUID2 = newID()

	retryTimes = 10
)

func (s *suite) studentExistInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.ExistingStudents = []*entity.LegacyStudent{student}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackages(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.addCourseDataToReq(ctx)
	s.addUpdateCourseDataToReq(ctx)
	req := stepState.Request.(*pb.UpsertStudentCoursePackageRequest)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpsertStudentPackage(contextWithToken(ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackagesWithStudentPackageExtra(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := s.addCourseDataWithPackageExtraToRequest()
	stepState.Request = req
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpsertStudentPackage(contextWithToken(ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackagesWithInvalid(ctx context.Context, account, studentID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.addCourseDataToReq(ctx)
	req := stepState.Request.(*pb.UpsertStudentCoursePackageRequest)
	switch studentID {
	case "empty":
		req.StudentId = ""
	case "non-exist":
		req.StudentId = fmt.Sprintf("edited-%s", req.StudentId)
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpsertStudentPackage(contextWithToken(ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackagesWithOnly(ctx context.Context, account, data string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch data {
	case "new course":
		s.addCourseDataToReq(ctx)
	case "edit existed course":
		s.addUpdateCourseDataToReq(ctx)
	}

	req := stepState.Request.(*pb.UpsertStudentCoursePackageRequest)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpsertStudentPackage(contextWithToken(ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) cannotUpsertStudentCoursePackages(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return nil, errors.New("expecting err but got nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addCourseDataToReq(ctx context.Context) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	profiles := []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		{
			Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: magicUID1,
			},
			StartTime:   timestamppb.New(now.Add(-24 * time.Hour)),
			EndTime:     timestamppb.New(now.Add(-24 * time.Hour).Add(time.Hour)),
			LocationIds: []string{constants.ManabieOrgLocation},
		},
		{
			Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: magicUID2,
			},
			StartTime:   timestamppb.New(now),
			EndTime:     timestamppb.New(now.Add(24 * time.Hour)),
			LocationIds: []string{constants.ManabieOrgLocation},
		},
	}

	stepState.Request = &pb.UpsertStudentCoursePackageRequest{
		StudentId:              s.ExistingStudents[0].ID.String,
		StudentPackageProfiles: profiles,
	}
}

func (s *suite) addCourseDataWithPackageExtraToRequest() *pb.UpsertStudentCoursePackageRequest {
	now := time.Now()
	profiles := []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		{
			Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: magicUID1,
			},
			StartTime: timestamppb.New(now.Add(-24 * time.Hour)),
			EndTime:   timestamppb.New(now.Add(-24 * time.Hour).Add(time.Hour)),
			StudentPackageExtra: []*pb.StudentPackageExtra{
				{
					LocationId: constants.ManabieOrgLocation,
					ClassId:    "existing-class-id-1",
				},
			},
		},
	}

	return &pb.UpsertStudentCoursePackageRequest{
		StudentId:              s.ExistingStudents[0].ID.String,
		StudentPackageProfiles: profiles,
	}
}

func (s *suite) addUpdateCourseDataToReq(ctx context.Context) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	stepState.Request = &pb.UpsertStudentCoursePackageRequest{
		StudentId: s.ExistingStudents[0].ID.String,
	}

	for packageID := range s.MapExistingPackageAndCourses {
		s.addUpdateStudentPackageToUpdateStudentRequest(ctx, packageID, now.Add(-60*24*time.Hour), now.Add(-30*24*time.Hour))
	}
}

func (s *suite) upsertStudentCoursePackagesSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if err := s.validateStudentPackageResponse(ctx); err != nil {
		return ctx, fmt.Errorf("validateStudentPackageResponse: %w", err)
	}
	if err := s.validateStudentPackageStored(ctx); err != nil {
		return ctx, fmt.Errorf("validateStudentPackageStored: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackagesSuccessfullyWithStudentPackageExtra(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if err := s.validateStudentPackageResponse(ctx); err != nil {
		return ctx, fmt.Errorf("validateStudentPackageResponse: %w", err)
	}
	if err := s.validateStudentPackageStoredWithStudentPackageExtra(ctx); err != nil {
		return ctx, fmt.Errorf("validateStudentPackageStored: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackagesWithCourseInvalidStartDateAndEndDate(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	profiles := []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		{
			Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: magicUID1,
			},
			StartTime: timestamppb.New(now),
			EndTime:   timestamppb.New(now.Add(-time.Hour)),
		},
	}

	req := &pb.UpsertStudentCoursePackageRequest{
		StudentId:              s.ExistingStudents[0].ID.String,
		StudentPackageProfiles: profiles,
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpsertStudentPackage(contextWithToken(ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackagesWithLocationIdsEmpty(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	profiles := []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		{
			Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: magicUID1,
			},
			StartTime:   timestamppb.New(now),
			EndTime:     timestamppb.New(now.Add(time.Hour)),
			LocationIds: nil,
		},
	}

	req := &pb.UpsertStudentCoursePackageRequest{
		StudentId:              s.ExistingStudents[0].ID.String,
		StudentPackageProfiles: profiles,
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpsertStudentPackage(contextWithToken(ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentCoursePackagesWithPackageIdEmpty(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	profiles := []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		{
			StartTime: timestamppb.New(now),
			EndTime:   timestamppb.New(now.Add(time.Hour)),
		},
	}

	req := &pb.UpsertStudentCoursePackageRequest{
		StudentId:              s.ExistingStudents[0].ID.String,
		StudentPackageProfiles: profiles,
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpsertStudentPackage(contextWithToken(ctx), s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) assignCoursePackageToExistStudent(ctx context.Context) (context.Context, error) {
	s.MapExistingPackageAndCourses = make(map[string]string)
	coursePackages := []*fpb.AddStudentPackageCourseRequest{
		{
			StudentId:   s.ExistingStudents[0].ID.String,
			CourseIds:   []string{"existing-course-id-1"},
			StartAt:     timestamppb.New(time.Now()),
			EndAt:       timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
			LocationIds: []string{constants.ManabieOrgLocation},
		},
		{
			StudentId:   s.ExistingStudents[0].ID.String,
			CourseIds:   []string{"existing-course-id-2"},
			StartAt:     timestamppb.New(time.Now()),
			EndAt:       timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
			LocationIds: []string{constants.ManabieOrgLocation},
		},
	}

	ctx, err := s.signedAsAccount(ctx, schoolAdminType)
	if err != nil {
		return nil, err
	}

	for _, req := range coursePackages {
		resp, err := tryAddStudentPackage(contextWithToken(ctx), s.FatimaConn, req)
		if err != nil {
			return nil, err
		}

		s.MapExistingPackageAndCourses[resp.StudentPackageId] = req.CourseIds[0]
	}

	return ctx, nil
}

func (s *suite) assignStudentPackageWithClassEmptyToExistStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &fpb.AddStudentPackageCourseRequest{
		StudentId: s.ExistingStudents[0].ID.String,
		CourseIds: []string{"existing-course-id-2"},
		StartAt:   timestamppb.New(time.Now()),
		EndAt:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		StudentPackageExtra: []*fpb.AddStudentPackageCourseRequest_AddStudentPackageExtra{
			{
				CourseId:   "existing-course-id-2",
				LocationId: constants.ManabieOrgLocation,
				ClassId:    "",
			},
		},
	}
	ctx, err := s.signedAsAccount(ctx, schoolAdminType)
	if err != nil {
		return ctx, err
	}
	resp, err := tryAddStudentPackage(contextWithToken(ctx), s.FatimaConn, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentPackageID = resp.StudentPackageId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addUpdateStudentPackageToUpdateStudentRequest(ctx context.Context, studentPackageId string, startTime time.Time, endTime time.Time) context.Context {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpsertStudentCoursePackageRequest)

	packageProfile := &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
			StudentPackageId: studentPackageId,
		},
		StartTime:   timestamppb.New(startTime),
		EndTime:     timestamppb.New(endTime),
		LocationIds: []string{constants.ManabieOrgLocation},
	}
	req.StudentPackageProfiles = append(req.StudentPackageProfiles, packageProfile)

	stepState.Request = req

	return StepStateToContext(ctx, stepState)
}

func (s *suite) validateStudentPackageStored(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.UpsertStudentCoursePackageResponse)
	studentID := resp.StudentId

	studentPackages, err := fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).ListStudentPackage(
		contextWithToken(ctx),
		&fpb.ListStudentPackageRequest{
			StudentIds: []string{studentID},
		})
	if err != nil {
		return err
	}
	if len(resp.StudentPackageProfiles) != len(studentPackages.StudentPackages) {
		return fmt.Errorf("expect student has %d number of course but got %d", len(resp.StudentPackageProfiles), len(studentPackages.StudentPackages))
	}

	studentPackageIdWithPackage := make(map[string]*fpb.StudentPackage)
	for _, studentPackage := range studentPackages.StudentPackages {
		studentPackageIdWithPackage[studentPackage.Id] = studentPackage
	}

	for _, packageResp := range resp.StudentPackageProfiles {
		studentPackage := studentPackageIdWithPackage[packageResp.StudentCoursePackageId]
		switch {
		case !stringutil.SliceEqual(studentPackage.LocationIds, packageResp.LocationIds):
			return fmt.Errorf("expect student package locations: %v but got %v", packageResp.LocationIds, studentPackage.LocationIds)
		case !studentPackage.StartAt.AsTime().Round(time.Second).Equal(packageResp.StartTime.AsTime().Round(time.Second)):
			return fmt.Errorf("expect student package start at %v but got %v", packageResp.StartTime.AsTime().String(), studentPackage.StartAt.AsTime().String())
		case !studentPackage.EndAt.AsTime().Round(time.Second).Equal(packageResp.EndTime.AsTime().Round(time.Second)):
			return fmt.Errorf("expect student package end at %v but got %v", packageResp.EndTime.AsTime().String(), studentPackage.EndAt.AsTime().String())
		// we don't have courseID in request and response incase updating student_package existed
		case packageResp.CourseId != "" && studentPackage.Properties.CanDoQuiz[0] != packageResp.CourseId:
			return fmt.Errorf("expect student package course: %v but got %v", packageResp.CourseId, studentPackage.Properties.CanDoQuiz[0])
		}
	}

	return nil
}

func (s *suite) validateStudentPackageStoredWithStudentPackageExtra(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.UpsertStudentCoursePackageResponse)
	studentID := resp.StudentId

	studentPackages, err := fpb.NewSubscriptionModifierServiceClient(s.FatimaConn).ListStudentPackage(
		contextWithToken(ctx),
		&fpb.ListStudentPackageRequest{
			StudentIds: []string{studentID},
		})
	if err != nil {
		return err
	}
	if len(resp.StudentPackageProfiles) != len(studentPackages.StudentPackages) {
		return fmt.Errorf("expect student has %d number of course but got %d", len(resp.StudentPackageProfiles), len(studentPackages.StudentPackages))
	}

	studentPackageIdWithPackage := make(map[string]*fpb.StudentPackage)
	for _, studentPackage := range studentPackages.StudentPackages {
		studentPackageIdWithPackage[studentPackage.Id] = studentPackage
	}

	for _, packageResp := range resp.StudentPackageProfiles {
		studentPackage := studentPackageIdWithPackage[packageResp.StudentCoursePackageId]
		locationIDs := make([]string, 0)
		for _, packageProfile := range resp.StudentPackageProfiles {
			for _, packageExtra := range packageProfile.StudentPackageExtra {
				locationIDs = append(locationIDs, packageExtra.LocationId)
			}
		}
		switch {
		case !stringutil.SliceEqual(studentPackage.LocationIds, locationIDs):
			return fmt.Errorf("expect student package locations: %v but got %v", locationIDs, studentPackage.LocationIds)
		case !studentPackage.StartAt.AsTime().Round(time.Second).Equal(packageResp.StartTime.AsTime().Round(time.Second)):
			return fmt.Errorf("expect student package start at %v but got %v", packageResp.StartTime.AsTime().String(), studentPackage.StartAt.AsTime().String())
		case !studentPackage.EndAt.AsTime().Round(time.Second).Equal(packageResp.EndTime.AsTime().Round(time.Second)):
			return fmt.Errorf("expect student package end at %v but got %v", packageResp.EndTime.AsTime().String(), studentPackage.EndAt.AsTime().String())
		// we don't have courseID in request and response incase updating student_package existed
		case packageResp.CourseId != "" && studentPackage.Properties.CanDoQuiz[0] != packageResp.CourseId:
			return fmt.Errorf("expect student package course: %v but got %v", packageResp.CourseId, studentPackage.Properties.CanDoQuiz[0])
		}
	}

	return nil
}

func (s *suite) validateStudentPackageResponse(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.UpsertStudentCoursePackageResponse)
	req := stepState.Request.(*pb.UpsertStudentCoursePackageRequest)

	if len(resp.StudentPackageProfiles) != len(req.StudentPackageProfiles) {
		return fmt.Errorf("expect student has %d number of course but got %d", len(resp.StudentPackageProfiles), len(req.StudentPackageProfiles))
	}

	studentPackageIdWithPackage := make(map[string]*pb.UpsertStudentCoursePackageResponse_StudentPackageProfile)
	for _, studentPackage := range resp.StudentPackageProfiles {
		if studentPackage.GetCourseId() != "" {
			studentPackageIdWithPackage[studentPackage.GetCourseId()] = studentPackage
		} else {
			studentPackageIdWithPackage[studentPackage.GetStudentCoursePackageId()] = studentPackage
		}
	}

	var packageResp *pb.UpsertStudentCoursePackageResponse_StudentPackageProfile
	for _, packageReq := range req.StudentPackageProfiles {
		if packageReq.GetCourseId() != "" {
			packageResp = studentPackageIdWithPackage[packageReq.GetCourseId()]
		} else {
			packageResp = studentPackageIdWithPackage[packageReq.GetStudentPackageId()]
		}

		switch {
		case !stringutil.SliceEqual(packageReq.LocationIds, packageResp.LocationIds):
			return fmt.Errorf("expect student package locations: %v but got %v", packageResp.LocationIds, packageReq.LocationIds)
		case !packageReq.StartTime.AsTime().Round(time.Second).Equal(packageResp.StartTime.AsTime().Round(time.Second)):
			return fmt.Errorf("expect student package start at %v but got %v", packageResp.StartTime.AsTime().String(), packageReq.StartTime.AsTime().String())
		case !packageReq.EndTime.AsTime().Round(time.Second).Equal(packageResp.EndTime.AsTime().Round(time.Second)):
			return fmt.Errorf("expect student package end at %v but got %v", packageResp.EndTime.AsTime().String(), packageReq.EndTime.AsTime().String())
			// we don't have courseID in request and response incase updating student_package existed
		case packageResp.CourseId != "" && packageReq.GetCourseId() != packageResp.CourseId:
			return fmt.Errorf("expect student package course: %v but got %v", packageResp.CourseId, packageReq.GetCourseId())
		}
	}

	return nil
}

// need to wait for user sync from bob db to fatima db
func tryAddStudentPackage(ctx context.Context, client *grpc.ClientConn, req *fpb.AddStudentPackageCourseRequest) (*fpb.AddStudentPackageCourseResponse, error) {
	var (
		resp = &fpb.AddStudentPackageCourseResponse{}
		err  error
	)

	err = try.Do(func(attempt int) (bool, error) {
		resp, err = fpb.NewSubscriptionModifierServiceClient(client).AddStudentPackageCourse(contextWithToken(ctx), req)
		if err == nil {
			return false, nil
		}
		if attempt < retryTimes {
			time.Sleep(time.Second)
			return true, err
		}
		return false, err
	})

	return resp, err
}

// need to wait for user sync from bob db to fatima db
func tryUpsertStudentPackage(ctx context.Context, client *grpc.ClientConn, req *pb.UpsertStudentCoursePackageRequest) (*pb.UpsertStudentCoursePackageResponse, error) {
	var (
		resp = &pb.UpsertStudentCoursePackageResponse{}
		err  error
	)

	err = try.Do(func(attempt int) (bool, error) {
		resp, err = pb.NewUserModifierServiceClient(client).UpsertStudentCoursePackage(ctx, req)
		if err == nil {
			return false, nil
		}
		if attempt < retryTimes {
			time.Sleep(time.Second)
			return true, err
		}
		return false, err
	})

	return resp, err
}
