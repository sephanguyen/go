package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/vmihailenco/taskq/v3"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// ErrStudentEnrollmentStatusUnknown returned when create or edit student
	// that set enrollment status to a value not in StudentEnrollmentStatus predefined values
	ErrStudentEnrollmentStatusUnknown = errors.New("student enrollment status unknown")

	// ErrStudentEnrollmentStatusNotAllowedTobeNone returned when create or edit student
	// that set enrollment status to STUDENT_ENROLLMENT_STATUS_NONE
	ErrStudentEnrollmentStatusNotAllowedTobeNone = errors.New("student enrollment status not allowed to be STUDENT_ENROLLMENT_STATUS_NONE")

	// ErrTagIDsMustBeExisted returned when length of passed tag ids is not equal
	// with the length when finding tag records in database
	ErrTagIDsMustBeExisted = errors.New("tag ids must be existed in system")
	ErrTagIsNotForStudent  = errors.New("this tag is not for student")
	ErrTagIsNotForParent   = errors.New("this tag is not for parent")

	StableTemplateImportParentHeaders = "external_user_id,last_name,first_name,last_name_phonetic,first_name_phonetic,email,student_email,relationship,parent_tag,primary_phone_number,secondary_phone_number,remarks"
	StableTemplateImportParentValues  = "externaluserid,parent last name,parent first name,phonetic name,phonetic name,parent@email.com,student1@email.com;student2@email.com,1;2,tag_partner_id_1;tag_partner_id_2,parent_primary_phone_number,parent_secondary_phone_number,parent-remarks"
)

type OrganizationRepo interface {
	GetTenantIDByOrgID(ctx context.Context, db database.QueryExecer, orgID string) (string, error)
}

type UsrEmailRepo interface {
	Create(ctx context.Context, db database.QueryExecer, usrID pgtype.Text, email pgtype.Text) (*entity.UsrEmail, error)
	CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser) ([]*entity.UsrEmail, error)
	UpdateEmail(ctx context.Context, db database.QueryExecer, usrID, resourcePath, newEmail pgtype.Text) error
}

type UserAccessPathRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, userAccessPaths []*entity.UserAccessPath) error
	FindLocationIDsFromUserID(ctx context.Context, db database.QueryExecer, userID string) ([]string, error)
}

type DomainUsrEmailRepo interface {
	UpdateEmail(ctx context.Context, db database.QueryExecer, user entity.User) error
}

type DomainUserRepo interface {
	UpdateEmail(ctx context.Context, db database.QueryExecer, usersToUpdate entity.User) error
	GetByEmails(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error)
	GetByEmailsInsensitiveCase(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error)
	GetByUserNames(ctx context.Context, db database.QueryExecer, usernames []string) (entity.Users, error)
	GetByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.Users, error)
	GetByExternalUserIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) (entity.Users, error)
}

type DomainTagRepo interface {
	GetByIDs(ctx context.Context, db database.QueryExecer, tagIDs []string) (entity.DomainTags, error)
	GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) (entity.DomainTags, error)
}

type DomainTaggedUserRepo interface {
	GetByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) ([]entity.DomainTaggedUser, error)
	UpsertBatch(ctx context.Context, db database.QueryExecer, taggedUsers ...entity.DomainTaggedUser) error
	SoftDelete(ctx context.Context, db database.QueryExecer, taggedUsers ...entity.DomainTaggedUser) error
}

type StudentParentRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, studentParents []*entity.StudentParent) error
	FindParentIDsFromStudentID(ctx context.Context, db database.QueryExecer, studentID string) ([]string, error)
	RemoveParentFromStudent(ctx context.Context, db database.QueryExecer, parentID pgtype.Text, studentID pgtype.Text) error
	UpsertParentAccessPathByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error
	UpsertParentAccessPathByID(ctx context.Context, db database.QueryExecer, parentIDs []string) error
	FindStudentParentsByParentID(ctx context.Context, db database.QueryExecer, parentID string) ([]*entity.StudentParent, error)
	InsertParentAccessPathByStudentID(ctx context.Context, db database.QueryExecer, parentID, studentID string) error
}

