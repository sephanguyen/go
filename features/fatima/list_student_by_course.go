package fatima

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) fatimaMustReturnCorrectListOfBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when call ListStudentByCourse %w", stepState.ResponseErr)
	}
	resp := stepState.Response.(*pb.ListStudentByCourseResponse)
	if (stepState.NumberOfId) != len(resp.Profiles) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total student retrieved not correctly, expected %d - got %d", stepState.NumberOfId, len(resp.Profiles))
	}

	studentIDs := []string{}
	if len(resp.Profiles) > 0 {
		if !golibs.InArrayString(resp.Profiles[0].UserId, stepState.StudentIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("0 error expected student_id %s in %v", resp.Profiles[0].UserId, stepState.StudentIDs)
		}
		studentIDs = append(studentIDs, resp.Profiles[0].UserId)
	}
	for i := 1; i < len(resp.Profiles); i++ {
		if resp.Profiles[i].Name < resp.Profiles[i-1].Name {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error alphabet sort by student name")
		}
		if !golibs.InArrayString(resp.Profiles[i].UserId, stepState.StudentIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error expected student_id %s in %v", resp.Profiles[i].UserId, stepState.StudentIDs)
		}
		studentIDs = append(studentIDs, resp.Profiles[i].UserId)
	}
	studentIDs = golibs.Uniq(studentIDs)
	if len(studentIDs) != stepState.NumberOfId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total uniq student retrieved not correctly, expected %d - got %d", stepState.NumberOfId, len(studentIDs))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListStudentByCourseValidRequestPayloadWith(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch condition {
	case "invalid course id":
		stepState.NumberOfId = 0
		stepState.Request = &pb.ListStudentByCourseRequest{
			CourseId: condition,
			Paging:   &cpb.Paging{Limit: 5},
		}

	case "100 records":
		stepState.NumberOfId = 100
		err := s.someStudentPackagesDataInDB(ctx, stepState.NumberOfId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.someStudentPackagesDataInDB err: %v", err)
		}
		stepState.Request = &pb.ListStudentByCourseRequest{
			CourseId: stepState.CourseIDs[0],
			Paging:   &cpb.Paging{Limit: 100},
		}

	case "location ids":
		stepState.NumberOfId = 5
		err := s.someStudentPackagesDataInDB(ctx, stepState.NumberOfId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.someStudentPackagesDataInDB err: %v", err)
		}
		userIDs := stepState.StudentIDs

		uapRepo := &repository.UserAccessPathRepo{}
		uapEnts := []*entity.UserAccessPath{}
		for i := 0; i < len(userIDs); i++ {
			uapEnt := &entity.UserAccessPath{}
			database.AllNullEntity(uapEnt)
			if err := multierr.Combine(
				uapEnt.UserID.Set(userIDs[i]),
				uapEnt.LocationID.Set(constants.ManabieOrgLocation),
				uapEnt.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
			); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			uapEnts = append(uapEnts, uapEnt)
		}
		err = uapRepo.Upsert(contextWithToken(s, ctx), s.BobDB, uapEnts)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.Request = &pb.ListStudentByCourseRequest{
			CourseId:    stepState.CourseIDs[0],
			Paging:      &cpb.Paging{Limit: 5},
			LocationIds: []string{constants.ManabieOrgLocation},
		}

	case "paging":
		stepState.NumberOfId = 20
		err := s.someStudentPackagesDataInDB(ctx, stepState.NumberOfId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.someStudentPackagesDataInDB err: %v", err)
		}
		courseReader := pb.NewCourseReaderServiceClient(conn)
		resp, err := courseReader.ListStudentByCourse(contextWithToken(s, ctx), &pb.ListStudentByCourseRequest{
			CourseId: stepState.CourseIDs[0],
			Paging:   &cpb.Paging{Limit: 10},
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.NumberOfId = 10
		stepState.Request = &pb.ListStudentByCourseRequest{
			CourseId: stepState.CourseIDs[0],
			Paging:   resp.NextPage,
		}

	case "search text with Japanese student":
		stepState.NumberOfId = 1
		err := s.someStudentPackagesDataInDB(ctx, stepState.NumberOfId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.someStudentPackagesDataInDB err: %v", err)
		}
		if err := s.aJapaneseStudent(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.aJapaneseStudent err: %v", err)
		}
		stepState.Request = &pb.ListStudentByCourseRequest{
			CourseId:   stepState.CourseIDs[0],
			SearchText: "に歩",
			Paging:     &cpb.Paging{Limit: 1},
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) callListStudentByCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseReader := pb.NewCourseReaderServiceClient(conn)

	resp, err := courseReader.ListStudentByCourse(contextWithToken(s, ctx), stepState.Request.(*pb.ListStudentByCourseRequest))
	if err != nil {
		stepState.ResponseErr = err
		return nil, err
	}
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentPackagesDataInDB(ctx context.Context, amountID int) error {
	studentIDs, err := s.insertMultiStudentIntoBob(ctx, amountID)
	if err != nil {
		return err
	}

	s.CourseID = s.CourseIDs[0]
	_, err = s.callMultipleAddStudentPackage(contextWithToken(s, ctx), studentIDs, s.CourseIDs[0])
	if err != nil {
		return err
	}

	return err
}

func (s *suite) callMultipleAddStudentPackage(ctx context.Context, studentIDs []string, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, id := range studentIDs {
		now := time.Now()
		startAt := timestamppb.Now()
		endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))

		req := &pb.AddStudentPackageCourseRequest{
			StudentId:   id,
			CourseIds:   []string{courseID},
			StartAt:     startAt,
			EndAt:       endAt,
			LocationIds: []string{constants.ManabieOrgLocation},
		}

		_, err := pb.NewSubscriptionModifierServiceClient(s.Conn).AddStudentPackageCourse(contextWithToken(s, ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.StudentIDs = append(stepState.StudentIDs, id)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertMultiStudentIntoBob(ctx context.Context, amountID int) ([]string, error) {
	studentIDs := []string{}
	for i := 0; i < amountID; i++ {
		studentID, err := s.insertStudentIntoBob(ctx)
		if err != nil {
			return nil, err
		}

		studentIDs = append(studentIDs, studentID)
	}
	return studentIDs, nil
}

func (s *suite) aJapaneseStudent(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	stmtTpl := `UPDATE "users" SET name='に歩きたい乗り' WHERE user_id=$1`
	_, err := bobDB.Exec(ctx, stmtTpl, stepState.StudentIDs[0])
	if err != nil {
		return fmt.Errorf("bobDB.Exec err: %v", err)
	}
	return nil
}

func (s *suite) insertStudentIntoBob(ctx context.Context) (string, error) {
	student, err := newStudentEntity()
	if err != nil {
		return "", err
	}

	if err := (&repository.StudentRepo{}).Create(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool)), s.BobDB, student); err != nil {
		return "", errors.Wrap(err, "s.StudentRepo.CreateTx")
	}

	return student.ID.String, err
}
