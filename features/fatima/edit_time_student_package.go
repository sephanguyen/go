package fatima

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) serverMustStoreThisStudentPackageWithTime(startTime string, endTime string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	packageID := s.Request.(*pb.EditTimeStudentPackageRequest).StudentPackageId

	repo := &repositories.StudentPackageRepo{}
	p, err := repo.Get(ctx, s.DB, database.Text(packageID))
	if err != nil {
		return fmt.Errorf("err find package: %w", err)
	}

	t, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return fmt.Errorf("err parse time input: %w", err)
	}
	startAt := timestamppb.New(t)
	t, err = time.Parse(time.RFC3339, endTime)
	if err != nil {
		return fmt.Errorf("err parse time input: %w", err)
	}
	endAt := timestamppb.New(t)
	if !p.StartAt.Time.Equal(startAt.AsTime()) || !p.EndAt.Time.Equal(endAt.AsTime()) {
		return fmt.Errorf("time does not match")
	}

	var locationIDs []string
	if err = p.LocationIDs.AssignTo(&locationIDs); err != nil {
		return fmt.Errorf("err convert: %w", err)
	}

	for _, locationID := range s.Request.(*pb.EditTimeStudentPackageRequest).LocationIds {
		if !golibs.InArrayString(locationID, locationIDs) {
			return fmt.Errorf("location does not match")
		}
	}

	err = s.validateStudentPackageAccessPath(ctx, p)
	if err != nil {
		return fmt.Errorf("err s.validateStudentPackageAccessPath: %w", err)
	}

	return nil
}

func (s *suite) serverMustStoreThisStudentPackageWithTimeAndClass(startTime string, endTime string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	packageID := s.Request.(*pb.EditTimeStudentPackageRequest).StudentPackageId

	repo := &repositories.StudentPackageRepo{}
	studentPackage, err := repo.Get(ctx, s.DB, database.Text(packageID))
	if err != nil {
		return fmt.Errorf("err find package: %w", err)
	}

	t, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return fmt.Errorf("err parse time input: %w", err)
	}
	startAt := timestamppb.New(t)
	t, err = time.Parse(time.RFC3339, endTime)
	if err != nil {
		return fmt.Errorf("err parse time input: %w", err)
	}
	endAt := timestamppb.New(t)
	if !studentPackage.StartAt.Time.Equal(startAt.AsTime()) || !studentPackage.EndAt.Time.Equal(endAt.AsTime()) {
		return fmt.Errorf("time does not match")
	}

	err = s.validateStudentPackageAccessPath(ctx, studentPackage)
	if err != nil {
		return fmt.Errorf("err s.validateStudentPackageAccessPath: %w", err)
	}
	err = s.validateStudentPackageClass(ctx, studentPackage)
	if err != nil {
		return fmt.Errorf("err s.validateStudentPackageClass: %w", err)
	}

	select {
	case <-s.FoundChanForJetStream:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) userEditTimeAStudentPackage(packageID string) error {
	req, err := s.editTimeStudentPackageRequest(packageID)
	if err != nil {
		return err
	}

	s.editTimeStudentPackage(req)

	return nil
}

func (s *suite) userEditTimeAStudentPackageWithTime(packageID string, startTime string, endTime string) error {
	req, err := s.editTimeStudentPackageRequest(packageID, startTime, endTime)
	if err != nil {
		return err
	}

	s.editTimeStudentPackage(req)

	return nil
}

func (s *suite) editTimeStudentPackage(req *pb.EditTimeStudentPackageRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.Response, s.ResponseErr = pb.NewSubscriptionModifierServiceClient(s.Conn).EditTimeStudentPackage(contextWithToken(s, ctx), req)
}

func (s *suite) userEditTimeAStudentPackageWithTimeAndStudentPackageExtra(packageID string, startTime string, endTime string) error {
	req, err := s.editTimeStudentPackageRequestWithStudentPackageExtra(packageID, startTime, endTime)
	if err != nil {
		return err
	}

	s.editTimeStudentPackage(req)

	return nil
}

func (s *suite) editTimeStudentPackageRequestWithStudentPackageExtra(packageID string, timeParam ...string) (*pb.EditTimeStudentPackageRequest, error) {
	req := &pb.EditTimeStudentPackageRequest{}
	now := time.Now()
	startAt := timestamppb.Now()
	endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))
	if len(timeParam) > 0 {
		t, err := time.Parse(time.RFC3339, timeParam[0])
		if err != nil {
			return nil, fmt.Errorf("err parse time input: %w", err)
		}
		startAt = timestamppb.New(t)
		t, err = time.Parse(time.RFC3339, timeParam[1])
		if err != nil {
			return nil, fmt.Errorf("err parse time input: %w", err)
		}
		endAt = timestamppb.New(t)
	}

	switch packageID {
	case "empty":
		req.StudentPackageId = ""
		req.StartAt = startAt
		req.EndAt = endAt
	case "not exist":
		req.StudentPackageId = ksuid.New().String()
		req.StartAt = startAt
		req.EndAt = endAt
	case "valid id":
		s.userAddCourseWithStudentPackageExtraForAStudent()
		req.StudentPackageId = s.Response.(*pb.AddStudentPackageCourseResponse).StudentPackageId
		req.StartAt = startAt
		req.EndAt = endAt
		studentPackageExtras := make([]*pb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra, 0)
		newClassId := idutil.ULIDNow()
		newLocationID := constants.ManabieOrgLocation
		s.ClassIDs = append(s.ClassIDs, newClassId)
		s.LocationIDs = append(s.LocationIDs, newLocationID)
		studentPackageExtras = append(studentPackageExtras, &pb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra{
			CourseId:   s.CourseIDs[0],
			LocationId: newLocationID,
			ClassId:    newClassId,
		})
		req.StudentPackageExtra = studentPackageExtras
	}
	s.Request = req

	return req, nil
}

func (s *suite) editTimeStudentPackageRequest(packageID string, timeParam ...string) (*pb.EditTimeStudentPackageRequest, error) {
	req := &pb.EditTimeStudentPackageRequest{}
	now := time.Now()
	startAt := timestamppb.Now()
	endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))
	if len(timeParam) > 0 {
		t, err := time.Parse(time.RFC3339, timeParam[0])
		if err != nil {
			return nil, fmt.Errorf("err parse time input: %w", err)
		}
		startAt = timestamppb.New(t)
		t, err = time.Parse(time.RFC3339, timeParam[1])
		if err != nil {
			return nil, fmt.Errorf("err parse time input: %w", err)
		}
		endAt = timestamppb.New(t)
	}

	switch packageID {
	case "empty":
		req.StudentPackageId = ""
		req.StartAt = startAt
		req.EndAt = endAt
	case "not exist":
		req.StudentPackageId = ksuid.New().String()
		req.StartAt = startAt
		req.EndAt = endAt
	case "valid id":
		s.userAddACourseForAStudent()
		req.StudentPackageId = s.Response.(*pb.AddStudentPackageCourseResponse).StudentPackageId
		req.StartAt = startAt
		req.EndAt = endAt
		req.LocationIds = []string{constants.ManabieOrgLocation}
	}
	s.Request = req

	return req, nil
}