type UserModifierService struct {
	pb.UnimplementedUserModifierServiceServer
	DB                 database.Ext
	FirebaseAuthClient internal_auth_tenant.TenantClient
	TenantManager      internal_auth_tenant.TenantManager
	FirebaseClient     firebase.AuthClient
	FatimaClient       fpb.SubscriptionModifierServiceClient
	JSM                nats.JetStreamManagement
	UnleashClient      unleashclient.ClientInstance
	Env                string

	OrganizationRepo     OrganizationRepo
	UsrEmailRepo         UsrEmailRepo
	DomainUserRepo       DomainUserRepo
	DomainUsrEmailRepo   DomainUsrEmailRepo
	DomainTagRepo        DomainTagRepo
	DomainTaggedUserRepo DomainTaggedUserRepo

	TaskQueue interface {
		Add(msg *taskq.Message) error
	}

	UserRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entity.LegacyUser, error)
		GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entity.LegacyUser, error)
		GetByEmailInsensitiveCase(ctx context.Context, db database.QueryExecer, emails []string) ([]*entity.LegacyUser, error)
		GetByPhone(ctx context.Context, db database.QueryExecer, phones pgtype.TextArray) ([]*entity.LegacyUser, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser) error
		UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
		UpdateEmail(ctx context.Context, db database.QueryExecer, u *entity.LegacyUser) error
		UpdateLastLoginDate(ctx context.Context, db database.QueryExecer, u *entity.LegacyUser) error
		UpdateProfileV1(ctx context.Context, db database.QueryExecer, u *entity.LegacyUser) error
		GetUserGroups(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserGroupV2, error)
		GetUserRoles(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (entity.Roles, error)
	}

	StudentRepo interface {
		Create(context.Context, database.QueryExecer, *entity.LegacyStudent) error
		Find(context.Context, database.QueryExecer, pgtype.Text) (*entity.LegacyStudent, error)
		Update(ctx context.Context, db database.QueryExecer, s *entity.LegacyStudent) error
		FindStudentProfilesByIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entity.LegacyStudent, error)
	}

	ParentRepo interface {
		GetByIds(ctx context.Context, db database.QueryExecer, parentIds pgtype.TextArray) (entity.Parents, error)
		Create(ctx context.Context, db database.QueryExecer, parent *entity.Parent) error
		CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entity.Parent) error
	}

	ImportUserEventRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, importUserEvents []*entity.ImportUserEvent) ([]*entity.ImportUserEvent, error)
	}

	StudentParentRepo StudentParentRepo

	TeacherRepo interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entity.Teacher) error
		UpsertMultiple(ctx context.Context, db database.QueryExecer, teachers []*entity.Teacher) error
		FindByID(context.Context, database.QueryExecer, pgtype.Text) (*entity.Teacher, error)
		SoftDeleteMultiple(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray) error
	}

	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entity.SchoolAdmin, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, schoolAdmins []*entity.SchoolAdmin) error
		UpsertMultiple(ctx context.Context, db database.QueryExecer, schoolAdmins []*entity.SchoolAdmin) error
		SoftDeleteMultiple(ctx context.Context, db database.QueryExecer, schoolAdminIDs pgtype.TextArray) error
	}

	UserGroupV2Repo interface {
		FindUserGroupByRoleName(ctx context.Context, db database.QueryExecer, roleName string) (*entity.UserGroupV2, error)
	}

	UserGroupRepo interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entity.UserGroup) error
		UpdateStatus(ctx context.Context, db database.QueryExecer, userID, status pgtype.Text) error
	}

	UserGroupsMemberRepo interface {
		AssignWithUserGroup(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser, userGroupID pgtype.Text) error
	}

	UserAccessPathRepo UserAccessPathRepo

	LocationRepo interface {
		GetLocationsByLocationIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, allowDeleted bool) ([]*domain.Location, error)
		GetLocationsByPartnerInternalIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) ([]*domain.Location, error)
	}

	SchoolHistoryRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, schoolHistories []*entity.SchoolHistory) error
		SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error
		GetSchoolHistoriesByGradeIDAndStudentID(ctx context.Context, db database.QueryExecer, gradeID pgtype.Text, studentID pgtype.Text, isCurrent pgtype.Bool) ([]*entity.SchoolHistory, error)
		SetCurrentSchoolByStudentIDAndSchoolID(ctx context.Context, db database.QueryExecer, schoolID pgtype.Text, studentID pgtype.Text) error
		RemoveCurrentSchoolByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) error
		UnsetCurrentSchoolByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) error
		GetCurrentSchoolByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entity.SchoolHistory, error)
	}

	SchoolInfoRepo interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.SchoolInfo, error)
	}

	SchoolCourseRepo interface {
		GetByIDsAndSchoolIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, schoolIDs pgtype.TextArray) ([]*entity.SchoolCourse, error)
	}

	UserAddressRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userAddresses []*entity.UserAddress) error
		SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) error
		GetByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserAddress, error)
	}

	PrefectureRepo interface {
		GetByPrefectureID(ctx context.Context, db database.QueryExecer, prefectureID pgtype.Text) (*entity.Prefecture, error)
		GetByPrefectureIDs(ctx context.Context, db database.QueryExecer, prefectureIDs pgtype.TextArray) ([]*entity.Prefecture, error)
	}

	UserPhoneNumberRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userPhoneNumbers []*entity.UserPhoneNumber) error
		SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) error
	}

	DomainGradeRepo interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]entity.DomainGrade, error)
		GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) ([]entity.DomainGrade, error)
	}

	GradeOrganizationRepo interface {
		GetByGradeIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*repository.GradeOrganization, error)
	}
	InternalConfigurationRepo interface {
		GetByKey(ctx context.Context, db database.QueryExecer, configKey string) (entity.DomainConfiguration, error)
	}

	DomainStudentRepo interface {
		GetByEmails(ctx context.Context, db database.QueryExecer, emails []string) ([]entity.DomainStudent, error)
	}
}

