package services

import (
	"context"
	"fmt"
	"strconv"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	ppb "github.com/manabie-com/backend/pkg/genproto/bob"
	pby "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pby_v1 "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"firebase.google.com/go/v4/auth"
	"github.com/go-pg/pg"
	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	DBPgx database.Ext

	UserRepo interface {
		Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities_bob.User, error)
		Create(ctx context.Context, db database.QueryExecer, u *entities_bob.User) error
		CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entities_bob.User) error
		GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entities_bob.User, error)
		GetByPhone(ctx context.Context, db database.QueryExecer, phones pgtype.TextArray) ([]*entities_bob.User, error)
		UpdateProfile(ctx context.Context, db database.QueryExecer, u *entities_bob.User) error
		SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
		FindByIDUnscope(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities_bob.User, error)
		Update(ctx context.Context, db database.QueryExecer, s *entities_bob.User) error
	}
	TeacherRepo interface {
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error
		Update(ctx context.Context, db database.QueryExecer, s *entities_bob.Teacher) error
		Create(ctx context.Context, db database.QueryExecer, t *entities_bob.Teacher) error
		CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entities_bob.Teacher) error
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities_bob.Teacher, error)
		FindRegardlessDeletion(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities_bob.Teacher, error)
	}
	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities_bob.SchoolAdmin, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, schoolAdmins []*entities_bob.SchoolAdmin) error
	}
	SchoolRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities_bob.School, error)
	}
	FirebaseClient *auth.Client
	UserController ppb.UserServiceClient
	UserGroupRepo  interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entities_bob.UserGroup) error
	}
	StudentRepo interface {
		Create(context.Context, database.QueryExecer, *entities_bob.Student) error
		Update(ctx context.Context, db database.QueryExecer, s *entities_bob.Student) error
		Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*entities_bob.Student, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error
	}
	UserModifierService interface {
		HandleCreateUser(ctx context.Context, userProfiles []*pby_v1.CreateUserProfile, userGroup, organization string, schoolID int64) ([]*entities_bob.User, error)
	}
	UserGroupV2Repo interface {
		FindUserGroupAndRoleByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (map[string][]*entity.Role, error)
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *pby.CreateUserRequest) (*pby.CreateUserResponse, error) {
	profiles := make([]*pby_v1.CreateUserProfile, len(req.Users))

	for i, v := range req.Users {
		profiles[i] = &pby_v1.CreateUserProfile{
			Name:        v.Name,
			Country:     cpb.Country(cpb.Country_value[v.Country.String()]),
			PhoneNumber: v.PhoneNumber,
			Email:       v.Email,
			Avatar:      v.Avatar,
			GivenName:   v.GivenName,
			Grade:       v.Grade,
		}
	}

	resourcePath, err := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	users, err := s.UserModifierService.HandleCreateUser(ctx, profiles, req.UserGroup.String(), fmt.Sprint(resourcePath), int64(resourcePath))
	if err != nil {
		return nil, err
	}

	userProfiles := make([]*pby.UserProfile, 0, len(users))
	for _, v := range users {
		createdAt, err := types.TimestampProto(v.CreatedAt.Time)
		if err != nil {
			return nil, err
		}

		updatedAt := createdAt

		userProfiles = append(userProfiles, &pby.UserProfile{
			Id:          v.ID.String,
			Name:        v.GivenName.String + v.LastName.String,
			Country:     ppb.Country(ppb.Country_value[v.Country.String]),
			PhoneNumber: v.PhoneNumber.String,
			Email:       v.Email.String,
			Avatar:      v.Avatar.String,
			DeviceToken: v.DeviceToken.String,
			UserGroup:   v.Group.String,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	return &pby.CreateUserResponse{
		Users: userProfiles,
	}, nil
}

func (s *UserService) GetBasicProfile(ctx context.Context, req *pby.GetBasicProfileRequest) (*pby.GetBasicProfileResponse, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserService.GetBasicProfile")
	defer span.End()

	currentUserID := interceptors.UserIDFromContext(ctx)

	user, err := s.UserRepo.Get(ctx, s.DBPgx, database.Text(currentUserID))
	if err != nil {
		return nil, fmt.Errorf("s.UserRepo.Get: %w", err)
	}

	currentUser := user
	userGroup := currentUser.Group.String
	schools := []*pby.UserProfile_SchoolInfo{}
	schoolIDs := []int64{}

	switch userGroup {
	case constant.UserGroupTeacher:
		teacher, err := s.TeacherRepo.FindByID(ctx, s.DBPgx, database.Text(currentUserID))
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return nil, status.Error(codes.NotFound, "cannot find teacher")
			}

			return nil, fmt.Errorf("s.TeacherRepo.Get: %w", err)
		}

		sIDs := []int32{}
		for _, v := range teacher.SchoolIDs.Elements {
			sIDs = append(sIDs, v.Int)
		}

		enSchools, err := s.SchoolRepo.Get(ctx, s.DBPgx, sIDs)
		if err != nil {
			return nil, fmt.Errorf("s.SchoolRepo.Get: %w", err)
		}

		if len(enSchools) == 0 {
			return nil, status.Error(codes.NotFound, "cannot find schools")
		}

		for _, v := range enSchools {
			schools = append(schools, &pby.UserProfile_SchoolInfo{
				SchoolId:   int64(v.ID.Int),
				SchoolName: v.Name.String,
			})
			schoolIDs = append(schoolIDs, int64(v.ID.Int))
		}
	case constant.UserGroupSchoolAdmin:
		schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DBPgx, database.Text(currentUserID))
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return nil, status.Error(codes.NotFound, "cannot find school admin")
			}

			return nil, fmt.Errorf("s.SchoolAdminRepo.Get: %w", err)
		}

		enSchools, err := s.SchoolRepo.Get(ctx, s.DBPgx, []int32{schoolAdmin.SchoolID.Int})
		if err != nil {
			return nil, fmt.Errorf("s.SchoolRepo.Get: %w", err)
		}

		enSchool, ok := enSchools[schoolAdmin.SchoolID.Int]
		if !ok {
			return nil, status.Error(codes.NotFound, "cannot find school")
		}

		schools = []*pby.UserProfile_SchoolInfo{{
			SchoolId:   int64(enSchool.ID.Int),
			SchoolName: enSchool.Name.String,
		}}

		schoolIDs = []int64{int64(schoolAdmin.SchoolID.Int)}
	default:
		break
	}

	userGroupV2, err := s.getUserGroupV2(ctx, s.DBPgx, currentUserID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("getUserGroupV2: %w", err).Error())
	}

	return &pby.GetBasicProfileResponse{
		User: &pby.UserProfile{
			Schools:     schools,
			Id:          currentUserID,
			Name:        currentUser.GetName(),
			Country:     ppb.Country(ppb.Country_value[currentUser.Country.String]),
			PhoneNumber: currentUser.PhoneNumber.String,
			DeviceToken: currentUser.DeviceToken.String,
			UserGroup:   currentUser.Group.String,
			Email:       currentUser.Email.String,
			Avatar:      currentUser.Avatar.String,
			CreatedAt:   &types.Timestamp{Seconds: currentUser.CreatedAt.Time.Unix()},
			UpdatedAt:   &types.Timestamp{Seconds: currentUser.UpdatedAt.Time.Unix()},
			SchoolIds:   schoolIDs,
			UserGroupV2: userGroupV2,
		},
	}, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, req *pby.UpdateUserProfileRequest) (*pby.UpdateUserProfileResponse, error) {
	return &pby.UpdateUserProfileResponse{
		User: &pby.UserProfile{},
	}, nil
}

