package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareStudentsData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentRepo := &repository.StudentRepo{}

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	for i := 0; i < 20; i++ {
		id := newID()
		if _, err := s.aValidStudentInDB(ctx, id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentEnt, err := studentRepo.Find(auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, database.Text(id))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		s.ExistingStudents = append(s.ExistingStudents, studentEnt)
	}

	return StepStateToContext(ctx, stepState), nil
}

func addLocationToUser(userID string, location string) (*entity.UserAccessPath, error) {
	uapEnt := &entity.UserAccessPath{}
	database.AllNullEntity(uapEnt)
	if err := multierr.Combine(
		uapEnt.UserID.Set(userID),
		uapEnt.LocationID.Set(location),
		uapEnt.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
	); err != nil {
		return nil, err
	}
	return uapEnt, nil
}

func (s *suite) searchBasicProfileRequestWithStudentIDsAnd(ctx context.Context, filter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	var userIDs []string
	for i := 0; i < len(s.ExistingStudents); i++ {
		userIDs = append(userIDs, s.ExistingStudents[i].ID.String)
	}

	stepState.Request = &pb.SearchBasicProfileRequest{
		UserIds: userIDs,
		Paging:  &cpb.Paging{Limit: 100},
	}

	switch filter {
	case "none":
		s.StudentIds = userIDs
		stepState.NumberOfIds = len(userIDs)
	case "paging":
		stepState.Request.(*pb.SearchBasicProfileRequest).Paging = &cpb.Paging{Limit: 10}
		ctxTemp, err := s.searchBasicProfile(StepStateToContext(ctx, stepState), schoolAdminType)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState := StepStateFromContext(ctxTemp)
		s.StudentIds = userIDs
		stepState.NumberOfIds = 10
		stepState.Request = &pb.SearchBasicProfileRequest{
			UserIds: userIDs,
			Paging:  stepState.Response.(*pb.SearchBasicProfileResponse).NextPage,
		}
		stepState.Response = nil
	case "location_ids":
		// add location to user
		uapRepo := &repository.UserAccessPathRepo{}
		uapEnts := []*entity.UserAccessPath{}
		for i := 0; i < (len(s.ExistingStudents) / 2); i++ {
			uapEnt1, err := addLocationToUser(s.ExistingStudents[i].ID.String, constants.ManabieOrgLocation)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			uapEnts = append(uapEnts, uapEnt1)

			uapEnt2, err := addLocationToUser(s.ExistingStudents[i].ID.String, constants.JPREPOrgLocation)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			uapEnts = append(uapEnts, uapEnt2)

			s.StudentIds = append(s.StudentIds, s.ExistingStudents[i].ID.String)
		}
		err := uapRepo.Upsert(auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool)), s.BobDBTrace, uapEnts)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.NumberOfIds = len(s.ExistingStudents) / 2
		stepState.Request.(*pb.SearchBasicProfileRequest).LocationIds = []string{constants.ManabieOrgLocation, constants.JPREPOrgLocation}
	case "search_text":
		// add search_text
		for i := 0; i < (len(s.ExistingStudents) / 2); i++ {
			s.StudentIds = append(s.StudentIds, s.ExistingStudents[i].ID.String)
		}
		stepState.NumberOfIds = len(s.StudentIds)

		newName := "new-name-vippro"
		stmt := `UPDATE "users" SET name = $1 WHERE user_id = ANY($2)`
		_, err := s.BobDBTrace.Exec(auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID)), stmt, newName, database.TextArray(s.StudentIds))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Request.(*pb.SearchBasicProfileRequest).SearchText = &wrapperspb.StringValue{Value: "new-name-vippro"}
	case "search_text_phonetic_name":
		// add search_text
		for i := 0; i < (len(s.ExistingStudents) / 2); i++ {
			s.StudentIds = append(s.StudentIds, s.ExistingStudents[i].ID.String)
		}
		stepState.NumberOfIds = len(s.StudentIds)

		newFullNamePhonetic := "久保田 聖良"
		stmt := `UPDATE "users" SET full_name_phonetic = $1 WHERE user_id = ANY($2)`
		_, err := s.BobDBTrace.Exec(auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID)), stmt, newFullNamePhonetic, database.TextArray(s.StudentIds))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Request.(*pb.SearchBasicProfileRequest).SearchText = &wrapperspb.StringValue{Value: "久保田 聖良"}
	case "search_text_combine_full_name_and_phonetic_name":
		// add search_text
		for i := 0; i < (len(s.ExistingStudents) / 2); i++ {
			s.StudentIds = append(s.StudentIds, s.ExistingStudents[i].ID.String)
		}
		stepState.NumberOfIds = len(s.StudentIds)

		newCombineName := "Combined name"
		toggle := false
		for _, studentId := range s.StudentIds {
			if !toggle {
				stmt := `UPDATE "users" SET name = $1 WHERE user_id = $2`
				_, err := s.BobDBTrace.Exec(auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID)), stmt, newCombineName, database.Text(studentId))
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				toggle = true
			} else {
				stmt := `UPDATE "users" SET full_name_phonetic = $1 WHERE user_id = $2`
				_, err := s.BobDBTrace.Exec(auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID)), stmt, newCombineName, database.Text(studentId))
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				toggle = false
			}
		}

		stepState.Request.(*pb.SearchBasicProfileRequest).SearchText = &wrapperspb.StringValue{Value: newCombineName}
	case "search_text_only_first_name_or_first_name_phonetic":
		// add search_text
		for i := 0; i < (len(s.ExistingStudents) / 2); i++ {
			s.StudentIds = append(s.StudentIds, s.ExistingStudents[i].ID.String)
		}
		stepState.NumberOfIds = len(s.StudentIds)

		newCombineName := "Khanh Le"
		toggle := false
		for _, studentId := range s.StudentIds {
			if !toggle {
				stmt := `UPDATE "users" SET name = $1 WHERE user_id = $2`
				_, err := s.BobDBTrace.Exec(auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID)), stmt, newCombineName, database.Text(studentId))
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				toggle = true
			} else {
				stmt := `UPDATE "users" SET full_name_phonetic = $1 WHERE user_id = $2`
				_, err := s.BobDBTrace.Exec(auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.CurrentSchoolID)), stmt, newCombineName, database.Text(studentId))
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				toggle = false
			}
		}
		stepState.Request.(*pb.SearchBasicProfileRequest).SearchText = &wrapperspb.StringValue{Value: "Le"}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) searchBasicProfile(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response, stepState.ResponseErr = pb.NewUserReaderServiceClient(s.UserMgmtConn).SearchBasicProfile(contextWithToken(ctx), stepState.Request.(*pb.SearchBasicProfileRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnAListBasicProfileCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp := stepState.Response.(*pb.SearchBasicProfileResponse)
	if (stepState.NumberOfIds) != len(resp.Profiles) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total student retrieved not correctly, expected %d - got %d", stepState.NumberOfIds, len(resp.Profiles))
	}

	for _, v := range resp.Profiles {
		if !golibs.InArrayString(v.UserId, s.StudentIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err returnAListBasicProfileCorrectly: %v not in %v", v.UserId, s.StudentIds)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