func (s *UserModifierService) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentResponse, error) {
	zapLogger := ctxzap.Extract(ctx)
	sugaredZapLogger := zapLogger.Sugar()

	if err := validCreateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tags, err := s.DomainTagRepo.GetByIDs(ctx, s.DB, req.StudentProfile.TagIds)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "DomainTagRepo.GetByIDs").Error())
	}

	if err := validUserTags(constant.RoleStudent, req.GetStudentProfile().GetTagIds(), tags); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	locations, err := s.GetLocations(ctx, req.StudentProfile.LocationIds)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	var studentPB *pb.Student
	var studentPhoneNumberPB *pb.StudentPhoneNumber
	// var parentProfilePBs []*pb.ParentProfileResponse

	// Convert student data in request to entities
	student, err := studentPbToStudentEntity(int32(resourcePath), req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	gradeMaster, err := s.getGradeMaster(ctx, req.StudentProfile.GradeId)
	if err != nil {
		return nil, err
	}
	if len(gradeMaster) > 0 {
		for k, v := range gradeMaster {
			currentGrade := student.CurrentGrade.Int
			gradeID := k.GradeID().String()
			if !(field.IsNull(v) && field.IsUndefined(v)) {
				currentGrade = int16(v.Int32())
			}
			if err := multierr.Combine(
				student.CurrentGrade.Set(currentGrade),
				student.GradeID.Set(gradeID),
			); err != nil {
				return nil, status.Errorf(codes.Internal, "multierr.Combine err: %v", err)
			}
		}
	}

	// Guarantee each user email has only corresponding one uid
	createdUsrEmail, err := s.UsrEmailRepo.Create(ctx, s.DB, student.ID, student.LoginEmail)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	student.ID = createdUsrEmail.UsrID
	student.LegacyUser.ID = createdUsrEmail.UsrID

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Valid email and phone number in student data
		existingUsers, err := s.UserRepo.GetByEmail(ctx, tx, database.TextArray([]string{student.Email.String}))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", err).Error())
		}
		if len(existingUsers) > 0 {
			return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create student with emails existing in system: %s", student.Email.String))
		}

		if student.PhoneNumber.String != "" && req.StudentProfile.StudentPhoneNumber == nil {
			phoneNumberExistingStudents, err := s.UserRepo.GetByPhone(ctx, tx, database.TextArray([]string{student.PhoneNumber.String}))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByPhone: %w", err).Error())
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

		/*
		 * TODO:
		 *   if user group is stable, return error directly when user group not found
		 */

		// find student user group id if existed then assign this user group for created user
		studentUserGroup, err := s.UserGroupV2Repo.FindUserGroupByRoleName(ctx, tx, constant.RoleStudent)
		if err == nil {
			if err := s.UserGroupsMemberRepo.AssignWithUserGroup(ctx, tx, []*entity.LegacyUser{&student.LegacyUser}, studentUserGroup.UserGroupID); err != nil {
				return status.Error(codes.Internal, errors.Wrapf(err, "can not assign student user group to user %s", student.GetUID()).Error())
			}
		} else {
			sugaredZapLogger.Warn(errors.Wrap(err, "can not find student user group"))
		}

		// create user_resource_path (add locations for student)
		if err := UpsertUserAccessPath(ctx, s.UserAccessPathRepo, tx, locations, student.ID.String); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		studentPB.UserProfile.LocationIds = req.StudentProfile.LocationIds

		// Create firebase accounts for student
		identityPlatformAccounts := entity.LegacyUsers{&student.LegacyUser}

		// Import to identity platform
		tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, tx, strconv.FormatInt(resourcePath, 10))
		if err != nil {
			zapLogger.Error(
				"cannot get tenant id",
				zap.Error(err),
				zap.Int64("organizationID", resourcePath),
			)
			switch err {
			case pgx.ErrNoRows:
				return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: strconv.FormatInt(resourcePath, 10)}.Error())
			default:
				return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
			}
		}

		err = s.CreateUsersInIdentityPlatform(ctx, tenantID, identityPlatformAccounts, resourcePath)
		if err != nil {
			zapLogger.Error(
				"cannot create users on identity platform",
				zap.Error(err),
				zap.Int64("organizationID", resourcePath),
				zap.String("tenantID", tenantID),
				zap.Strings("emails", identityPlatformAccounts.Limit(10).Emails()),
			)
			switch err {
			case internal_auth_user.ErrUserNotFound:
				return status.Error(codes.NotFound, errcode.NewUserNotFoundErr(identityPlatformAccounts.IDs()...).Error())
			default:
				return status.Error(codes.Internal, err.Error())
			}
		}

		// Upsert school_history
		if err := s.validateSchoolHistoriesReq(ctx, req.SchoolHistories); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("validateSchoolHistoriesReq: %v", err).Error())
		}
		schoolHistories, err := schoolHistoryPbToStudentSchoolHistory(req.SchoolHistories, student.ID.String, fmt.Sprint(resourcePath))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("schoolHistoryPbToStudentSchoolHistory: %v", err).Error())
		}
		if err := s.SchoolHistoryRepo.SoftDeleteByStudentIDs(ctx, tx, database.TextArray([]string{student.ID.String})); err != nil {
			return errorx.ToStatusError(err)
		}
		if err := s.SchoolHistoryRepo.Upsert(ctx, tx, schoolHistories); err != nil {
			return errorx.ToStatusError(err)
		}

		// Upsert user_address
		if err := s.validateUserAddressesReq(ctx, req.UserAddresses); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("validateUserAddressesReq: %v", err).Error())
		}
		homeAddresses, err := userAddressPbToStudentHomeAddress(req.UserAddresses, student.LegacyUser.ID.String, fmt.Sprint(resourcePath))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("userAddressPbToStudentHomeAddress: %v", err).Error())
		}
		if err := s.UserAddressRepo.SoftDeleteByUserIDs(ctx, tx, database.TextArray([]string{student.LegacyUser.ID.String})); err != nil {
			return errorx.ToStatusError(err)
		}
		if err := s.UserAddressRepo.Upsert(ctx, tx, homeAddresses); err != nil {
			return errorx.ToStatusError(err)
		}

		if req.StudentProfile.StudentPhoneNumber != nil {
			if err := validateStudentPhoneNumber(req.StudentProfile.StudentPhoneNumber); err != nil {
				return status.Error(codes.InvalidArgument, fmt.Errorf("validateStudentPhoneNumber: %v", err).Error())
			}
			userPhoneNumbers, err := userPhoneNumbersPbToStudentPhoneNumbers(req.StudentProfile.StudentPhoneNumber, student.LegacyUser.ID.String, fmt.Sprint(resourcePath))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("userPhoneNumbersPbToStudentPhoneNumbers: %v", err).Error())
			}
			if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
				return errorx.ToStatusError(err)
			}
			studentPhoneNumberPB = &pb.StudentPhoneNumber{
				PhoneNumber:       req.StudentProfile.StudentPhoneNumber.PhoneNumber,
				HomePhoneNumber:   req.StudentProfile.StudentPhoneNumber.HomePhoneNumber,
				ContactPreference: req.StudentProfile.StudentPhoneNumber.ContactPreference,
			}
		}

		// after split port layer completely, we should update below code style here,
		// we shouldn't use port/repo in the domain layer.
		userProfile := grpc.NewUserProfile(studentPB.UserProfile)
		userWithTags := map[entity.User][]entity.DomainTag{userProfile: tags}
		if err := s.UpsertTaggedUsers(ctx, tx, userWithTags, nil); err != nil {
			return errors.Wrap(err, "UpsertTaggedUsers")
		}

		userEvents := newCreateStudentEvents(int32(resourcePath), locations, student)
		if err = s.publishUserEvent(ctx, constants.SubjectUserCreated, userEvents...); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(req.SchoolHistories) != 0 {
		currentSchools, err := s.SchoolHistoryRepo.GetSchoolHistoriesByGradeIDAndStudentID(ctx, s.DB, database.Text(req.StudentProfile.GradeId), database.Text(student.ID.String), database.Bool(false))
		if err != nil {
			return nil, err
		}
		if len(currentSchools) != 0 {
			schoolIDs := []string{}
			for _, currentSchool := range currentSchools {
				schoolIDs = append(schoolIDs, currentSchool.SchoolID.String)
			}
			err = s.SchoolHistoryRepo.SetCurrentSchoolByStudentIDAndSchoolID(ctx, s.DB, database.Text(schoolIDs[0]), database.Text(student.ID.String))
			if err != nil {
				return nil, err
			}
		}
	}

	response := &pb.CreateStudentResponse{
		StudentProfile: &pb.CreateStudentResponse_StudentProfile{
			Student:            studentPB,
			StudentPassword:    req.StudentProfile.Password,
			StudentPhoneNumber: studentPhoneNumberPB,
		},
	}

	return response, nil
}

func schoolHistoryPbToStudentSchoolHistory(schoolHistories []*pb.SchoolHistory, studentID, resourcePath string) ([]*entity.SchoolHistory, error) {
	schoolHistoryEntities := make([]*entity.SchoolHistory, 0, len(schoolHistories))
	for _, schoolHistory := range schoolHistories {
		schoolHistoryEntity := &entity.SchoolHistory{}
		database.AllNullEntity(schoolHistoryEntity)

		if schoolHistory.StartDate.IsValid() {
			if err := schoolHistoryEntity.StartDate.Set(schoolHistory.StartDate.AsTime()); err != nil {
				return nil, fmt.Errorf("schoolHistoryPbToStudentSchoolHistory schoolHistoryEntity.StartDate.Set: %v", err)
			}
		}

		if schoolHistory.EndDate.IsValid() {
			if err := schoolHistoryEntity.EndDate.Set(schoolHistory.EndDate.AsTime()); err != nil {
				return nil, fmt.Errorf("schoolHistoryPbToStudentSchoolHistory schoolHistoryEntity.EndDate.Set: %v", err)
			}
		}

		if schoolHistory.SchoolCourseId != "" {
			if err := schoolHistoryEntity.SchoolCourseID.Set(schoolHistory.SchoolCourseId); err != nil {
				return nil, fmt.Errorf("schoolHistoryPbToStudentSchoolHistory schoolHistoryEntity.SchoolCourseID.Set: %v", err)
			}
		}

		err := multierr.Combine(
			schoolHistoryEntity.StudentID.Set(studentID),
			schoolHistoryEntity.ResourcePath.Set(resourcePath),
			schoolHistoryEntity.IsCurrent.Set(false),
			schoolHistoryEntity.SchoolID.Set(schoolHistory.SchoolId),
		)
		if err != nil {
			return nil, fmt.Errorf("schoolHistoryPbToStudentSchoolHistory multierr.Combine: %v", err)
		}

		schoolHistoryEntities = append(schoolHistoryEntities, schoolHistoryEntity)
	}
	return schoolHistoryEntities, nil
}

