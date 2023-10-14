package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repositories "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	usermgmt_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	ppb_v1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"firebase.google.com/go/v4/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrStudentEnrollmentStatusUnknown returned when create or edit student
	// that set enrollment status to a value not in StudentEnrollmentStatus predefined values
	ErrStudentEnrollmentStatusUnknown = errors.New("student enrollment status unknown")

	// ErrStudentEnrollmentStatusNotAllowedTobeNone returned when create or edit student
	// that set enrollment status to STUDENT_ENROLLMENT_STATUS_NONE
	ErrStudentEnrollmentStatusNotAllowedTobeNone = errors.New("student enrollment status not allowed to be STUDENT_ENROLLMENT_STATUS_NONE")
)

func NewUserModifierService(c *configurations.Config, db database.Ext, jsm nats.JetStreamManagement, firebaseClient *auth.Client, firebaseAuthClient multitenant.TenantClient, tenantManager multitenant.TenantManager, fatimaCourseModifierClient fpb.SubscriptionModifierServiceClient) *UserModifierService {
	return &UserModifierService{
		DB:                 db,
		JSM:                jsm,
		FirebaseClient:     firebase.NewAuthFromApp(firebaseClient),
		FirebaseAuthClient: firebaseAuthClient,
		TenantManager:      tenantManager,
		FatimaClient:       fatimaCourseModifierClient,
		UserRepo:           &bob_repositories.UserRepo{},
		TeacherRepo:        &bob_repositories.TeacherRepo{},
		StudentRepo:        &bob_repositories.StudentRepo{},
		SchoolAdminRepo:    &bob_repositories.SchoolAdminRepo{},
		UserGroupRepo:      &bob_repositories.UserGroupRepo{},
		ParentRepo:         &bob_repositories.ParentRepo{},
		StudentParentRepo:  &bob_repositories.StudentParentRepo{},
		OrganizationRepo:   (&repository.OrganizationRepo{}).WithDefaultValue(c.Common.Environment),
		UsrEmailRepo:       &repository.UsrEmailRepo{},
	}
}

type UserModifierService struct {
	pb.UnimplementedUserModifierServiceServer
	JSM                nats.JetStreamManagement
	DB                 database.Ext
	FirebaseClient     firebase.AuthClient
	FirebaseAuthClient multitenant.TenantClient
	TenantManager      multitenant.TenantManager
	FatimaClient       fpb.SubscriptionModifierServiceClient

	UserRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities_bob.User, error)
		UpdateProfile(context.Context, database.QueryExecer, *entities_bob.User) error
		UpdateEmail(ctx context.Context, db database.QueryExecer, u *entities.User) error
		UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
		GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entities_bob.User, error)
		GetByPhone(ctx context.Context, db database.QueryExecer, phones pgtype.TextArray) ([]*entities_bob.User, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entities_bob.User) error
	}
	TeacherRepo interface {
		FindByID(context.Context, database.QueryExecer, pgtype.Text) (*entities_bob.Teacher, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entities_bob.Teacher) error
	}
	StudentRepo interface {
		Find(context.Context, database.QueryExecer, pgtype.Text) (*entities_bob.Student, error)
		Update(context.Context, database.QueryExecer, *entities_bob.Student) error
		UpdateV2(ctx context.Context, db database.QueryExecer, s *entities.Student) error
		Create(context.Context, database.QueryExecer, *entities_bob.Student) error
		FindStudentProfilesByIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities_bob.Student, error)
	}
	SchoolAdminRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities_bob.SchoolAdmin, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, schoolAdmins []*entities_bob.SchoolAdmin) error
	}
	UserGroupRepo interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entities_bob.UserGroup) error
	}
	ParentRepo interface {
		GetByIds(ctx context.Context, db database.QueryExecer, parentIds pgtype.TextArray) (entities.Parents, error)
		Create(ctx context.Context, db database.QueryExecer, parent *entities.Parent) error
		CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entities_bob.Parent) error
	}
	StudentParentRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, parents []*entities.StudentParent) error
		GetStudentParents(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.StudentParent, error)
	}

	OrganizationRepo OrganizationRepo
	UsrEmailRepo     UsrEmailRepo
}

type OrganizationRepo interface {
	GetTenantIDByOrgID(ctx context.Context, db database.QueryExecer, orgID string) (string, error)
}

type UsrEmailRepo interface {
	Create(ctx context.Context, db database.QueryExecer, usrID pgtype.Text, email pgtype.Text) (*usermgmt_entities.UsrEmail, error)
}

func IsSchoolAdmin(userGroup string) bool {
	return userGroup == constant.UserGroupSchoolAdmin
}

func UserPbToUserEn(ctx context.Context, userProfiles []*pb.CreateUserProfile, userGroup, organization string) ([]*entities_bob.User, []*entities_bob.UserGroup, error) {
	users := []*entities_bob.User{}
	userGroups := []*entities_bob.UserGroup{}

	for _, v := range userProfiles {
		u := &entities_bob.User{}
		database.AllNullEntity(u)

		if IsSchoolAdmin(userGroup) {
			u.ResourcePath.Set(organization)
		}

		err := multierr.Combine(
			u.ID.Set(idutil.ULIDNow()),
			u.LastName.Set(v.Name),
			u.GivenName.Set(v.GivenName),
			u.Country.Set(v.Country.String()),
			u.PhoneNumber.Set(v.PhoneNumber),
			u.Group.Set(userGroup),
			u.Email.Set(v.Email),
			u.Avatar.Set(v.Avatar),
		)

		if err != nil {
			return nil, nil, fmt.Errorf("err set user: %w", err)
		}

		users = append(users, u)

		// set for userGroup
		group := &entities_bob.UserGroup{}
		database.AllNullEntity(group)

		err = multierr.Combine(
			group.UserID.Set(u.ID.String),
			group.GroupID.Set(userGroup),
			group.IsOrigin.Set(true),
			group.Status.Set(entities_bob.UserGroupStatusActive),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("err set userGroup: %w", err)
		}

		userGroups = append(userGroups, group)
	}

	return users, userGroups, nil
}

func (s *UserModifierService) CheckPermissionToCreateUser(ctx context.Context, userGroup string, schoolID int64) error {
	currentUserID := interceptors.UserIDFromContext(ctx)
	currentUser, err := s.UserRepo.Get(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return err
	}

	if (userGroup == constant.UserGroupTeacher || IsSchoolAdmin(userGroup)) && schoolID == 0 {
		return status.Error(codes.InvalidArgument, "school_id cannot be empty")
	}

	checkSchoolID := func(ctx context.Context, schoolID int64, schoolAdminID string) error {
		schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DB, database.Text(schoolAdminID))
		if err != nil {
			return errors.Wrapf(err, "s.SchoolAdminRepo.Get: schoolAdminID: %s", currentUserID)
		}

		if schoolAdmin == nil {
			return status.Error(codes.InvalidArgument, "school admin not found")
		}

		if schoolAdmin.SchoolID.Int != int32(schoolID) {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("school id not match expected %d, got %d", schoolID, schoolAdmin.SchoolID.Int))
		}

		return nil
	}
	currentUserGroup := currentUser.Group.String

	//TODO: add other user group later
	switch userGroup {
	case constant.UserGroupStudent:
	case constant.UserGroupParent:
	case constant.UserGroupTeacher:
		if currentUserGroup != constant.UserGroupAdmin && currentUserGroup != constant.UserGroupSchoolAdmin {
			return status.Error(codes.PermissionDenied, "only admin, school admin and school staff can create teachers")
		}

		if currentUserGroup == constant.UserGroupSchoolAdmin {
			if err := checkSchoolID(ctx, schoolID, currentUserID); err != nil {
				return err
			}
		}
	case constant.UserGroupAdmin:
		if currentUserGroup != constant.UserGroupAdmin {
			return status.Error(codes.PermissionDenied, "only admin can create admins")
		}
	case constant.UserGroupSchoolAdmin:
		if currentUserGroup != constant.UserGroupAdmin {
			return status.Error(codes.PermissionDenied, "only admin can create school admin")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid user group")
	}

	return nil
}

