package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	notificationEntities "github.com/manabie-com/backend/internal/notification/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"firebase.google.com/go/v4/auth"
	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	JSM            nats.JetStreamManagement
	DB             database.Ext
	FirebaseClient *auth.Client
	UserRepo       repositories.UserRepository
	StudentRepo    interface {
		Find(context.Context, database.QueryExecer, pgtype.Text) (*entities.Student, error)
	}
	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error)
	}
	ActivityLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.ActivityLog) error
	}
	TeacherRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error)
		FindByID(context.Context, database.QueryExecer, pgtype.Text) (*entities.Teacher, error)
	}
	AppleUserRepo interface {
		Get(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entities.AppleUser, error)
	}
	UserDeviceTokenRepo interface {
		UpsertUserDeviceToken(ctx context.Context, db database.QueryExecer, u *notificationEntities.UserDeviceToken) error
	}
}

func (s *UserService) UpdateUserDeviceToken(ctx context.Context, req *pb.UpdateUserDeviceTokenRequest) (*pb.UpdateUserDeviceTokenResponse, error) {
	if req.UserId == "" || req.DeviceToken == "" {
		return nil, status.Error(codes.InvalidArgument, "userID or device token empty")
	}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// user := entities.User{}
		// user.ID.Set(req.UserId)
		// user.DeviceToken.Set(req.DeviceToken)
		// user.AllowNotification.Set(req.AllowNotification)
		// err := s.UserRepo.StoreDeviceToken(ctx, tx, &user)
		// if err != nil {
		// 	return fmt.Errorf("s.UserRepo.StoreDeviceToken: %w", err)
		// }

		userDeviceToken := notificationEntities.UserDeviceToken{}
		userDeviceToken.UserID.Set(req.UserId)
		userDeviceToken.DeviceToken.Set(req.DeviceToken)
		userDeviceToken.AllowNotification.Set(req.AllowNotification)
		err := s.UserDeviceTokenRepo.UpsertUserDeviceToken(ctx, tx, &userDeviceToken)
		if err != nil {
			return fmt.Errorf("s.UserDeviceTokenRepo.UpsertUserDeviceToken: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	userProfileModels, err := s.UserRepo.Retrieve(ctx, s.DB, database.TextArray([]string{req.UserId}))
	if err != nil {
		return nil, fmt.Errorf("s.UserRepo.Retrieve: %w", err)
	}

	data := &pb.EvtUserInfo{
		UserId:            req.UserId,
		DeviceToken:       req.DeviceToken,
		AllowNotification: req.AllowNotification,
		Name:              userProfileModels[0].GetName(),
	}
	msg, _ := data.Marshal()

	var msgID string
	msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectUserDeviceTokenUpdated, msg)
	if err != nil {
		ctxzap.Extract(ctx).Error("UpdateUserDeviceToken s.JSM.PublishAsyncContext failed", zap.String("msg-id", msgID), zap.Error(err))
	}

	if err != nil {
		// TODO: store msg can not push
	}

	return &pb.UpdateUserDeviceTokenResponse{
		Successful: true,
	}, nil
}

func (s *UserService) checkPermissionUpdateUser(ctx context.Context, profile *pb.UserProfile, currentUserID, currentUGroup string) error {
	// only school admin can update teacher at their school
	if currentUGroup != entities.UserGroupSchoolAdmin || profile.UserGroup != entities.UserGroupTeacher {
		return status.Error(codes.PermissionDenied, "user can only update own profile")
	}

	schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return errors.Wrapf(err, "s.SchoolAdminRepo.Get: userID: %q", currentUserID)
	}
	if schoolAdmin == nil {
		return status.Error(codes.PermissionDenied, "only school admin can update their teacher profile")
	}

	if profile.UserGroup == entities.UserGroupTeacher {
		teacher, err := s.TeacherRepo.FindByID(ctx, s.DB, database.Text(profile.Id))
		if err != nil {
			return errors.Wrapf(err, "s.TeacherRepo.FindByID: userID: %q", profile.Id)
		}
		if teacher == nil {
			return status.Error(codes.InvalidArgument, "teacher profile not found")
		}
		if len(teacher.SchoolIDs.Elements) == 0 {
			return status.Error(codes.PermissionDenied, "school admin can only update their teacher profile")
		}

		flag := func() int {
			for _, v := range teacher.SchoolIDs.Elements {
				if v == schoolAdmin.SchoolID {
					return 1
				}
			}
			return 0
		}
		if flag() == 0 {
			return status.Error(codes.PermissionDenied, "school admin can only update their teacher profile")
		}
	}
	return nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	profile := req.Profile
	if profile.Name == "" || profile.UserGroup == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid profile")
	}
	currentUserID := interceptors.UserIDFromContext(ctx)
	currentUGroup, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return nil, errors.Wrapf(err, "s.UserRepo.UserGroup: userID: %q", currentUserID)
	}
	if profile.Id == "" {
		profile.Id = currentUserID
	}

	if currentUserID != profile.Id && currentUGroup != entities.UserGroupAdmin {
		err = s.checkPermissionUpdateUser(ctx, profile, currentUserID, currentUGroup)
		if err != nil {
			return nil, err
		}
	}
	if currentUGroup == entities.UserGroupAdmin && currentUserID != profile.Id && profile.UserGroup == entities.UserGroupAdmin {
		return nil, status.Error(codes.PermissionDenied, "user can only update own profile")
	}

	e := toUserEntity(profile)
	err = s.UserRepo.UpdateProfile(ctx, s.DB, e)
	if err != nil {
		return nil, err
	}

	data := &pb.EvtUserInfo{
		UserId: req.Profile.Id,
		Name:   e.GetName(),
	}
	msg, _ := data.Marshal()

	var msgId string
	msgId, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectUserDeviceTokenUpdated, msg)
	if err != nil {
		ctxzap.Extract(ctx).Error("UpdateUserDeviceToken s.JSM.PublishAsyncContext failed", zap.String("msg-id", msgId), zap.Error(err))
	}

	if err != nil {
		// TODO: store msg can not push
	}

	return &pb.UpdateUserProfileResponse{
		Successful: true,
	}, nil
}