func (s *UserModifierService) GetGradeMaster(ctx context.Context, gradeID string) (map[entity.DomainGrade]field.Int32, error) {
	return s.getGradeMaster(ctx, gradeID)
}

func (s *UserModifierService) getGradeMaster(ctx context.Context, gradeID string) (map[entity.DomainGrade]field.Int32, error) {
	mapGradeMaster := map[entity.DomainGrade]field.Int32{}
	if gradeID == "" {
		return mapGradeMaster, nil
	}

	grades, err := s.DomainGradeRepo.GetByIDs(ctx, s.DB, []string{gradeID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "s.DomainGradeRepo.GetByIDs err: %v", err)
	}

	if len(grades) == 0 {
		grades, err = s.DomainGradeRepo.GetByPartnerInternalIDs(ctx, s.DB, []string{gradeID})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "s.DomainGradeRepo.GetByPartnerInternalIDs err: %v", err)
		}
		if len(grades) == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "grade_id does not exist")
		}
	}

	// for _, grade := range grades {
	// 	if grade.IsArchived().Boolean() {
	// 		return nil, status.Errorf(codes.InvalidArgument, "grade is archived")
	// 	}
	// }

	mapGradeMaster[grades[0]] = field.NewNullInt32()
	gradeOrgs, err := s.GradeOrganizationRepo.GetByGradeIDs(ctx, s.DB, []string{grades[0].GradeID().String()})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "s.GradeOrganizationRepo.GetByGradeIDs err: %v", err)
	}
	if len(gradeOrgs) > 0 {
		mapGradeMaster[grades[0]] = gradeOrgs[0].GradeValue()
	}
	return mapGradeMaster, nil
}

func (s *UserModifierService) validateUserAddressesReq(ctx context.Context, userAddressesReq []*pb.UserAddress) error {
	for idx, userAddressReq := range userAddressesReq {
		if userAddressReq.AddressType != pb.AddressType_HOME_ADDRESS {
			return fmt.Errorf("address_type cannot be other type, must be HOME_ADDRESS: %v", idx+1)
		}

		if userAddressReq.Prefecture != "" {
			_, err := s.PrefectureRepo.GetByPrefectureID(ctx, s.DB, database.Text(userAddressReq.Prefecture))
			if err != nil {
				return fmt.Errorf("PrefectureRepo.GetByPrefectureCode: %v", err)
			}
		}
	}

	return nil
}

func userAddressPbToStudentHomeAddress(userAddressesReq []*pb.UserAddress, userID, resourcePath string) ([]*entity.UserAddress, error) {
	userAddressEntities := make([]*entity.UserAddress, 0, len(userAddressesReq))
	for _, userAddressReq := range userAddressesReq {
		userAddressEntity := &entity.UserAddress{}
		database.AllNullEntity(userAddressEntity)

		if userAddressReq.AddressId == "" {
			userAddressReq.AddressId = idutil.ULIDNow()
		}
		if userAddressReq.Prefecture == "" {
			userAddressEntity.PrefectureID.Set(sql.NullString{})
		} else {
			userAddressEntity.PrefectureID.Set(userAddressReq.Prefecture)
		}

		err := multierr.Combine(
			userAddressEntity.UserAddressID.Set(userAddressReq.AddressId),
			userAddressEntity.UserID.Set(userID),
			userAddressEntity.AddressType.Set(pb.AddressType_name[int32(userAddressReq.AddressType)]),
			userAddressEntity.PostalCode.Set(userAddressReq.PostalCode),
			userAddressEntity.City.Set(userAddressReq.City),
			userAddressEntity.FirstStreet.Set(userAddressReq.FirstStreet),
			userAddressEntity.SecondStreet.Set(userAddressReq.SecondStreet),

			userAddressEntity.ResourcePath.Set(resourcePath),
		)
		if err != nil {
			return nil, fmt.Errorf("userAddressPbToStudentHomeAddress multierr.Combine: %v", err)
		}
		userAddressEntities = append(userAddressEntities, userAddressEntity)
	}
	return userAddressEntities, nil
}

func validateStudentPhoneNumbersUpdateStudent(userPhoneNumbers []*pb.StudentPhoneNumberWithID) error {
	var studentPhoneNumber, studentHomePhoneNumber *pb.StudentPhoneNumberWithID
	if len(userPhoneNumbers) > 2 {
		return fmt.Errorf("phone numbers's length must less or equal 2")
	}
	for _, phoneNumber := range userPhoneNumbers {
		switch phoneNumber.PhoneNumberType {
		case pb.StudentPhoneNumberType_PHONE_NUMBER:
			studentPhoneNumber = phoneNumber
		case pb.StudentPhoneNumberType_HOME_PHONE_NUMBER:
			studentHomePhoneNumber = phoneNumber
		}
	}
	if studentPhoneNumber != nil && studentHomePhoneNumber != nil && studentPhoneNumber.PhoneNumber == studentHomePhoneNumber.PhoneNumber {
		return fmt.Errorf("phone number and home phone number must not be the same")
	}
	return nil
}

func validateStudentPhoneNumber(userPhoneNumber *pb.StudentPhoneNumber) error {
	if userPhoneNumber.PhoneNumber != "" {
		err := MatchingRegex(PhoneNumberPattern, userPhoneNumber.PhoneNumber)
		if err != nil {
			return err
		}
	}
	if userPhoneNumber.HomePhoneNumber != "" {
		err := MatchingRegex(PhoneNumberPattern, userPhoneNumber.HomePhoneNumber)
		if err != nil {
			return err
		}
	}
	if userPhoneNumber.PhoneNumber != "" && userPhoneNumber.HomePhoneNumber != "" && userPhoneNumber.PhoneNumber == userPhoneNumber.HomePhoneNumber {
		return fmt.Errorf("phone number and home phone number must not be the same")
	}
	return nil
}