func (s *UserModifierService) createStudent(ctx context.Context, tx pgx.Tx, additionalData *entities_bob.StudentAdditionalData, req *pb.CreateUserProfile, studentId string, schoolId int64) (*entities_bob.Student, error) {
	student := &entities_bob.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)

	if studentId == "" {
		studentId = idutil.ULIDNow()
	}

	err := multierr.Combine(
		student.ID.Set(studentId),
		student.PhoneNumber.Set(req.PhoneNumber),
		student.GivenName.Set(req.GivenName),
		student.LastName.Set(req.Name),
		student.Country.Set(req.Country.String()),
		student.AdditionalData.Set(additionalData),
		student.SchoolID.Set(schoolId),
		student.ResourcePath.Set(fmt.Sprint(schoolId)),
		student.CurrentGrade.Set(req.Grade),
		student.Avatar.Set(req.Avatar),
	)

	if req.Email != "" {
		err = multierr.Append(err, student.Email.Set(req.Email))
	}

	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	if err := s.StudentRepo.Create(ctx, tx, student); err != nil {
		return nil, errors.Wrap(err, "s.StudentRepo.Create")
	}

	return student, nil
}

func (s *UserModifierService) CreateUserInFirebase(ctx context.Context, users []*entities_bob.User, schoolID int64) error {
	params := []*auth.UserToImport{}
	indexUsers := map[int]*entities_bob.User{}

	for i, v := range users {
		indexUsers[i] = v
		claims := utils.CustomUserClaims(v.Group.String, v.ID.String, schoolID)

		params = append(params, (&auth.UserToImport{}).
			UID(v.ID.String).
			Email(v.Email.String).
			CustomClaims(claims))
	}

	result, err := s.FirebaseClient.ImportUsers(ctx, params)
	if err != nil {
		return errors.Wrapf(err, "s.FirebaseClient.ImportUsers")
	}

	if len(result.Errors) > 0 {
		errs := []string{}
		for _, v := range result.Errors {
			if indexUsers[v.Index] != nil {
				errs = append(errs, fmt.Sprintf("%s - %s", indexUsers[v.Index].Email.String, v.Reason))
				break
			}

			errs = append(errs, v.Reason)
		}

		return status.Error(codes.InvalidArgument, fmt.Sprintf("create user in firebase: %s", strings.Join(errs, ", ")))
	}

	if result.FailureCount > 0 {
		return status.Error(codes.InvalidArgument, "firebase can not create user")
	}

	return nil
}