func (s *UserService) SyncStudent(ctx context.Context, req []*npb.EventUserRegistration_Student) error {
	var err error
	deleteIDs := []string{}
	for _, r := range req {
		switch r.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			errU := s.UpsertStudent(ctx, r)
			if errU != nil {
				err = multierr.Append(err, fmt.Errorf("s.UpsertStudent studentID %s: %w", r.StudentId, errU))
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			deleteIDs = append(deleteIDs, r.StudentId)
		}
	}

	errD := s.DeleteStudent(ctx, deleteIDs)
	if errD != nil {
		err = multierr.Append(err, fmt.Errorf("s.DeleteStudent studentIDs %v: %w", deleteIDs, errD))
	}

	return err
}

func (s *UserService) createStudent(ctx context.Context, tx pgx.Tx, additionalData *entities_bob.StudentAdditionalData, req *pby.CreateUserProfile, studentId string, schoolId int64) (*entities_bob.Student, error) {
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
		student.CurrentGrade.Set(req.Grade),
		student.ResourcePath.Set(fmt.Sprint(schoolId)),
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

func (s *UserService) UpsertStudent(ctx context.Context, req *npb.EventUserRegistration_Student) error {
	student, err := s.StudentRepo.Find(ctx, s.DBPgx, database.Text(req.StudentId))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("err GetUser: %w", err)
	}
	var fn func(ctx context.Context, tx pgx.Tx) error
	if student != nil {
		// update
		fn = func(ctx context.Context, tx pgx.Tx) error {
			additionalData := &entities_bob.StudentAdditionalData{
				JprefDivs: req.StudentDivs,
			}

			err := multierr.Combine(
				student.ID.Set(req.StudentId),
				student.AdditionalData.Set(additionalData),
				student.SchoolID.Set(constants.JPREPSchool),
				student.DeletedAt.Set(nil),
			)
			if err != nil {
				return fmt.Errorf("multierr.Combine: %w", err)
			}

			if err := s.StudentRepo.Update(ctx, tx, student); err != nil {
				return errors.Wrap(err, "s.StudentRepo.Update")
			}

			user, err := s.UserRepo.FindByIDUnscope(ctx, tx, student.ID)
			if err != nil {
				return fmt.Errorf("err FindUser: %w", err)
			}

			user.GivenName = database.Text(req.GivenName)
			user.LastName = database.Text(req.LastName)
			user.Country = database.Text(ppb.COUNTRY_JP.String())
			_ = user.DeletedAt.Set(nil)

			if err := s.UserRepo.Update(ctx, tx, user); err != nil {
				return fmt.Errorf("err Update: %w", err)
			}

			return nil
		}
	} else {
		// insert
		fn = func(ctx context.Context, tx pgx.Tx) error {
			additionalData := &entities_bob.StudentAdditionalData{
				JprefDivs: req.StudentDivs,
			}

			student, err = s.createStudent(ctx, tx, additionalData, &pby.CreateUserProfile{
				Country:     ppb.COUNTRY_JP,
				Name:        req.LastName,
				GivenName:   req.GivenName,
				PhoneNumber: req.StudentId, // to by pass not null contraint since JPREF does not send phoneNumber
			}, req.StudentId, constants.JPREPSchool)

			if err != nil {
				return err
			}

			return nil
		}
	}

	err = database.ExecInTx(ctx, s.DBPgx, fn)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteStudent(ctx context.Context, studentIDs []string) error {
	if len(studentIDs) == 0 {
		return nil
	}

	err := database.ExecInTx(ctx, s.DBPgx, func(ctx context.Context, tx pgx.Tx) error {
		err := s.StudentRepo.SoftDelete(ctx, tx, database.TextArray(studentIDs))
		if err != nil {
			return fmt.Errorf("s.StudentRepo.SoftDelete: %w", err)
		}

		err = s.UserRepo.SoftDelete(ctx, tx, database.TextArray(studentIDs))
		if err != nil {
			return fmt.Errorf("s.UserRepo.SoftDelete: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) SyncTeacher(ctx context.Context, req []*npb.EventUserRegistration_Staff) error {
	var err error
	deleteIDs := []string{}
	for _, r := range req {
		switch r.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			errU := s.UpsertTeacher(ctx, r)
			if errU != nil {
				err = multierr.Append(err, fmt.Errorf("s.UpsertTeacher teacherID %s: %w", r.StaffId, errU))
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			deleteIDs = append(deleteIDs, r.StaffId)
		}
	}

	errD := s.DeleteTeacher(ctx, deleteIDs)
	if errD != nil {
		err = multierr.Append(err, fmt.Errorf("s.DeleteTeacher teacherIDs %v: %w", deleteIDs, errD))
	}

	return err
}

func (s *UserService) UpsertTeacher(ctx context.Context, req *npb.EventUserRegistration_Staff) error {
	teacher, err := s.TeacherRepo.FindRegardlessDeletion(ctx, s.DBPgx, database.Text(req.StaffId))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("err GetUser: %w", err)
	}

	var fn func(ctx context.Context, tx pgx.Tx) error
	if teacher != nil {
		// update
		fn = func(ctx context.Context, tx pgx.Tx) error {
			_ = teacher.DeletedAt.Set(nil)
			if err := s.TeacherRepo.Update(ctx, tx, teacher); err != nil {
				return errors.Wrap(err, "s.TeacherRepo.Update")
			}

			user, err := s.UserRepo.FindByIDUnscope(ctx, tx, teacher.ID)
			if err != nil {
				return fmt.Errorf("err FindUser: %w", err)
			}

			user.LastName = database.Text(req.Name)
			user.Country = database.Text(ppb.COUNTRY_JP.String())
			_ = user.DeletedAt.Set(nil)

			if err := s.UserRepo.Update(ctx, tx, user); err != nil {
				return fmt.Errorf("err Update: %w", err)
			}

			return nil
		}
	} else {
		// insert
		fn = func(ctx context.Context, tx pgx.Tx) error {
			teacher = &entities_bob.Teacher{}
			database.AllNullEntity(teacher)
			database.AllNullEntity(&teacher.User)

			err := multierr.Combine(
				teacher.ID.Set(req.StaffId),
				teacher.PhoneNumber.Set(req.StaffId), // to by pass not null contraint since JPREF does not send phoneNumber
				teacher.LastName.Set(req.Name),
				teacher.Country.Set(ppb.COUNTRY_JP.String()),
				teacher.SchoolIDs.Set([]int32{constants.JPREPSchool}),
				teacher.ResourcePath.Set(fmt.Sprint(constants.JPREPSchool)),
			)
			if err != nil {
				return fmt.Errorf("multierr.Combine: %w", err)
			}

			if err := s.TeacherRepo.Create(ctx, tx, teacher); err != nil {
				return errors.Wrap(err, "s.TeacherRepo.Create")
			}

			return nil
		}
	}

	err = database.ExecInTx(ctx, s.DBPgx, fn)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteTeacher(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	err := database.ExecInTx(ctx, s.DBPgx, func(ctx context.Context, tx pgx.Tx) error {
		err := s.TeacherRepo.SoftDelete(ctx, tx, database.TextArray(ids))
		if err != nil {
			return fmt.Errorf("s.TeacherRepo.SoftDelete: %w", err)
		}

		err = s.UserRepo.SoftDelete(ctx, tx, database.TextArray(ids))
		if err != nil {
			return fmt.Errorf("s.UserRepo.SoftDelete: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) getUserGroupV2(ctx context.Context, db database.QueryExecer, userID string) ([]*pby.UserProfile_UserGroup, error) {
	userGroupAndRole, err := s.UserGroupV2Repo.FindUserGroupAndRoleByUserID(ctx, db, database.Text(userID))
	if err != nil {
		return nil, err
	}

	userGroupV2 := []*pby.UserProfile_UserGroup{}
	for userGroup, roleEntities := range userGroupAndRole {
		roles := []*pby.UserProfile_Role{}
		for _, role := range roleEntities {
			roles = append(roles, &pby.UserProfile_Role{
				Role:      role.RoleName.String,
				CreatedAt: &types.Timestamp{Seconds: role.CreatedAt.Time.Unix()},
			})
		}
		user := &pby.UserProfile_UserGroup{
			UserGroup: userGroup,
			Roles:     roles,
		}
		userGroupV2 = append(userGroupV2, user)
	}

	return userGroupV2, nil
}