func updateUserPhoneNumbersPbToStudentPhoneNumbers(updateStudentPhoneNumbers []*pb.StudentPhoneNumberWithID, userID string, resourcePath string) ([]*entity.UserPhoneNumber, string, string, error) {
	userPhoneNumbers := make([]*entity.UserPhoneNumber, 0)
	var phoneNumber, homePhoneNumber string
	var setPhoneNumberEntity = func(phoneNumber string, phoneNumberType string) (*entity.UserPhoneNumber, error) {
		studentPhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(studentPhoneNumber)
		userPhoneNumberID := idutil.ULIDNow()
		if err := multierr.Combine(
			studentPhoneNumber.ID.Set(userPhoneNumberID),
			studentPhoneNumber.UserID.Set(userID),
			studentPhoneNumber.PhoneNumber.Set(phoneNumber),
			studentPhoneNumber.PhoneNumberType.Set(phoneNumberType),
			studentPhoneNumber.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, fmt.Errorf("userPhoneNumbersPbToStudentPhoneNumbers multierr.Combine: %v", err)
		}
		return studentPhoneNumber, nil
	}
	for _, userPhoneNumber := range updateStudentPhoneNumbers {
		switch userPhoneNumber.PhoneNumberType {
		case pb.StudentPhoneNumberType_PHONE_NUMBER:
			phoneNumber = userPhoneNumber.PhoneNumber
			studentPhoneNumber, err := setPhoneNumberEntity(userPhoneNumber.PhoneNumber, entity.StudentPhoneNumber)
			if err != nil {
				return nil, "", "", fmt.Errorf("userPhoneNumbersPbToStudentPhoneNumbers multierr.Combine: %v", err)
			}
			userPhoneNumbers = append(userPhoneNumbers, studentPhoneNumber)
		case pb.StudentPhoneNumberType_HOME_PHONE_NUMBER:
			homePhoneNumber = userPhoneNumber.PhoneNumber
			studentHomePhoneNumber, err := setPhoneNumberEntity(userPhoneNumber.PhoneNumber, entity.StudentHomePhoneNumber)
			if err != nil {
				return nil, "", "", fmt.Errorf("userPhoneNumbersPbToStudentPhoneNumbers multierr.Combine: %v", err)
			}
			userPhoneNumbers = append(userPhoneNumbers, studentHomePhoneNumber)
		}
	}
	return userPhoneNumbers, phoneNumber, homePhoneNumber, nil
}

func userPhoneNumbersPbToStudentPhoneNumbers(userPhoneNumbersReq *pb.StudentPhoneNumber, userID string, resourcePath string) ([]*entity.UserPhoneNumber, error) {
	userPhoneNumbers := make([]*entity.UserPhoneNumber, 0)
	if userPhoneNumbersReq.PhoneNumber != "" {
		studentPhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(studentPhoneNumber)
		userPhoneNumberID := idutil.ULIDNow()
		if err := multierr.Combine(
			studentPhoneNumber.ID.Set(userPhoneNumberID),
			studentPhoneNumber.UserID.Set(userID),
			studentPhoneNumber.PhoneNumber.Set(userPhoneNumbersReq.PhoneNumber),
			studentPhoneNumber.PhoneNumberType.Set(entity.StudentPhoneNumber),
			studentPhoneNumber.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, fmt.Errorf("userPhoneNumbersPbToStudentPhoneNumbers multierr.Combine: %v", err)
		}
		userPhoneNumbers = append(userPhoneNumbers, studentPhoneNumber)
	}

	if userPhoneNumbersReq.HomePhoneNumber != "" {
		studentHomePhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(studentHomePhoneNumber)
		userPhoneNumberID := idutil.ULIDNow()
		if err := multierr.Combine(
			studentHomePhoneNumber.ID.Set(userPhoneNumberID),
			studentHomePhoneNumber.UserID.Set(userID),
			studentHomePhoneNumber.PhoneNumber.Set(userPhoneNumbersReq.HomePhoneNumber),
			studentHomePhoneNumber.PhoneNumberType.Set(entity.StudentHomePhoneNumber),
			studentHomePhoneNumber.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, fmt.Errorf("userPhoneNumbersPbToStudentPhoneNumbers multierr.Combine: %v", err)
		}
		userPhoneNumbers = append(userPhoneNumbers, studentHomePhoneNumber)
	}

	return userPhoneNumbers, nil
}

func validCreateRequest(req *pb.CreateStudentRequest) error {
	if req.StudentProfile == nil {
		return errors.New("student profile is null")
	}

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
	case req.StudentProfile.Name == "" && req.StudentProfile.FirstName == "" && req.StudentProfile.LastName == "":
		return errors.New("student name cannot be empty")
	case req.StudentProfile.Name == "" && req.StudentProfile.FirstName == "":
		return errors.New("student first name cannot be empty")
	case req.StudentProfile.Name == "" && req.StudentProfile.LastName == "":
		return errors.New("student last name cannot be empty")
	case len(req.StudentProfile.LocationIds) < 1:
		return errors.New("student location length must be at least 1")
	}

	if _, ok := pb.StudentEnrollmentStatus_value[req.StudentProfile.EnrollmentStatus.String()]; !ok {
		return ErrStudentEnrollmentStatusUnknown
	}
	if req.StudentProfile.EnrollmentStatus == pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
		return ErrStudentEnrollmentStatusNotAllowedTobeNone
	}

	return nil
}

func validUserTags(role string, tagIDs []string, existingTags entity.DomainTags) error {
	if ok := existingTags.ContainIDs(tagIDs...); !ok {
		return ErrTagIDsMustBeExisted
	}

	for _, tag := range existingTags {
		if role == constant.RoleParent && !entity.IsParentTag(tag) {
			return ErrTagIsNotForParent
		}

		if role == constant.RoleStudent && !entity.IsStudentTag(tag) {
			return ErrTagIsNotForStudent
		}
	}

	return nil
}