func (s *UserModifierService) CreateUserInIdentityPlatform(ctx context.Context, tenantID string, users []*entities_bob.User, schoolID int64) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	tenantClient, err := s.TenantManager.TenantClient(ctx, tenantID)
	if err != nil {
		zapLogger.Warnw(
			"cannot get tenant client",
			"tenantID", tenantID,
			"err", err.Error(),
		)
		return errors.Wrap(err, "TenantClient")
	}

	var authUsers internal_auth_user.Users

	for i := range users {
		users[i].CustomClaims = utils.CustomUserClaims(users[i].Group.String, users[i].ID.String, schoolID)
		users[i].PhoneNumber.Status = pgtype.Null
		users[i].PhoneNumber = database.Text("")

		authUsers = append(authUsers, users[i])
	}

	result, err := tenantClient.ImportUsers(ctx, authUsers, nil)
	if err != nil {
		return errors.Wrapf(err, "ImportUsers")
	}

	if len(result.UsersFailedToImport) > 0 {
		var errs []string
		for _, userFailedToImport := range result.UsersFailedToImport {
			errs = append(errs, fmt.Sprintf("%s - %s", userFailedToImport.User.GetEmail(), userFailedToImport.Err))
		}
		return fmt.Errorf("create user in identity platform: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (s *UserModifierService) CreateStudents(ctx context.Context, tx pgx.Tx, schoolID int64, userProfiles []*pb.CreateUserProfile, users []*entities_bob.User) error {
	for _, v := range userProfiles {
		student, err := s.createStudent(ctx, tx, &entities_bob.StudentAdditionalData{}, v, "", schoolID)
		if err != nil {
			return fmt.Errorf("s.createStudent: %w", err)
		}

		for _, u := range users {
			if v.Email == u.Email.String {
				u.ID = student.ID
				break
			}
		}
		data := &ppb_v1.EvtUser{
			Message: &ppb_v1.EvtUser_CreateStudent_{
				CreateStudent: &ppb_v1.EvtUser_CreateStudent{
					StudentId:   student.ID.String,
					StudentName: student.GetName(),
				},
			},
		}

		err = s.publicAsyncUserEvent(ctx, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UserModifierService) CreateTeachers(ctx context.Context, tx database.QueryExecer, schoolID int64, users []*entities_bob.User) error {
	teacherEn := []*entities_bob.Teacher{}
	for _, v := range users {
		t := &entities_bob.Teacher{}
		database.AllNullEntity(t)
		t.ID = v.ID
		t.SchoolIDs.Set([]int64{schoolID})
		t.ResourcePath.Set(fmt.Sprint(schoolID))
		teacherEn = append(teacherEn, t)
	}

	err := s.TeacherRepo.CreateMultiple(ctx, tx, teacherEn)
	if err != nil {
		return errors.Wrapf(err, "s.TeacherRepo.CreateMultiple")
	}
	return nil
}

func (s *UserModifierService) CreateSchoolAdmins(ctx context.Context, tx database.QueryExecer, schoolID int64, org string, users []*entities_bob.User) error {
	schoolAdminEn := []*entities_bob.SchoolAdmin{}
	for _, v := range users {
		schoolAccount := &entities_bob.SchoolAdmin{}
		database.AllNullEntity(schoolAccount)
		err := multierr.Combine(
			schoolAccount.SchoolAdminID.Set(v.ID.String),
			schoolAccount.SchoolID.Set(schoolID),
			schoolAccount.ResourcePath.Set(fmt.Sprint(schoolID)),
		)

		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}

		schoolAdminEn = append(schoolAdminEn, schoolAccount)
	}

	err := s.SchoolAdminRepo.CreateMultiple(ctx, tx, schoolAdminEn)
	if err != nil {
		return errors.Wrapf(err, "s.SchoolAdminRepo.CreateMultiple")
	}
	return nil
}

func (s *UserModifierService) CreateParents(ctx context.Context, tx database.QueryExecer, schoolID int64, users []*entities_bob.User) error {
	parents := []*entities_bob.Parent{}

	for _, v := range users {
		parent := &entities_bob.Parent{}
		database.AllNullEntity(parent)
		err := multierr.Combine(
			parent.ID.Set(v.ID.String),
			parent.SchoolID.Set(schoolID),
			parent.ResourcePath.Set(fmt.Sprint(schoolID)),
		)

		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}

		parents = append(parents, parent)
	}

	err := s.ParentRepo.CreateMultiple(ctx, tx, parents)
	if err != nil {
		return fmt.Errorf("s.ParentRepo.CreateMultiple: %w", err)
	}

	return nil
}

func (s *UserModifierService) HandleCreateUser(ctx context.Context, userProfiles []*pb.CreateUserProfile, userGroup, organization string, schoolID int64) ([]*entities_bob.User, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	if userGroup == "" {
		return nil, status.Error(codes.InvalidArgument, "user group cannot be empty")
	}

	if len(userProfiles) == 0 {
		return nil, status.Error(codes.InvalidArgument, "users cannot be empty")
	}

	if err := s.CheckPermissionToCreateUser(ctx, userGroup, schoolID); err != nil {
		return nil, err
	}

	phones := []string{}
	emails := []string{}

	for _, v := range userProfiles {
		if v.Country == cpb.Country_COUNTRY_NONE {
			return nil, status.Error(codes.InvalidArgument, "country cannot be empty")
		}

		if v.PhoneNumber == "" {
			return nil, status.Error(codes.InvalidArgument, "phone number cannot be empty")
		}

		phones = append(phones, v.PhoneNumber)
		if v.Email == "" {
			return nil, status.Error(codes.InvalidArgument, "email cannot be empty")
		}

		user, err := s.FirebaseClient.GetUserByEmail(ctx, v.Email)
		if err != nil {
			switch {
			case auth.IsUserNotFound(err):
				emails = append(emails, v.Email)
			// Extend more cases
			default:
				return nil, errors.Wrapf(err, "FirebaseClient.GetUserByEmail")
			}
		}

		if user != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%s is exist in Firebase", v.Email))
		}
	}

	usersInDB, err := s.UserRepo.GetByEmail(ctx, s.DB, database.TextArray(emails))
	if err != nil {
		return nil, errors.Wrapf(err, "s.UserRepo.GetByEmail")
	}

	if len(usersInDB) != 0 {
		emails = []string{}
		for _, v := range usersInDB {
			emails = append(emails, v.Email.String)
		}

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can not create user with emails exist in db %s", strings.Join(emails, ", ")))
	}

	usersInDB, err = s.UserRepo.GetByPhone(ctx, s.DB, database.TextArray(phones))
	if err != nil {
		return nil, errors.Wrapf(err, "s.UserRepo.GetByPhone")
	}

	if len(usersInDB) != 0 {
		phones = []string{}
		for _, v := range usersInDB {
			phones = append(phones, v.PhoneNumber.String)
		}

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can not create user with phone number exist in db %s", strings.Join(phones, ", ")))
	}

	users, groups, err := UserPbToUserEn(ctx, userProfiles, userGroup, organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("err set entities: %w", err).Error())
	}

	// Guarantee each user email has only corresponding one uid
	for _, user := range users {
		createdUsrEmail, err := s.UsrEmailRepo.Create(ctx, s.DB, user.ID, user.Email)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		user.ID = createdUsrEmail.UsrID
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if userGroup == constant.UserGroupStudent {
			errCreateStudent := s.CreateStudents(ctx, tx, schoolID, userProfiles, users)
			if errCreateStudent != nil {
				return errCreateStudent
			}
		} else {
			err = s.UserRepo.CreateMultiple(ctx, tx, users)
			if err != nil {
				return errors.Wrapf(err, "s.UserRepo.CreateMultiple")
			}

			err = s.UserGroupRepo.CreateMultiple(ctx, tx, groups)
			if err != nil {
				return errors.Wrapf(err, "s.UserGroupRepo.CreateMultiple")
			}

			if userGroup == constant.UserGroupTeacher {
				errCreateTeacher := s.CreateTeachers(ctx, tx, schoolID, users)
				if errCreateTeacher != nil {
					return errCreateTeacher
				}
			}

			if IsSchoolAdmin(userGroup) {
				errCreateSchoolAdmin := s.CreateSchoolAdmins(ctx, tx, schoolID, organization, users)
				if errCreateSchoolAdmin != nil {
					return errCreateSchoolAdmin
				}
			}

			if userGroup == constant.UserGroupParent {
				errCreateParent := s.CreateParents(ctx, tx, schoolID, users)
				if errCreateParent != nil {
					return errCreateParent
				}
			}
		}

		err = s.CreateUserInFirebase(ctx, users, schoolID)
		if err != nil {
			return err
		}

		// Import to identity platform
		tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, tx, strconv.FormatInt(schoolID, 10))
		if err != nil {
			zapLogger.Warnw(
				"cannot get tenant id",
				"orgID", schoolID,
				"err", err.Error(),
			)
			// return status.Error(codes.FailedPrecondition, errors.New("tenant does not exists").Error())
		} else {
			err = s.CreateUserInIdentityPlatform(ctx, tenantID, users, schoolID)
			if err != nil {
				zapLogger.Warnw("failed to create users in identity platform")
				// ignore err for now
				// return status.Error(codes.Internal, err.Error())
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}
func (s *UserModifierService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	userGroup := req.UserGroup.String()

	if req.Organization == "" && IsSchoolAdmin(userGroup) {
		return nil, status.Error(codes.InvalidArgument, "organization cannot be empty")
	}

	resourcePath, err := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	users, err := s.HandleCreateUser(ctx, req.Users, userGroup, fmt.Sprint(resourcePath), int64(resourcePath))
	if err != nil {
		return nil, err
	}
	profiles := make([]*cpb.BasicProfile, 0, len(users))
	for _, v := range users {
		profile := &cpb.BasicProfile{
			UserId:      v.ID.String,
			GivenName:   v.GivenName.String,
			Name:        v.LastName.String,
			Avatar:      v.Avatar.String,
			Group:       cpb.UserGroup(cpb.UserGroup_value[v.Group.String]),
			FacebookId:  v.FacebookID.String,
			AppleUserId: v.AppleUser.ID.String,
		}

		profiles = append(profiles, profile)
	}

	return &pb.CreateUserResponse{
		Users: profiles,
	}, nil
}

func (s *UserModifierService) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid params")
	}

	profile, err := s.UserRepo.Get(ctx, s.DB, database.Text(req.Id))
	if err != nil {
		return nil, fmt.Errorf("s.UserRepo.Get: %w", err)
	}

	err = multierr.Combine(
		profile.LastName.Set(req.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	userGroup, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(req.Id))
	if err != nil {
		return nil, fmt.Errorf("s.UserRepo.UserGroup: %w", err)
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		switch userGroup {
		case entities_bob.UserGroupStudent:
			student, err := s.StudentRepo.Find(ctx, s.DB, database.Text(req.Id))
			if err != nil {
				return fmt.Errorf("s.StudentRepo.Find: %w", err)
			}

			err = multierr.Combine(
				student.CurrentGrade.Set(req.Grade),
			)
			if err != nil {
				return fmt.Errorf("multierr.Combine: %w", err)
			}

			err = s.StudentRepo.Update(ctx, s.DB, student)
			if err != nil {
				return fmt.Errorf("s.StudentRepo.Update: %w", err)
			}
		}

		err = s.UserRepo.UpdateProfile(ctx, s.DB, profile)
		if err != nil {
			return fmt.Errorf("s.UserRepo.UpdateProfile: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserProfileResponse{}, nil
}

func toStudentParentEn(studentParents []*pb.AssignToParentRequest_AssignParent) ([]*entities_bob.StudentParent, error) {
	s := make([]*entities.StudentParent, 0, len(studentParents))
	for _, studentParent := range studentParents {
		e := &entities.StudentParent{}
		now := time.Now()
		database.AllNullEntity(e)
		err := multierr.Combine(
			e.StudentID.Set(studentParent.StudentId),
			e.ParentID.Set(studentParent.ParentId),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
			e.Relationship.Set(studentParent.Relationship.String()),
		)
		if err != nil {
			return nil, fmt.Errorf("error convert pb to entities for student parent: %w", err)
		}
		s = append(s, e)
	}
	return s, nil
}

func (s *UserModifierService) AssignToParent(ctx context.Context, req *pb.AssignToParentRequest) (*pb.AssignToParentResponse, error) {
	studentParents, err := toStudentParentEn(req.AssignParents)
	if err != nil {
		return nil, err
	}
	studentIDs := make([]string, len(req.AssignParents))
	for i, assignParent := range req.AssignParents {
		studentIDs[i] = assignParent.StudentId
	}

	students, err := s.StudentRepo.FindStudentProfilesByIDs(ctx, s.DB, database.TextArray(studentIDs))
	if err != nil {
		return nil, err
	}

	studentMap := make(map[string]*entities.Student, len(students))
	for _, student := range students {
		studentMap[student.ID.String] = student
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = s.StudentParentRepo.Upsert(ctx, tx, studentParents)
		if err != nil {
			return fmt.Errorf("StudentParentRepo.Upsert: %w", err)
		}

		for _, assignParent := range req.AssignParents {
			studentProfiles := studentMap[assignParent.StudentId]
			data := &ppb_v1.EvtUser{
				Message: &ppb_v1.EvtUser_CreateParent_{
					CreateParent: &ppb_v1.EvtUser_CreateParent{
						StudentId:   assignParent.StudentId,
						ParentId:    assignParent.ParentId,
						StudentName: studentProfiles.GetName(),
						SchoolId:    strconv.Itoa(int(studentProfiles.SchoolID.Int)),
					},
				},
			}
			_ = s.publicAsyncUserEvent(ctx, data)
		}
		return nil
	})

	return &pb.AssignToParentResponse{
		Successful: true,
	}, nil
}

func overrideUserPassword(ctx context.Context, authClient firebase.AuthClient, userId string, password string) error {
	_, err := authClient.GetUser(ctx, userId)
	if err != nil {
		return err
	}

	userToUpdate := (&auth.UserToUpdate{}).Password(password)

	_, err = authClient.UpdateUser(ctx, userId, userToUpdate)
	if err != nil {
		return errors.Wrap(err, "overrideUserPassword()")
	}
	return nil
}

func updateUserEmailInFirebase(ctx context.Context, authClient firebase.AuthClient, userID string, email string) error {
	_, err := authClient.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	userToUpdate := (&auth.UserToUpdate{}).Email(email)

	_, err = authClient.UpdateUser(ctx, userID, userToUpdate)
	if err != nil {
		return errors.Wrapf(err, "overrideUserEmail userID=%v, email=%v", userID, email)
	}
	return nil
}

func newCreateStudentEvents(schoolId int32, students ...*entities_bob.Student) []*ppb_v1.EvtUser {
	createStudentEvents := make([]*ppb_v1.EvtUser, 0, len(students))
	for _, student := range students {
		createStudentEvent := &ppb_v1.EvtUser{
			Message: &ppb_v1.EvtUser_CreateStudent_{
				CreateStudent: &ppb_v1.EvtUser_CreateStudent{
					StudentId:   student.ID.String,
					StudentName: student.GetName(),
					SchoolId:    strconv.FormatInt(int64(schoolId), 10),
				},
			},
		}
		createStudentEvents = append(createStudentEvents, createStudentEvent)
	}
	return createStudentEvents
}

func newCreateParentEvents(schoolId int32, student *entities_bob.Student, parents ...*entities_bob.Parent) []*ppb_v1.EvtUser {
	createParentEvents := make([]*ppb_v1.EvtUser, 0, len(parents))
	for _, parent := range parents {
		createParentEvent := &ppb_v1.EvtUser{
			Message: &ppb_v1.EvtUser_CreateParent_{
				CreateParent: &ppb_v1.EvtUser_CreateParent{
					StudentId:   student.ID.String,
					ParentId:    parent.ID.String,
					StudentName: student.GetName(),
					SchoolId:    strconv.FormatInt(int64(schoolId), 10),
				},
			},
		}
		createParentEvents = append(createParentEvents, createParentEvent)
	}
	return createParentEvents
}

func (s *UserModifierService) publicAsyncUserEvent(ctx context.Context, userEvents ...*ppb_v1.EvtUser) error {
	for idx := range userEvents {
		data, err := proto.Marshal(userEvents[idx])
		if err != nil {
			return fmt.Errorf("proto.Marshal: %w", err)
		}
		msgID, err := s.JSM.TracedPublishAsync(ctx, "nats.TracedPublishAsync", constants.SubjectUserCreated, data)
		if err != nil {
			return fmt.Errorf("s.JSM.TracedPublishAsync: publish msg %s error, %w", msgID, err)
		}
	}
	return nil
}

func studentPbToStudentEntity(schoolId int32, req *pb.CreateStudentRequest) (*entities_bob.Student, error) {
	studentEnt := &entities_bob.Student{}
	database.AllNullEntity(studentEnt)
	database.AllNullEntity(&studentEnt.User)
	studentId := idutil.ULIDNow()
	err := multierr.Combine(
		studentEnt.User.ID.Set(studentId),
		studentEnt.User.Email.Set(req.StudentProfile.Email),
		studentEnt.User.PhoneNumber.Set(req.StudentProfile.PhoneNumber),
		studentEnt.User.LastName.Set(req.StudentProfile.Name),
		studentEnt.User.Country.Set(req.StudentProfile.CountryCode.String()),
		studentEnt.ID.Set(studentId),
		studentEnt.SchoolID.Set(schoolId),
		studentEnt.ResourcePath.Set(fmt.Sprint(schoolId)),
		studentEnt.CurrentGrade.Set(req.StudentProfile.Grade),
		studentEnt.EnrollmentStatus.Set(req.StudentProfile.EnrollmentStatus.String()),
		studentEnt.StudentNote.Set(req.StudentProfile.StudentNote),
	)
	if req.StudentProfile.PhoneNumber == "" {
		if err := studentEnt.PhoneNumber.Set(nil); err != nil {
			return nil, err
		}
	}
	if req.StudentProfile.StudentExternalId != "" {
		if err := studentEnt.StudentExternalID.Set(req.StudentProfile.StudentExternalId); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	studentEnt.UserAdditionalInfo.Password = req.StudentProfile.Password

	return studentEnt, nil
}

func parentPbToParentEntity(schoolId int32, parentPb *pb.CreateStudentRequest_ParentProfile, parentId string) (*entities_bob.Parent, error) {
	parentEnt := &entities_bob.Parent{
		ParentAdditionalInfo: &entities_bob.ParentAdditionalInfo{},
	}
	database.AllNullEntity(parentEnt)
	database.AllNullEntity(&parentEnt.User)
	err := multierr.Combine(
		parentEnt.User.ID.Set(parentId),
		parentEnt.User.Email.Set(parentPb.Email),
		parentEnt.User.LastName.Set(parentPb.Name),
		parentEnt.User.Group.Set(entities_bob.UserGroupParent),
		parentEnt.User.Country.Set(parentPb.CountryCode.String()),
		parentEnt.User.PhoneNumber.Set(parentPb.PhoneNumber),
		parentEnt.ID.Set(parentId),
		parentEnt.SchoolID.Set(schoolId),
	)
	if parentPb.PhoneNumber == "" {
		if err := parentEnt.PhoneNumber.Set(nil); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	parentEnt.User.UserAdditionalInfo.Password = parentPb.Password
	parentEnt.ParentAdditionalInfo.Relationship = parentPb.Relationship.String()

	return parentEnt, nil
}

func validCreateRequest(req *pb.CreateStudentRequest) error {
	req.StudentProfile.Email = strings.TrimSpace(req.StudentProfile.Email)
	req.StudentProfile.PhoneNumber = strings.TrimSpace(req.StudentProfile.PhoneNumber)

	switch {
	case req.StudentProfile.Email == "":
		return errors.New("student email cannot be empty")
	case req.StudentProfile.Password == "":
		return errors.New("student password cannot be empty")
	case cpb.Country_name[int32(req.StudentProfile.CountryCode.Enum().Number())] == "":
		return errors.New("student country code is not valid")
	case len(req.StudentProfile.Password) < firebase.MinimumPasswordLength:
		return errors.New("student password length should be at least 6")
	case req.StudentProfile.Name == "":
		return errors.New("student name cannot be empty")
	}

	if _, ok := cpb.StudentEnrollmentStatus_value[req.StudentProfile.EnrollmentStatus.String()]; !ok {
		return ErrStudentEnrollmentStatusUnknown
	}
	if req.StudentProfile.EnrollmentStatus == cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
		return ErrStudentEnrollmentStatusNotAllowedTobeNone
	}

	for _, parentProfile := range req.ParentProfiles {
		parentProfile.PhoneNumber = strings.TrimSpace(parentProfile.PhoneNumber)
		parentProfile.Email = strings.TrimSpace(parentProfile.Email)

		if pb.FamilyRelationship_name[int32(parentProfile.Relationship.Enum().Number())] == "" {
			return errors.New("parent relationship is not valid")
		}
		if parentProfile.Id == "" {
			switch {
			case parentProfile.Email == "":
				return errors.New("parent email cannot be empty")
			case parentProfile.Name == "":
				return errors.New("parent name cannot be empty")
			case cpb.Country_name[int32(parentProfile.CountryCode.Enum().Number())] == "":
				return errors.New("parent country code is not valid")
			case parentProfile.Password == "":
				return errors.New("parent password cannot be empty")
			case len(parentProfile.Password) < firebase.MinimumPasswordLength:
				return errors.New("parent password length should be at least 6")
			}
		}
	}

	for _, profile := range req.StudentPackageProfiles {
		if profile.Start.AsTime().After(profile.End.AsTime()) {
			return fmt.Errorf("UserModifier.validCreateRequest: package profile start date must before end date")
		}
	}

	return nil
}

func classifyParentProfile(schoolId int32, req *pb.CreateStudentRequest) (entities_bob.Parents, entities_bob.Parents, error) {
	toCreateParents := make(entities_bob.Parents, 0)
	toAssignParents := make(entities_bob.Parents, 0)

	for _, parentProfile := range req.ParentProfiles {
		switch parentProfile.Id {
		case "":
			parentEnt, err := parentPbToParentEntity(schoolId, parentProfile, idutil.ULIDNow())
			if err != nil {
				return nil, nil, err
			}

			toCreateParents = append(toCreateParents, parentEnt)
		default:
			parentEnt, err := parentPbToParentEntity(schoolId, parentProfile, parentProfile.Id)
			if err != nil {
				return nil, nil, err
			}
			toAssignParents = append(toAssignParents, parentEnt)
		}
	}
	return toCreateParents, toAssignParents, nil
}

func (s *UserModifierService) createParents(ctx context.Context, db database.QueryExecer, parents entities_bob.Parents) error {
	// UserGroups need to be inserted with new parent
	userGroups := make([]*entities_bob.UserGroup, 0, len(parents))
	for _, parent := range parents {
		userGroup := &entities_bob.UserGroup{}
		database.AllNullEntity(userGroup)
		err := multierr.Combine(
			userGroup.UserID.Set(parent.ID.String),
			userGroup.GroupID.Set(entities_bob.UserGroupParent),
			userGroup.IsOrigin.Set(true),
			userGroup.Status.Set(entities_bob.UserGroupStatusActive),
		)
		if err != nil {
			return err
		}
		userGroups = append(userGroups, userGroup)
	}

	// Insert new parents
	err := s.UserRepo.CreateMultiple(ctx, db, parents.Users())
	if err != nil {
		return errorx.ToStatusError(err)
	}
	err = s.ParentRepo.CreateMultiple(ctx, db, parents)
	if err != nil {
		return errorx.ToStatusError(err)
	}
	err = s.UserGroupRepo.CreateMultiple(ctx, db, userGroups)
	if err != nil {
		return errorx.ToStatusError(err)
	}

	return nil
}

func (s *UserModifierService) assignParentsToStudent(ctx context.Context, db database.QueryExecer, student *entities.Student, parents ...*entities_bob.Parent) error {
	studentParentEntities := make([]*entities_bob.StudentParent, 0, len(parents))

	if len(parents) == 0 {
		studentParent := &entities.StudentParent{}
		database.AllNullEntity(studentParent)
		err := multierr.Combine(
			studentParent.StudentID.Set(student.ID),
		)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
		}
		studentParentEntities = append(studentParentEntities, studentParent)
	} else {
		for _, parent := range parents {
			studentParent := &entities.StudentParent{}
			database.AllNullEntity(studentParent)
			err := multierr.Combine(
				studentParent.StudentID.Set(student.ID),
				studentParent.ParentID.Set(parent.ID),
				studentParent.Relationship.Set(parent.ParentAdditionalInfo.Relationship),
			)
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
			}
			studentParentEntities = append(studentParentEntities, studentParent)
		}
	}

	// Insert relationship between student and parent
	err := s.StudentParentRepo.Upsert(ctx, db, studentParentEntities)
	if err != nil {
		return errorx.ToStatusError(err)
	}

	return nil
}

func parentsToParentPBsInCreateStudentResponse(parents ...*entities_bob.Parent) []*pb.CreateStudentResponse_ParentProfile {
	// Parent profile pb to response
	parentProfilePBs := make([]*pb.CreateStudentResponse_ParentProfile, 0, len(parents))
	for _, parent := range parents {
		parentProfilePB := &pb.CreateStudentResponse_ParentProfile{
			Parent: &pb.Parent{
				UserProfile: &pb.UserProfile{
					UserId:      parent.User.ID.String,
					Email:       parent.User.Email.String,
					Name:        parent.User.GetName(),
					Avatar:      parent.User.Avatar.String,
					Group:       cpb.UserGroup(cpb.UserGroup_value[parent.User.Group.String]),
					PhoneNumber: parent.User.PhoneNumber.String,
					FacebookId:  parent.User.FacebookID.String,
					AppleUserId: parent.User.AppleUser.UserID.String,
					GivenName:   parent.User.GivenName.String,
					CountryCode: cpb.Country(cpb.Country_value[parent.User.Country.String]),
				},
				SchoolId: parent.SchoolID.Int,
			},
			ParentPassword: parent.User.UserAdditionalInfo.Password,
			Relationship:   pb.FamilyRelationship(pb.FamilyRelationship_value[parent.ParentAdditionalInfo.Relationship]),
		}
		parentProfilePBs = append(parentProfilePBs, parentProfilePB)
	}
	return parentProfilePBs
}

func studentToStudentPBInCreateStudentResponse(student *entities_bob.Student) *pb.Student {
	studentPB := &pb.Student{
		UserProfile: &pb.UserProfile{
			UserId:      student.ID.String,
			Email:       student.Email.String,
			Name:        student.GetName(),
			Avatar:      student.Avatar.String,
			Group:       cpb.UserGroup(cpb.UserGroup_value[student.Group.String]),
			PhoneNumber: student.PhoneNumber.String,
			CountryCode: cpb.Country(cpb.Country_value[student.User.Country.String]),
		},
		Grade:    int32(student.CurrentGrade.Int),
		SchoolId: student.SchoolID.Int,
	}
	return studentPB
}

func (s *UserModifierService) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentResponse, error) {
	if err := validCreateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resourcePath, err := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	var studentPB *pb.Student
	var parentProfilePBs []*pb.CreateStudentResponse_ParentProfile
	var studentPackageProfilePBs []*pb.CreateStudentResponse_StudentPackageProfile

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Convert student data in request to entities
		student, err := studentPbToStudentEntity(int32(resourcePath), req)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Valid email and phone number in student data
		existingUsers, err := s.UserRepo.GetByEmail(ctx, tx, database.TextArray([]string{student.Email.String}))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", err).Error())
		}
		if len(existingUsers) > 0 {
			return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create student with emails existing in system: %s", student.Email.String))
		}
		if student.PhoneNumber.String != "" {
			phoneNumberExistingStudents, err := s.UserRepo.GetByPhone(ctx, tx, database.TextArray([]string{student.PhoneNumber.String}))
			if err != nil {
				return fmt.Errorf("s.UserRepo.GetByPhone: %w", err)
			}
			if len(phoneNumberExistingStudents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create student with phone number existing in system: %s", student.PhoneNumber.String))
			}
		}

		// Insert new student
		if err := s.StudentRepo.Create(ctx, tx, student); err != nil {
			return errorx.ToStatusError(err)
		}
		studentPB = studentToStudentPBInCreateStudentResponse(student)

		// Classify parent data into parent data to create and parent data to assign
		toCreateParents, toAssignParents, err := classifyParentProfile(int32(resourcePath), req)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		allParents := make(entities_bob.Parents, 0, len(toCreateParents)+len(toAssignParents))

		// Resolve parent data will be create
		if toCreateParents.Len() > 0 {
			// Valid emails for new parents to create
			emailExistingParents, err := s.UserRepo.GetByEmail(ctx, tx, database.TextArray(toCreateParents.Emails()))
			if err != nil {
				return errorx.ToStatusError(err)
			}
			if len(emailExistingParents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with emails existing in system: %s", strings.Join(entities_bob.ToUsers(emailExistingParents...).Emails(), ", ")))
			}
			// Valid phone numbers for new parents to create
			phoneNumberExistingParents, err := s.UserRepo.GetByPhone(ctx, tx, database.TextArray(toCreateParents.PhoneNumbers()))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByPhone: %w", err).Error())
			}
			if len(phoneNumberExistingParents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with phone number existing in system: %s", strings.Join(entities_bob.ToUsers(phoneNumberExistingParents...).PhoneNumbers(), ", ")))
			}

			// Create multiple parents
			err = s.createParents(ctx, tx, toCreateParents)
			if err != nil {
				return errorx.ToStatusError(err)
			}

			allParents = append(allParents, toCreateParents...)
		}

		// Resolve parent data will be assign to student
		if toAssignParents.Len() > 0 {
			// Valid parent will be assigned to student
			existingParents, err := s.ParentRepo.GetByIds(ctx, tx, database.TextArray(toAssignParents.Ids()))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByIds: %w", err).Error())
			}
			// Convert existing parents slice to map with id as key
			existingParentsMap := make(map[string]*entities_bob.Parent, len(existingParents))
			for _, existingParent := range existingParents {
				existingParentsMap[existingParent.ID.String] = existingParent
			}
			// Check parents will be assign to student are exists or not
			for _, toAssignParent := range toAssignParents {
				existingParent, ok := existingParentsMap[toAssignParent.ID.String]
				if !ok {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot assign non-existing parent to student: %s", toAssignParent.ID.String))
				}
				existingParent.ParentAdditionalInfo = toAssignParent.ParentAdditionalInfo
				toAssignParent = existingParent
			}

			allParents = append(allParents, existingParents...)
		}

		if allParents.Len() > 0 {
			// Upsert ignore conflict for relationships between student and parent
			err = s.assignParentsToStudent(ctx, tx, student, allParents...)
			if err != nil {
				return errorx.ToStatusError(err)
			}

			// Protobuf responses
			parentProfilePBs = append(parentProfilePBs, parentsToParentPBsInCreateStudentResponse(allParents...)...)
		}

		// Create firebase accounts for student and parents
		firebaseAccounts := toCreateParents.Users().Append(&student.User)

		err = s.CreateUserInFirebase(ctx, firebaseAccounts, int64(resourcePath))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.CreateUserInFirebase: %w", err).Error())
		}
		for _, firebaseAccount := range firebaseAccounts {
			err := overrideUserPassword(ctx, s.FirebaseClient, firebaseAccount.ID.String, firebaseAccount.UserAdditionalInfo.Password)
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("overrideUserPassword: %w", err).Error())
			}
		}

		// Add student packages by calling fatima service
		for _, studentPackageProfile := range req.StudentPackageProfiles {
			startAt := studentPackageProfile.Start
			endAt := studentPackageProfile.End

			addStudentPackageCourseReq := &fpb.AddStudentPackageCourseRequest{
				StudentId: student.ID.String,
				CourseIds: []string{studentPackageProfile.CourseId},
				StartAt:   startAt,
				EndAt:     endAt,
			}

			resp, err := s.FatimaClient.AddStudentPackageCourse(signCtx(ctx), addStudentPackageCourseReq)
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.FatimaClient.AddStudentPackageCourse: %w", err).Error())
			}

			studentCourse := &pb.CreateStudentResponse_StudentPackageProfile{
				StudentPackageId: resp.StudentPackageId,
				Start:            startAt,
				End:              endAt,
			}
			studentPackageProfilePBs = append(studentPackageProfilePBs, studentCourse)
		}

		// Firebase accounts and nat streaming events
		userEvents := make([]*ppb_v1.EvtUser, 0, len(allParents)+1)
		userEvents = append(userEvents, newCreateStudentEvents(int32(resourcePath), student)...)
		userEvents = append(userEvents, newCreateParentEvents(int32(resourcePath), student, allParents...)...)

		_ = s.publicAsyncUserEvent(ctx, userEvents...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	response := &pb.CreateStudentResponse{
		StudentProfile: &pb.CreateStudentResponse_StudentProfile{
			Student:         studentPB,
			StudentPassword: req.StudentProfile.Password,
		},
		ParentProfiles:         parentProfilePBs,
		StudentPackageProfiles: studentPackageProfilePBs,
	}

	return response, nil
}