func (s *UserService) GetCurrentUserProfile(ctx context.Context, req *pb.GetCurrentUserProfileRequest) (*pb.GetCurrentUserProfileResponse, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	ids := []string{
		currentUserID,
	}
	userProfileModel, err := s.UserRepo.Retrieve(ctx, s.DB, database.TextArray(ids))
	if err != nil {
		return nil, err
	}
	if len(userProfileModel) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user does not exist")
	}
	userProfile := toUserProfilePb(userProfileModel[0])
	return &pb.GetCurrentUserProfileResponse{
		Profile: userProfile,
	}, nil
}

func toUserProfilePb(src *entities.User) *pb.UserProfile {
	return &pb.UserProfile{
		Id:          src.ID.String,
		Name:        src.GetName(),
		Country:     pb.Country(pb.Country_value[src.Country.String]),
		PhoneNumber: src.PhoneNumber.String,
		Email:       src.Email.String,
		Avatar:      src.Avatar.String,
		DeviceToken: src.DeviceToken.String,
		UserGroup:   src.Group.String,
		CreatedAt:   &types.Timestamp{Seconds: src.CreatedAt.Time.Unix()},
		UpdatedAt:   &types.Timestamp{Seconds: src.UpdatedAt.Time.Unix()},
	}
}

func toUserEntity(src *pb.UserProfile) *entities.User {
	e := new(entities.User)
	database.AllNullEntity(e)

	e.ID.Set(src.Id)
	e.Avatar.Set(src.Avatar)
	e.Group.Set(src.UserGroup)
	e.LastName.Set(src.Name)
	e.Country.Set(src.Country.String())
	e.Email.Set(src.Email)
	e.DeviceToken.Set(src.DeviceToken)
	e.IsTester.Set(nil)
	if src.PhoneNumber != "" {
		e.PhoneNumber.Set(src.PhoneNumber)
	} else {
		e.PhoneNumber.Set(nil)
	}
	if src.UpdatedAt != nil {
		e.UpdatedAt.Set(time.Unix(src.UpdatedAt.Seconds, int64(src.UpdatedAt.Nanos)))
	} else {
		e.UpdatedAt.Set(time.Now())
	}
	if src.CreatedAt != nil {
		e.CreatedAt.Set(time.Unix(src.CreatedAt.Seconds, int64(src.CreatedAt.Nanos)))
	} else {
		e.CreatedAt.Set(time.Now())
	}
	return e
}

func (s *UserService) ClaimsUserAuth(ctx context.Context, req *pb.ClaimsUserAuthRequest) (*pb.ClaimsUserAuthResponse, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	userGroup, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return nil, err
	}

	hasura := map[string]interface{}{
		"x-hasura-allowed-roles": []string{userGroup},
		"x-hasura-default-role":  userGroup,
		"x-hasura-user-id":       currentUserID,
	}

	claims := map[string]interface{}{
		"https://hasura.io/jwt/claims": hasura,
	}
	err = s.FirebaseClient.SetCustomUserClaims(ctx, currentUserID, claims)
	if err != nil {
		return nil, err
	}

	return &pb.ClaimsUserAuthResponse{
		Successful: true,
	}, nil
}