func (s *UserModifierService) validateSchoolHistoriesReq(ctx context.Context, schoolHistories []*pb.SchoolHistory) error {
	schoolIDs := []string{}
	courseIDs := []string{}

	for i, schoolHistory := range schoolHistories {
		if schoolHistory.SchoolId == "" {
			return fmt.Errorf("school_id cannot be empty at row: %v", i+1)
		}

		if schoolHistory.StartDate.IsValid() && schoolHistory.EndDate.IsValid() {
			startDate := schoolHistory.StartDate.AsTime()
			endDate := schoolHistory.EndDate.AsTime()

			if startDate.After(endDate) {
				return fmt.Errorf("start_date must be before end_date at row: %v", i+1)
			}
		}

		if schoolHistory.SchoolCourseId != "" {
			courseIDs = append(courseIDs, schoolHistory.SchoolCourseId)
		}

		schoolIDs = append(schoolIDs, schoolHistory.SchoolId)
	}

	schoolInfos, err := s.SchoolInfoRepo.GetByIDs(ctx, s.DB, database.TextArray(schoolIDs))
	if err != nil {
		return fmt.Errorf("schoolInfoRepo.GetByIDs: %v", err)
	}
	if len(schoolInfos) != len(schoolIDs) {
		return fmt.Errorf("school_info does not match with req.SchoolHistories")
	}

	levelWithSchoolInfo := map[string]string{}
	for _, schoolInfo := range schoolInfos {
		if schoolInfo.IsArchived.Bool {
			return fmt.Errorf("school_info %v is archived", schoolInfo.ID.String)
		}
		if _, ok := levelWithSchoolInfo[schoolInfo.LevelID.String]; ok {
			return fmt.Errorf("duplicate school_level_id in school_info %v", schoolInfo.ID.String)
		}
		levelWithSchoolInfo[schoolInfo.LevelID.String] = schoolInfo.ID.String
	}

	schoolCourses, err := s.SchoolCourseRepo.GetByIDsAndSchoolIDs(ctx, s.DB, database.TextArray(courseIDs), database.TextArray(schoolIDs))
	if err != nil {
		return fmt.Errorf("schoolCourseRepo.GetByIDsAndSchoolIDs: %v", err)
	}
	if len(schoolCourses) != len(courseIDs) {
		return fmt.Errorf("school_course does not match with req.SchoolHistories")
	}

	for _, schoolCourse := range schoolCourses {
		if schoolCourse.IsArchived.Bool {
			return fmt.Errorf("school_course %v is archived", schoolCourse.ID.String)
		}
	}

	return nil
}

func studentPbToStudentEntity(schoolID int32, req *pb.CreateStudentRequest) (*entity.LegacyStudent, error) {
	studentEnt := &entity.LegacyStudent{}
	database.AllNullEntity(studentEnt)
	database.AllNullEntity(&studentEnt.LegacyUser)
	studentID := idutil.ULIDNow()
	enrollmentStatus := req.StudentProfile.EnrollmentStatus.String()
	if req.StudentProfile.EnrollmentStatusStr != "" {
		enrollmentStatus = req.StudentProfile.EnrollmentStatusStr
	}
	if err := multierr.Combine(
		studentEnt.LegacyUser.ID.Set(studentID),
		studentEnt.LegacyUser.Email.Set(req.StudentProfile.Email),
		// nolint:staticcheck //lint:ignore SA1019 Ignore the deprecation warnings until we completely remove this field
		studentEnt.LegacyUser.PhoneNumber.Set(req.StudentProfile.PhoneNumber),
		// nolint:staticcheck //lint:ignore SA1019 Ignore the deprecation warnings until we completely remove this field
		studentEnt.LegacyUser.FullName.Set(req.StudentProfile.Name),
		studentEnt.LegacyUser.FirstName.Set(""),
		studentEnt.LegacyUser.LastName.Set(""),
		studentEnt.FirstNamePhonetic.Set(req.StudentProfile.FirstNamePhonetic),
		studentEnt.LastNamePhonetic.Set(req.StudentProfile.LastNamePhonetic),
		studentEnt.LegacyUser.Country.Set(req.StudentProfile.CountryCode.String()),
		studentEnt.ID.Set(studentID),
		studentEnt.SchoolID.Set(schoolID),
		studentEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
		studentEnt.CurrentGrade.Set(req.StudentProfile.Grade),
		studentEnt.EnrollmentStatus.Set(enrollmentStatus),
		studentEnt.StudentNote.Set(req.StudentProfile.StudentNote),
		studentEnt.LegacyUser.Group.Set(entity.UserGroupStudent),
		studentEnt.LegacyUser.UserRole.Set(constant.UserRoleStudent),
		studentEnt.LegacyUser.UserName.Set(req.StudentProfile.Email),
		studentEnt.LegacyUser.LoginEmail.Set(req.StudentProfile.Email),
	); err != nil {
		return nil, err
	}
	if req.StudentProfile.FirstName != "" && req.StudentProfile.LastName != "" {
		if err := multierr.Combine(
			studentEnt.LegacyUser.FullName.Set(CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName)),
			studentEnt.LegacyUser.FirstName.Set(req.StudentProfile.FirstName),
			studentEnt.LegacyUser.LastName.Set(req.StudentProfile.LastName),
		); err != nil {
			return nil, err
		}
	}

	fullNamePhonetic := CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic)
	if fullNamePhonetic != "" {
		if err := studentEnt.LegacyUser.FullNamePhonetic.Set(fullNamePhonetic); err != nil {
			return nil, err
		}
	}

	if err := req.StudentProfile.Birthday.CheckValid(); err == nil {
		_ = studentEnt.LegacyUser.Birthday.Set(req.StudentProfile.Birthday.AsTime())
	}
	if req.StudentProfile.Gender != pb.Gender_NONE {
		_ = studentEnt.LegacyUser.Gender.Set(req.StudentProfile.Gender.String())
	}

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

	if req.StudentProfile.StudentPhoneNumber != nil {
		if err := studentEnt.ContactPreference.Set(req.StudentProfile.StudentPhoneNumber.ContactPreference.String()); err != nil {
			return nil, err
		}

		if req.StudentProfile.StudentPhoneNumber.PhoneNumber != "" {
			if err := studentEnt.PhoneNumber.Set(req.StudentProfile.StudentPhoneNumber.PhoneNumber); err != nil {
				return nil, err
			}
		}
	}

	studentEnt.UserAdditionalInfo.Password = req.StudentProfile.Password

	return studentEnt, nil
}

func studentToStudentPBInCreateStudentResponse(student *entity.LegacyStudent) *pb.Student {
	studentPB := &pb.Student{
		UserProfile: &pb.UserProfile{
			UserId:            student.ID.String,
			Email:             student.Email.String,
			Name:              student.GetName(),
			Avatar:            student.Avatar.String,
			Group:             pb.UserGroup(pb.UserGroup_value[student.Group.String]),
			PhoneNumber:       student.PhoneNumber.String,
			CountryCode:       cpb.Country(cpb.Country_value[student.LegacyUser.Country.String]),
			FirstName:         student.FirstName.String,
			LastName:          student.LastName.String,
			FirstNamePhonetic: student.FirstNamePhonetic.String,
			LastNamePhonetic:  student.LastNamePhonetic.String,
			FullNamePhonetic:  student.FullNamePhonetic.String,
		},
		Grade:    int32(student.CurrentGrade.Int),
		SchoolId: student.SchoolID.Int,
		GradeId:  student.GradeID.String,
	}

	if student.Birthday.Status != pgtype.Null {
		studentPB.UserProfile.Birthday = timestamppb.New(student.Birthday.Time)
	}
	if student.Gender.Status != pgtype.Null {
		studentPB.UserProfile.Gender = pb.Gender(pb.Gender_value[student.Gender.String])
	}

	return studentPB
}

