package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) anInvalidAuthenticationToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = invalidToken
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrievesStudentProfile(ctx context.Context, req *pb.GetStudentProfileRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).GetStudentProfile(contextWithToken(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRetrievesStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.GetStudentProfileRequest{
		StudentIds: stepState.OtherStudentIDs,
	}
	stepState.Request = req
	return s.retrievesStudentProfile(ctx, req)
}

func (s *suite) anOtherStudentProfileInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := newID()

	if ctx, err := s.aValidStudentProfileInDB(ctx, id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.OtherStudentIDs = append(stepState.OtherStudentIDs, id)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentWithName(ctx context.Context, id, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	randomID := newID()
	firstName := fmt.Sprintf("valid-student-first-name-%s", randomID)
	lastName := fmt.Sprintf("valid-student-last-name-%s", randomID)
	if name == "" {
		name = helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)
	}

	student, err := newStudentEntity()
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newStudentEntity")
	}
	if err := multierr.Combine(
		student.ID.Set(id),
		student.FullName.Set(name),
		student.LegacyUser.ID.Set(id),
		student.SchoolID.Set(stepState.CurrentSchoolID),
		student.LegacyUser.ResourcePath.Set(fmt.Sprint(stepState.CurrentSchoolID)),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = database.ExecInTx(ctx, s.BobPostgresDB, func(ctx context.Context, tx pgx.Tx) error {
		if err := (&repository.StudentRepo{}).Create(ctx, tx, student); err != nil {
			return errors.Wrap(err, "s.StudentRepo.CreateTx")
		}

		if student.AppleUser.ID.String != "" {
			if err := (&repository.AppleUserRepo{}).Create(ctx, tx, &student.AppleUser); err != nil {
				return errors.Wrap(err, "s.AppleUserRepo.Create")
			}
		}
		return nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudentWithName(ctx context.Context, id, name string) (context.Context, error) {
	if ctx, err := s.createStudentWithName(ctx, id, name); err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.StudentID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudentProfileInDB(ctx context.Context, id string) (context.Context, error) {
	return s.aValidStudentWithName(ctx, id, "")
}

func (s *suite) returnsRequestedStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.GetStudentProfileRequest)
	resp := stepState.Response.(*pb.GetStudentProfileResponse)

	for _, studentID := range req.StudentIds {
		found := false
		for _, profile := range resp.GetProfiles() {
			if profile.Id == stepState.CurrentUserID {
				continue
			}
			if profile.Id == studentID {
				found = true
			}
			if found && !isBasicProfile(profile) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("return profile is not basic profile")
			}
		}

		if !found {
			return StepStateToContext(ctx, stepState), errors.Errorf("expecting return request student profile, requester id:%s, requested ids: %v", stepState.CurrentUserID, req.StudentIds)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsRequesterStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetStudentProfileResponse)

	found := false
	for _, profile := range resp.GetProfiles() {
		if profile.Id == stepState.CurrentUserID {
			found = true
		}
	}

	if !found {
		return StepStateToContext(ctx, stepState), errors.Errorf("expecting return request student profile, requester id:%s, requested id: %s", stepState.CurrentUserID, resp.GetProfiles()[0].Id)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsEmptyStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetStudentProfileResponse)

	if len(resp.GetProfiles()) != 0 {
		return StepStateToContext(ctx, stepState), errors.Errorf("expecting return empty student profile, requester id:%s, res:%+v", stepState.CurrentUserID, resp.GetProfiles())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsStudentProfileWithCorrectGradeInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetStudentProfileResponse)
	gradeName := stepState.GradeName
	if resp.Profiles[0].GradeName != gradeName {
		return StepStateToContext(ctx, stepState), errors.Errorf("expecting return student profile with grade name: %s, requester id:%s, but got %s", gradeName, stepState.CurrentUserID, resp.Profiles[0].GradeName)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherRetrievesAStudentProfile(ctx context.Context, kindOfStudent string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := newID()
	if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch kindOfStudent {
	case "has signed in before":
		query := "UPDATE users SET last_login_date = $1 WHERE user_id = $2"
		if _, err := s.BobPostgresDB.Exec(ctx, query, time.Now().UTC().Add(-time.Hour), &id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "newly created":
		break
	default:
		return StepStateToContext(ctx, stepState), errors.New("not supported scenario step")
	}

	req := &pb.GetStudentProfileRequest{
		StudentIds: []string{id},
	}
	stepState.Request = req
	return s.retrievesStudentProfile(ctx, req)
}

func (s *suite) teacherRetrievesTheStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.GetStudentProfileRequest{
		StudentIds: []string{stepState.CurrentStudentID},
	}
	stepState.Request = req
	return s.retrievesStudentProfile(ctx, req)
}