func validUpdateStudentRequest(req *pb.UpdateStudentRequest) error {
	switch {
	case req.StudentProfile.Id == "":
		return errors.New("student id is not valid")
	case req.StudentProfile.Name == "":
		return errors.New("student name is not valid")
	case req.StudentProfile.Email == "":
		return errors.New("student email cannot be empty")
	}
	for _, parentProfile := range req.ParentProfiles {
		if req.StudentProfile.Email == parentProfile.Email {
			return errors.New("student email cannot be the same as parent email")
		}
	}

	if _, ok := cpb.StudentEnrollmentStatus_value[req.StudentProfile.EnrollmentStatus.String()]; !ok {
		return ErrStudentEnrollmentStatusUnknown
	}
	if req.StudentProfile.EnrollmentStatus == cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
		return ErrStudentEnrollmentStatusNotAllowedTobeNone
	}

	for _, profile := range req.StudentPackageProfiles {
		if profile.StartTime.AsTime().After(profile.EndTime.AsTime()) {
			return fmt.Errorf("UserModifier.validUpdateStudentRequest: start time %s could not after end time %s", profile.StartTime.AsTime(), profile.EndTime.AsTime())
		}
	}

	return nil
}

func studentPbInUpdateStudentRequestToStudentEnt(req *pb.UpdateStudentRequest) (*entities_bob.Student, error) {
	student := new(entities_bob.Student)
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	err := multierr.Combine(
		student.ID.Set(req.StudentProfile.Id),
		student.LastName.Set(req.StudentProfile.Name),
		student.CurrentGrade.Set(req.StudentProfile.Grade),
		student.EnrollmentStatus.Set(req.StudentProfile.EnrollmentStatus.String()),
		student.StudentNote.Set(req.StudentProfile.StudentNote),
		student.Email.Set(req.StudentProfile.Email),
	)
	if req.StudentProfile.StudentExternalId != "" {
		if err := student.StudentExternalID.Set(req.StudentProfile.StudentExternalId); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return student, nil
}

func parentPBInUpdateStudentRequestToParentEnt(req *pb.UpdateStudentRequest, schoolId int32) (entities_bob.Parents, entities_bob.Parents, error) {
	toCreateParents := make(entities_bob.Parents, 0)
	toAssignParents := make(entities_bob.Parents, 0)

	for _, parentProfile := range req.ParentProfiles {
		parentProfile.Email = strings.TrimSpace(parentProfile.Email)
		parentProfile.PhoneNumber = strings.TrimSpace(parentProfile.PhoneNumber)

		parentEnt := &entities_bob.Parent{
			ParentAdditionalInfo: &entities_bob.ParentAdditionalInfo{},
		}
		database.AllNullEntity(parentEnt)
		database.AllNullEntity(&parentEnt.User)

		var parentId string
		if parentProfile.Id == "" {
			parentId = idutil.ULIDNow()
			toCreateParents = append(toCreateParents, parentEnt)
		} else {
			parentId = parentProfile.Id
			toAssignParents = append(toAssignParents, parentEnt)
		}

		err := multierr.Combine(
			parentEnt.User.ID.Set(parentId),
			parentEnt.User.Email.Set(parentProfile.Email),
			parentEnt.User.LastName.Set(parentProfile.Name),
			parentEnt.User.Group.Set(entities_bob.UserGroupParent),
			parentEnt.User.Country.Set(cpb.Country_name[int32(parentProfile.CountryCode.Enum().Number())]),
			parentEnt.User.PhoneNumber.Set(parentProfile.PhoneNumber),
			parentEnt.ID.Set(parentId),
			parentEnt.SchoolID.Set(schoolId),
		)
		if parentProfile.PhoneNumber == "" {
			if err := parentEnt.PhoneNumber.Set(nil); err != nil {
				return nil, nil, err
			}
		}
		if err != nil {
			return nil, nil, err
		}
		parentEnt.User.UserAdditionalInfo.Password = parentProfile.Password
		parentEnt.ParentAdditionalInfo.Relationship = pb.FamilyRelationship_name[int32(parentProfile.Relationship.Enum().Number())]
	}
	return toCreateParents, toAssignParents, nil
}

func (s *UserModifierService) validToCreateParents(parents entities_bob.Parents) error {
	for _, parent := range parents {
		switch {
		case parent.Email.String == "":
			return errors.New("parent email cannot be empty")
		case parent.GetName() == "":
			return errors.New("parent name cannot be empty")
		case parent.Country.String == "":
			return errors.New("parent country code is not valid")
		case parent.ParentAdditionalInfo.Relationship == "":
			return errors.New("parent relationship is not valid")
		case len(parent.User.UserAdditionalInfo.Password) < firebase.MinimumPasswordLength:
			return errors.New("parent password length should be at least 6")
		}
	}
	return nil
}

func (s *UserModifierService) validToAssignParents(parents entities_bob.Parents) error {
	for _, parent := range parents {
		switch {
		case parent.ID.String == "":
			return errors.New("parent id cannot be empty")
		case parent.ParentAdditionalInfo.Relationship == "":
			return errors.New("parent relationship is not valid")
		case parent.Email.String == "":
			return errors.New("parent email cannot be empty")
		}
	}
	return nil
}

func (s *UserModifierService) updateStudentPackageProfiles(ctx context.Context, studentID string, profiles []*pb.UpdateStudentRequest_StudentPackageProfile) ([]*pb.UpdateStudentResponse_StudentPackageProfile, error) {
	var res = make([]*pb.UpdateStudentResponse_StudentPackageProfile, 0, len(profiles))
	addItems := make([]*fpb.AddStudentPackageCourseRequest, 0)
	editItems := make([]*fpb.EditTimeStudentPackageRequest, 0)

	for _, profile := range profiles {
		startAt := profile.StartTime
		endAt := profile.EndTime

		switch id := profile.Id.(type) {
		case *pb.UpdateStudentRequest_StudentPackageProfile_CourseId:
			addItem := &fpb.AddStudentPackageCourseRequest{
				StudentId: studentID,
				CourseIds: []string{id.CourseId},
				StartAt:   startAt,
				EndAt:     endAt,
			}

			addItems = append(addItems, addItem)
		case *pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId:
			editItem := &fpb.EditTimeStudentPackageRequest{
				StudentPackageId: id.StudentPackageId,
				StartAt:          startAt,
				EndAt:            endAt,
			}

			editItems = append(editItems, editItem)
		default:
			return nil, fmt.Errorf("UpdateStudentPackageProfiles: not yet implement for type %T", profile.Id)
		}
	}

	// call fatima service to add package
	for _, addItem := range addItems {
		addResp, err := s.FatimaClient.AddStudentPackageCourse(signCtx(ctx), addItem)
		if err != nil {
			return nil, fmt.Errorf("s.FatimaClient.AddStudentPackageCourse: %w", err)
		}
		pgk := &pb.UpdateStudentResponse_StudentPackageProfile{
			StudentPackageId: addResp.StudentPackageId,
		}
		res = append(res, pgk)
	}

	// call fatima service to update package
	for _, editItem := range editItems {
		editResp, err := s.FatimaClient.EditTimeStudentPackage(signCtx(ctx), editItem)
		if err != nil {
			return nil, fmt.Errorf("s.FatimaClient.EditTimeStudentPackage: %w", err)
		}
		pgk := &pb.UpdateStudentResponse_StudentPackageProfile{
			StudentPackageId: editResp.StudentPackageId,
		}
		res = append(res, pgk)
	}

	return res, nil
}

func (s *UserModifierService) publishAsyncUserDeviceToken(ctx context.Context, userID string, name string) error {
	data := &ppb_v1.EvtUserInfo{
		UserId: userID,
		Name:   name,
	}
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}

	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectUserDeviceTokenUpdated, msg)
	if err != nil {
		ctxzap.Extract(ctx).Error("UpdateUserProfile s.BusFactory.PublishAsync failed", zap.String("msg-id", msgID), zap.Error(err))
	}
	return nil
}

