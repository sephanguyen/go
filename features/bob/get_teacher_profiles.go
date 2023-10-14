package bob

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//nolint:errcheck
func (s *suite) aInvalidTeacherProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacher := entities.Teacher{}
	database.AllNullEntity(&teacher.User)
	database.AllNullEntity(&teacher)
	teacher.ID.Set(id)
	var schoolIDs []int32
	if len(stepState.Schools) > 0 {
		schoolIDs = []int32{stepState.Schools[0].ID.Int}
	}
	if schoolID != 0 {
		schoolIDs = append(schoolIDs, schoolID)
	}
	now := time.Now()
	err := multierr.Combine(
		teacher.CreatedAt.Set(now),
		teacher.UpdatedAt.Set(now),
		teacher.DeletedAt.Set(now),
		teacher.SchoolIDs.Set(schoolIDs),
	)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}
	num := rand.Int()
	user := entities.User{}
	database.AllNullEntity(&user)
	err = multierr.Combine(
		user.ID.Set(teacher.ID),
		user.LastName.Set(fmt.Sprintf("valid-teacher-%d", num)),
		user.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		user.Email.Set(fmt.Sprintf("valid-teacher-%d@email.com", num)),
		user.Avatar.Set(fmt.Sprintf("http://valid-teacher-%d", num)),
		user.Country.Set(pb.COUNTRY_VN.String()),
		user.Group.Set(entities.UserGroupTeacher),
		user.DeviceToken.Set(nil),
		user.AllowNotification.Set(true),
		user.CreatedAt.Set(teacher.CreatedAt),
		user.UpdatedAt.Set(teacher.UpdatedAt),
		user.DeletedAt.Set(teacher.DeletedAt),
		user.IsTester.Set(nil),
		user.FacebookID.Set(nil),
	)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}
	userGroup := entities.UserGroup{}
	err = multierr.Combine(
		userGroup.UserID.Set(teacher.ID),
		userGroup.GroupID.Set(database.Text(pb.USER_GROUP_TEACHER.String())),
		userGroup.IsOrigin.Set(database.Bool(true)),
		userGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
		userGroup.CreatedAt.Set(teacher.CreatedAt),
		userGroup.UpdatedAt.Set(teacher.UpdatedAt),
	)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}
	_, err = database.InsertExcept(ctx, &user, []string{"resource_path"}, s.DB.Exec)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("InsertExcept user fail: %w", err).Error())
	}
	_, err = database.InsertExcept(ctx, &teacher, []string{"resource_path"}, s.DB.Exec)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("InsertExcept teacher fail: %w", err).Error())
	}
	cmdTag, err := database.InsertExcept(ctx, &userGroup, []string{"resource_path"}, s.DB.Exec)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("InsertExcept user group fail: %w", err).Error())
	}
	if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert teacher for testing")
	}
	return ctx, nil
}

func (s *suite) aValidTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	stepState.Request = &pb.GetTeacherProfilesRequest{
		Ids: []string{id},
	}
	return s.aValidTeacherProfileWithId(ctx, id, 0)
}
func (s *suite) userRetrievesTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).GetTeacherProfiles(s.signedCtx(ctx), stepState.Request.(*pb.GetTeacherProfilesRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnsTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	if stepState.Response == nil {
		return StepStateToContext(ctx, stepState), errors.New("bob does not return user profile")
	}
	profiles := stepState.Response.(*pb.GetTeacherProfilesResponse).Profiles
	userIds := stepState.Request.(*pb.GetTeacherProfilesRequest).Ids
	query := "SELECT count(*) FROM teachers where teacher_id = ANY($1)"
	row := s.DB.QueryRow(ctx, query, userIds)
	var count int
	row.Scan(&count)
	if count != len(profiles) {
		return StepStateToContext(ctx, stepState), errors.New("bob did not return all teacher")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aInvalidTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	stepState.Request = &pb.GetTeacherProfilesRequest{
		Ids: []string{id},
	}
	return s.aInvalidTeacherProfileWithId(ctx, id, constants.ManabieSchool)
}

func (s *suite) bobMustReturnsTeacherProfileNotFound(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	if stepState.Response == nil {
		return StepStateToContext(ctx, stepState), errors.New("bob does not return user profile")
	}
	profiles := stepState.Response.(*pb.GetTeacherProfilesResponse).Profiles
	if len(profiles) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), errors.New("bob does return user profile")
}