func (s *UserModifierService) publishAsyncUserEvent(ctx context.Context, userEvents ...*pb.EvtUser) error {
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

func (s *UserModifierService) CreateUsersInIdentityPlatform(ctx context.Context, tenantID string, users []*entity.LegacyUser, resourcePath int64) error {
	zapLogger := ctxzap.Extract(ctx)

	tenantClient, err := s.TenantManager.TenantClient(ctx, tenantID)
	if err != nil {
		zapLogger.Sugar().Warnw(
			"cannot get tenant client",
			"tenantID", tenantID,
			"err", err.Error(),
		)
		return errors.Wrap(err, "TenantClient")
	}

	err = createUserInAuthPlatform(ctx, tenantClient, users, resourcePath)
	if err != nil {
		zapLogger.Sugar().Warnw(
			"cannot create users on identity platform",
			"err", err.Error(),
		)
		return errors.Wrap(err, "createUserInAuthPlatform")
	}

	return nil
}

func CreateUserInAuthPlatform(ctx context.Context, authClient internal_auth_tenant.TenantClient, users []*entity.LegacyUser, schoolID int64) error {
	return createUserInAuthPlatform(ctx, authClient, users, schoolID)
}

func createUserInAuthPlatform(ctx context.Context, authClient internal_auth_tenant.TenantClient, users []*entity.LegacyUser, schoolID int64) error {
	var authUsers internal_auth_user.Users
	for i := range users {
		users[i].CustomClaims = utils.CustomUserClaims(users[i].Group.String, users[i].ID.String, schoolID)

		passwordSalt := []byte(idutil.ULIDNow())

		var hashedPwd []byte
		if users[i].Password != "" {
			var err error
			hashedPwd, err = internal_auth.HashedPassword(authClient.GetHashConfig(), []byte(users[i].Password), passwordSalt)
			if err != nil {
				return errors.Wrap(err, "HashedPassword")
			}
		}

		users[i].PhoneNumber.Status = pgtype.Null
		users[i].PhoneNumber = database.Text("")
		users[i].PasswordSalt = passwordSalt
		users[i].PasswordHash = hashedPwd

		authUsers = append(authUsers, users[i])
	}

	result, err := authClient.ImportUsers(ctx, authUsers, authClient.GetHashConfig())
	if err != nil {
		return errors.Wrapf(err, "ImportUsers")
	}

	if len(result.UsersFailedToImport) > 0 {
		var errs []string
		for _, userFailedToImport := range result.UsersFailedToImport {
			errs = append(errs, fmt.Sprintf("%s - %s", userFailedToImport.User.GetEmail(), userFailedToImport.Err))
		}
		return status.Error(codes.InvalidArgument, fmt.Sprintf("create user in auth platform: %s", strings.Join(errs, ", ")))
	}
	return nil
}

func newCreateStudentEvents(schoolID int32, locations []*domain.Location, students ...*entity.LegacyStudent) []*pb.EvtUser {
	createStudentEvents := make([]*pb.EvtUser, 0, len(students))
	locationIDs := []string{}
	for _, location := range locations {
		locationIDs = append(locationIDs, location.LocationID)
	}

	for _, student := range students {
		createStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_CreateStudent_{
				CreateStudent: &pb.EvtUser_CreateStudent{
					StudentId:                student.ID.String,
					StudentName:              student.GetName(),
					SchoolId:                 strconv.FormatInt(int64(schoolID), 10),
					LocationIds:              locationIDs,
					StudentFirstName:         student.FirstName.String,
					StudentLastName:          student.LastName.String,
					StudentFirstNamePhonetic: student.FirstNamePhonetic.String,
					StudentLastNamePhonetic:  student.LastNamePhonetic.String,
				},
			},
		}
		createStudentEvents = append(createStudentEvents, createStudentEvent)
	}
	return createStudentEvents
}

func (s *UserModifierService) publishUserEvent(ctx context.Context, eventType string, userEvents ...*pb.EvtUser) error {
	for _, event := range userEvents {
		data, err := proto.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event %s error, %w", eventType, err)
		}
		_, err = s.JSM.TracedPublish(ctx, "publishUserEvent", eventType, data)
		if err != nil {
			return fmt.Errorf("publishUserEvent with %s: s.JSM.Publish failed: %w", eventType, err)
		}
	}

	return nil
}

func (s *UserModifierService) GetLocations(ctx context.Context, locationIDsReq []string) ([]*domain.Location, error) {
	if len(locationIDsReq) == 0 {
		return []*domain.Location{}, nil
	}

	locations := []*domain.Location{}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		locationIDs := make([]string, 0, len(locationIDsReq))
		for _, id := range locationIDsReq {
			if id == "" {
				return fmt.Errorf("getLocations invalid params: location_id empty")
			}
			locationIDs = append(locationIDs, id)
		}

		var ids pgtype.TextArray
		if err := ids.Set(locationIDs); err != nil {
			return errors.Wrap(err, "getLocations combine locationId fail")
		}

		locationsExisted, err := s.LocationRepo.GetLocationsByLocationIDs(ctx, s.DB, ids, false)
		if err != nil {
			return errors.Wrap(err, "getLocations fail")
		}
		if len(locationIDsReq) != len(locationsExisted) {
			return errors.Errorf("getLocations fail: expect GetLocationsByLocationIDs return %d locations, but got %d", len(locationIDsReq), len(locationsExisted))
		}

		resourcePath := golibs.ResourcePathFromCtx(ctx)
		for _, location := range locationsExisted {
			if location.ResourcePath != resourcePath {
				return errors.Errorf("getLocations fail: resource path invalid, expect %s, but actual %s", resourcePath, location.ResourcePath)
			}
		}

		locations = locationsExisted
		return nil
	}); err != nil {
		return locations, err
	}

	return locations, nil
}