func (s *UserModifierService) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentResponse, error) {
	if err := validUpdateStudentRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	toUpdateStudent, err := studentPbInUpdateStudentRequestToStudentEnt(req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// validate request
	// validate student
	existingStudent, err := s.StudentRepo.Find(ctx, s.DB, toUpdateStudent.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "student id is not exists")
	}

	existingUser, err := s.UserRepo.Get(ctx, s.DB, toUpdateStudent.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id is not exists")
	}

	// validate parent
	toCreateParents, toAssignParents, err := parentPBInUpdateStudentRequestToParentEnt(req, existingStudent.SchoolID.Int)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.validToCreateParents(toCreateParents)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.validToAssignParents(toAssignParents)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var studentProfile *pb.Student
	var parentProfilePBs []*pb.UpdateStudentResponse_ParentProfile
	var studentPackagePBs []*pb.UpdateStudentResponse_StudentPackageProfile
	var userEvents []*ppb_v1.EvtUser

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		existingStudent.User = *existingUser

		// if student email edited
		if toUpdateStudent.Email != existingStudent.Email {
			// Check if edited email already exists
			emailExistingStudents, err := s.UserRepo.GetByEmail(ctx, tx, database.TextArray([]string{toUpdateStudent.Email.String}))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", err).Error())
			}
			if len(emailExistingStudents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot edit student with email existing in system: %s", toUpdateStudent.Email.String))
			}
			existingStudent.User.Email = toUpdateStudent.User.Email

			// Update new email in firebase
			err = updateUserEmailInFirebase(ctx, s.FirebaseClient, existingStudent.ID.String, existingStudent.Email.String)
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("overrideUserEmail: %w", err).Error())
			}
		}

		existingStudent.User.LastName = toUpdateStudent.User.LastName
		existingStudent.CurrentGrade = toUpdateStudent.CurrentGrade
		existingStudent.EnrollmentStatus = toUpdateStudent.EnrollmentStatus
		existingStudent.StudentExternalID = toUpdateStudent.StudentExternalID
		existingStudent.StudentNote = toUpdateStudent.StudentNote

		err = s.StudentRepo.UpdateV2(ctx, tx, existingStudent)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.StudentRepo.Update: %w", err).Error())
		}

		if toCreateParents.Len() > 0 {
			// Valid emails for new parents to create
			emailExistingParents, err := s.UserRepo.GetByEmail(ctx, tx, database.TextArray(toCreateParents.Emails()))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", err).Error())
			}
			if len(emailExistingParents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with emails existing in system: %s", strings.Join(entities_bob.ToUsers(emailExistingParents...).Emails(), ", ")))
			}
			// Valid phone numbers for new parents to create
			phoneNumberExistingParents, err := s.UserRepo.GetByPhone(ctx, tx, database.TextArray(toCreateParents.PhoneNumbers()))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByPhone: %w", err).Error())
			}
			if len(phoneNumberExistingParents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with phone number existing in system: %s", strings.Join(entities_bob.ToUsers(phoneNumberExistingParents...).PhoneNumbers(), ", ")))
			}

			err = s.createParents(ctx, tx, toCreateParents)
			if err != nil {
				return errorx.ToStatusError(err)
			}

			err = s.CreateUserInFirebase(ctx, toCreateParents.Users(), int64(existingStudent.SchoolID.Int))
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			for _, toCreateUser := range toCreateParents.Users() {
				err := overrideUserPassword(ctx, s.FirebaseClient, toCreateUser.ID.String, toCreateUser.UserAdditionalInfo.Password)
				if err != nil {
					return status.Error(codes.Internal, fmt.Errorf("overrideUserPassword: %w", err).Error())
				}
			}
		}

		if toAssignParents.Len() > 0 {
			// Valid parent will be assigned to student
			existingParents, err := s.ParentRepo.GetByIds(ctx, tx, database.TextArray(toAssignParents.Ids()))
			if err != nil {
				return errorx.ToStatusError(errors.Wrap(err, "s.ParentRepo.GetByIds"))
			}
			// Convert existing parents slice to map with id as key
			existingParentsMap := make(map[string]*entities_bob.Parent, len(existingParents))
			for _, existingParent := range existingParents {
				existingParentsMap[existingParent.ID.String] = existingParent
			}
			// Check parents will be assign to student are exists or not
			for i, toAssignParent := range toAssignParents {
				existingParent := existingParentsMap[toAssignParent.ID.String]
				if existingParent == nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot assign non-existing parent to student: %s", toAssignParent.ID.String))
				}

				if existingParent.SchoolID.Int != existingStudent.SchoolID.Int {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("parent %s not same school with student", toAssignParent.ID.String))
				}

				// If parent email edited
				if existingParent.Email != toAssignParent.Email {
					// Check if edited email already exists
					emailExistingParents, err := s.UserRepo.GetByEmail(ctx, tx, database.TextArray(entities_bob.Parents{toAssignParent}.Emails()))
					if err != nil {
						return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", err).Error())
					}
					if len(emailExistingParents) > 0 {
						return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot edit parent with email existing in system: %s", toAssignParent.Email.String))
					}
					// Update new email in DB
					existingParent.Email = toAssignParent.Email
					err = s.UserRepo.UpdateEmail(ctx, tx, &entities_bob.User{ID: existingParent.ID, Email: existingParent.Email})
					if err != nil {
						return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.UpdateEmail: %w", err).Error())
					}
					// Update new email in firebase
					err = updateUserEmailInFirebase(ctx, s.FirebaseClient, existingParent.ID.String, existingParent.Email.String)
					if err != nil {
						return status.Error(codes.Internal, fmt.Errorf("overrideUserEmail: %w", err).Error())
					}
				}

				existingParent.ParentAdditionalInfo = toAssignParent.ParentAdditionalInfo
				toAssignParents[i] = existingParent
			}
		}

		allParents := make(entities_bob.Parents, 0, len(toCreateParents)+len(toAssignParents))
		allParents = append(allParents, toCreateParents...)
		allParents = append(allParents, toAssignParents...)

		parentsRemoveIds, err := s.listParentsRemovedFromStudent(ctx, tx, existingStudent, allParents...)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("removeParentRelationship: %w", err).Error())
		}
		// Nat streaming events
		userEvents = append(userEvents, newCreateParentEvents(existingStudent.SchoolID.Int, existingStudent, allParents...)...)
		userEvents = append(userEvents, newParentRemovedFromStudentEvents(existingStudent, parentsRemoveIds)...)

		err = s.assignParentsToStudent(ctx, tx, existingStudent, allParents...)
		if err != nil {
			return errorx.ToStatusError(errors.Wrap(err, "assignParentsToStudent"))
		}

		for _, parent := range allParents {
			parentProfilePB := &pb.UpdateStudentResponse_ParentProfile{
				Parent: &pb.Parent{
					UserProfile: &pb.UserProfile{
						UserId:      parent.User.ID.String,
						Email:       parent.User.Email.String,
						Name:        parent.User.GetName(),
						Avatar:      parent.User.Avatar.String,
						Group:       cpb.UserGroup(cpb.UserGroup_value[parent.User.Group.String]),
						PhoneNumber: parent.User.PhoneNumber.String,
						CountryCode: cpb.Country(cpb.Country_value[parent.User.Country.String]),
					},
					SchoolId: existingStudent.SchoolID.Int,
				},
				ParentPassword: parent.User.UserAdditionalInfo.Password,
				Relationship:   pb.FamilyRelationship(pb.FamilyRelationship_value[parent.ParentAdditionalInfo.Relationship]),
			}
			parentProfilePBs = append(parentProfilePBs, parentProfilePB)
		}

		studentProfile = &pb.Student{
			UserProfile: &pb.UserProfile{
				UserId:      existingStudent.ID.String,
				Email:       existingStudent.Email.String,
				Name:        existingStudent.GetName(),
				Avatar:      existingStudent.Avatar.String,
				Group:       cpb.UserGroup(cpb.UserGroup_value[existingStudent.Group.String]),
				PhoneNumber: existingStudent.PhoneNumber.String,
				CountryCode: cpb.Country(cpb.Country_value[existingStudent.User.Country.String]),
			},
			Grade:    int32(existingStudent.CurrentGrade.Int),
			SchoolId: existingStudent.SchoolID.Int,
		}

		// update student package profile
		studentPackagePBs, err = s.updateStudentPackageProfiles(ctx, existingStudent.User.ID.String, req.StudentPackageProfiles)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.publicAsyncUserEvent(ctx, userEvents...)
	_ = s.publishAsyncUserDeviceToken(ctx, existingStudent.ID.String, existingStudent.GetName())

	return &pb.UpdateStudentResponse{
		StudentProfile:         studentProfile,
		ParentProfiles:         parentProfilePBs,
		StudentPackageProfiles: studentPackagePBs,
	}, nil
}