func (s *UserService) GetTeacherProfiles(ctx context.Context, req *pb.GetTeacherProfilesRequest) (*pb.GetTeacherProfilesResponse, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	if len(req.Ids) == 0 {
		req.Ids = []string{currentUserID}
	}

	// TODO: add permission control
	teachers, err := s.TeacherRepo.Retrieve(ctx, s.DB, database.TextArray(req.Ids))
	if err != nil {
		return nil, status.Error(codes.Unknown, "Error finding teacher")
	}

	teacherProfiles := []*pb.TeacherProfile{}
	for _, teacher := range teachers {
		teacherProfiles = append(teacherProfiles, toTeacherProfilePb(&teacher))
	}

	return &pb.GetTeacherProfilesResponse{
		Profiles: teacherProfiles,
	}, nil
}

func toTeacherProfilePb(src *entities.Teacher) *pb.TeacherProfile {
	var schoolIds []int32
	src.SchoolIDs.AssignTo(&schoolIds)

	return &pb.TeacherProfile{
		Id:          src.User.ID.String,
		Name:        src.User.LastName.String,
		Country:     pb.Country(pb.Country_value[src.User.Country.String]),
		PhoneNumber: src.User.PhoneNumber.String,
		Email:       src.User.Email.String,
		Avatar:      src.User.Avatar.String,
		DeviceToken: src.User.DeviceToken.String,
		UserGroup:   src.User.Group.String,
		SchoolIds:   schoolIds,
		CreatedAt:   &types.Timestamp{Seconds: src.CreatedAt.Time.Unix()},
		UpdatedAt:   &types.Timestamp{Seconds: src.UpdatedAt.Time.Unix()},
	}
}

func (s *UserService) GetBasicProfile(ctx context.Context, req *pb.GetBasicProfileRequest) (*pb.GetBasicProfileResponse, error) {
	if len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing userIds")
	}

	users, err := s.UserRepo.Retrieve(ctx, s.DB, database.TextArray(req.UserIds))
	if err != nil {
		return nil, toStatusError(err)
	}

	data := make([]*pb.BasicProfile, 0, len(users))

	for _, e := range users {
		data = append(data, &pb.BasicProfile{
			UserId:     e.ID.String,
			Name:       e.GetName(),
			Avatar:     e.Avatar.String,
			FacebookId: e.FacebookID.String,
			UserGroup:  e.Group.String,
		})
	}

	return &pb.GetBasicProfileResponse{
		Profiles: data,
	}, nil
}

// CheckProfile can be use on registration screen
func (s *UserService) CheckProfile(ctx context.Context, req *pb.CheckProfileRequest) (*pb.CheckProfileResponse, error) {
	filter := repositories.UserFindFilter{}
	err := multierr.Combine(
		filter.UserGroup.Set(nil),
		filter.Email.Set(nil),
		filter.Phone.Set(nil),
		filter.IDs.Set(nil),
	)
	if err != nil {
		return nil, err
	}

	errEmptyAll := status.Error(codes.InvalidArgument, "Email or Phone must be specific")

	switch v := req.Filter.(type) {
	case *pb.CheckProfileRequest_Email:
		filter.Email.Set(v.Email)
	case *pb.CheckProfileRequest_Phone:
		filter.Phone.Set(v.Phone)
	default:
		return nil, errEmptyAll
	}

	if filter.Phone.String == "" && filter.Email.String == "" {
		return nil, errEmptyAll
	}

	users, err := s.UserRepo.Find(ctx, s.DB, &filter, "user_id", "name", "avatar", "user_group", "facebook_id")
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, status.Error(codes.NotFound, repositories.ErrUserNotFound.Error())
	}

	// check is registered with apple
	appleUser, err := s.AppleUserRepo.Get(ctx, s.DB, users[0].ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toStatusError(err)
	}

	appleID := ""
	if appleUser != nil {
		appleID = appleUser.ID.String
	}

	return &pb.CheckProfileResponse{
		Found: true,
		Profile: &pb.BasicProfile{
			UserId:      users[0].ID.String,
			Name:        users[0].GetName(),
			Avatar:      users[0].Avatar.String,
			UserGroup:   users[0].Group.String,
			FacebookId:  users[0].FacebookID.String,
			AppleUserId: appleID,
		},
	}, nil
}

func (s *UserService) Get(ctx context.Context, userID string) (*entities.User, error) {
	user, err := s.UserRepo.Get(ctx, s.DB, database.Text(userID))
	if err != nil {
		return nil, fmt.Errorf("u.UserRepo.Get: userID: %q, %w", userID, err)
	}
	return user, nil
}