func (s *UserModifierService) GetLocationsByPartnerInternalIDs(ctx context.Context, partnerLocationIDs []string) ([]*domain.Location, error) {
	if len(partnerLocationIDs) == 0 {
		return []*domain.Location{}, nil
	}

	locations := []*domain.Location{}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		locationsExisted, err := s.LocationRepo.GetLocationsByPartnerInternalIDs(ctx, tx, database.TextArray(partnerLocationIDs))
		if err != nil {
			return errors.Wrap(err, "GetLocationsByPartnerInternalIDs fail")
		}
		if len(partnerLocationIDs) != len(locationsExisted) {
			return errors.Errorf("GetLocationsByPartnerInternalIDs fail: expect GetLocationsByPartnerInternalIDs return %d locations, but got %d", len(partnerLocationIDs), len(locationsExisted))
		}

		resourcePath := golibs.ResourcePathFromCtx(ctx)
		for _, location := range locationsExisted {
			if location.ResourcePath != resourcePath {
				return errors.Errorf("GetLocationsByPartnerInternalIDs fail: resource path invalid, expect %s, but actual %s", resourcePath, location.ResourcePath)
			}
		}

		locations = locationsExisted
		return nil
	}); err != nil {
		return locations, err
	}

	return locations, nil
}

const (
	featureToggleExternalParentID = "User_StudentManagement_BackOffice_ExternalUserID_For_Parent"
)

func (s *UserModifierService) GenerateImportParentsAndAssignToStudentTemplate(ctx context.Context, req *pb.GenerateImportParentsAndAssignToStudentTemplateRequest) (*pb.GenerateImportParentsAndAssignToStudentTemplateResponse, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	csvTemplateHeaders := StableTemplateImportParentHeaders
	csvTemplateValues := StableTemplateImportParentValues

	featureUserNameStudentParentEnabled := unleash.IsFeatureUserNameStudentParentEnabled(s.UnleashClient, s.Env, organization)
	if featureUserNameStudentParentEnabled {
		csvTemplateHeaders, csvTemplateValues = prependBeforeColumn(
			csvTemplateHeaders, csvTemplateValues,
			"last_name",
			"username", "username",
		)
	}

	templateCSV := convertDataTemplateCSVToBase64(csvTemplateHeaders + "\n" + csvTemplateValues)
	return &pb.GenerateImportParentsAndAssignToStudentTemplateResponse{Data: []byte(templateCSV)}, nil
}

func UpsertUserAccessPath(ctx context.Context, userAccessPathRepo UserAccessPathRepo, db database.QueryExecer, locations []*domain.Location, userID string) error {
	userAccessPathEnts, err := toUserAccessPathEntities(locations, []string{userID})
	if err != nil {
		return errors.Wrap(err, "toUserAccessPathEntities")
	}

	if err := userAccessPathRepo.Upsert(ctx, db, userAccessPathEnts); err != nil {
		return errors.Wrap(err, "userAccessPathRepo.Upsert")
	}

	return nil
}

// Only used for students and parents. Get the student's ID to get all parent relationships with the student, and then give a new location to both.
func UpsertUserAccessPathForStudentParents(
	ctx context.Context,
	userAccessPathRepo UserAccessPathRepo,
	studentParent StudentParentRepo,
	db database.QueryExecer,
	locations []*domain.Location,
	studentID string,
) error {
	parentIDs, err := studentParent.FindParentIDsFromStudentID(ctx, db, studentID)
	if err != nil {
		return errors.Wrap(err, "studentParent.FindParentIDsFromStudentID")
	}

	parentIDs = append(parentIDs, studentID)
	userAccessPathEnts, err := toUserAccessPathEntities(locations, parentIDs)
	if err != nil {
		return errors.Wrap(err, "toUserAccessPathEntities")
	}

	if err := userAccessPathRepo.Upsert(ctx, db, userAccessPathEnts); err != nil {
		return errors.Wrap(err, "userAccessPathRepo.Upsert")
	}

	return nil
}

func (s *UserModifierService) UpsertTaggedUsers(ctx context.Context, db database.QueryExecer, userWithTags map[entity.User][]entity.DomainTag, existedTaggedUsers []entity.DomainTaggedUser) error {
	createTaggedUsers, deleteTaggedUsers, err := classifyTaggedUserParams(ctx, userWithTags, existedTaggedUsers)
	if err != nil {
		return errors.Wrap(err, "classifyTaggedUserParams")
	}

	if len(createTaggedUsers) > 0 {
		if err := s.DomainTaggedUserRepo.UpsertBatch(ctx, db, createTaggedUsers...); err != nil {
			return errors.Wrap(err, "DomainTaggedUserRepo.UpsertBatch")
		}
	}

	if len(deleteTaggedUsers) > 0 {
		if err := s.DomainTaggedUserRepo.SoftDelete(ctx, db, deleteTaggedUsers...); err != nil {
			return errors.Wrap(err, "DomainTaggedUserRepo.SoftDelete")
		}
	}
	return nil
}

// classifyTaggedUserParams from user and tag and existed user tags in db
//
//	which separate with user tag should be created
//	which separate with user tag should be deleted
func classifyTaggedUserParams(ctx context.Context, userWithTags map[entity.User][]entity.DomainTag, existedTaggedUsers []entity.DomainTaggedUser) ([]entity.DomainTaggedUser, []entity.DomainTaggedUser, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "OrganizationFromContext")
	}

	users := map[field.String]entity.User{}
	for user := range userWithTags {
		users[user.UserID()] = user
	}

	// user tags need create
	var createTaggedUsers []entity.DomainTaggedUser
	mapUserWithExistedTag := map[string]struct{}{}
	for _, tag := range existedTaggedUsers {
		key := tag.UserID().String() + tag.TagID().String()
		mapUserWithExistedTag[key] = struct{}{}
	}

	for user, tags := range userWithTags {
		for _, tag := range tags {
			key := user.UserID().String() + tag.TagID().String()
			if _, ok := mapUserWithExistedTag[key]; !ok {
				taggedUser := entity.DelegateToTaggedUser(user, tag, organization)
				createTaggedUsers = append(createTaggedUsers, taggedUser)
			}
		}
	}

	var deleteTaggedUsers []entity.DomainTaggedUser
	mapUserWithTags := map[string]struct{}{}
	for user, tags := range userWithTags {
		for _, tag := range tags {
			key := user.UserID().String() + tag.TagID().String()
			mapUserWithTags[key] = struct{}{}
		}
	}

	for _, tag := range existedTaggedUsers {
		key := tag.UserID().String() + tag.TagID().String()
		if _, ok := mapUserWithTags[key]; !ok {
			taggedUser := entity.DelegateToTaggedUser(users[tag.UserID()], tag, organization)
			deleteTaggedUsers = append(deleteTaggedUsers, taggedUser)
		}
	}

	return createTaggedUsers, deleteTaggedUsers, nil
}