func (s *UserModifierService) listParentsRemovedFromStudent(ctx context.Context, db database.QueryExecer, student *entities.Student, parents ...*entities_bob.Parent) ([]string, error) {
	parentsRemoveIds := []string{}
	studentParentsExists, err := s.StudentParentRepo.GetStudentParents(ctx, db, database.TextArray([]string{student.ID.String}))
	if err != nil {
		return nil, errorx.ToStatusError(err)
	}
	for _, studentParent := range studentParentsExists {
		if find := func(parents []*entities_bob.Parent, studentParent *entities.StudentParent) bool {
			for _, parent := range parents {
				if studentParent.ParentID == parent.ID {
					return true
				}
			}
			return false
		}(parents, studentParent); !find {
			parentsRemoveIds = append(parentsRemoveIds, studentParent.ParentID.String)
		}
	}

	return parentsRemoveIds, nil
}

func newParentRemovedFromStudentEvents(student *entities_bob.Student, removeParentIDs []string) []*ppb_v1.EvtUser {
	parentRemovedFromStudentEvents := make([]*ppb_v1.EvtUser, 0, len(removeParentIDs))
	for _, parentID := range removeParentIDs {
		parentRemovedFromStudentEvent := &ppb_v1.EvtUser{
			Message: &ppb_v1.EvtUser_ParentRemovedFromStudent_{
				ParentRemovedFromStudent: &ppb_v1.EvtUser_ParentRemovedFromStudent{
					StudentId: student.ID.String,
					ParentId:  parentID,
				},
			},
		}
		parentRemovedFromStudentEvents = append(parentRemovedFromStudentEvents, parentRemovedFromStudentEvent)
	}

	return parentRemovedFromStudentEvents
}

func signCtx(ctx context.Context) context.Context {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}
	return metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token)
}
