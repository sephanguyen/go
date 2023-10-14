package service

import (
	"context"
	"time"

	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pbc "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/vmihailenco/taskq/v3"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StudentService struct {
	pb.UnimplementedStudentServiceServer
	DB                  database.Ext
	JSM                 nats.JetStreamManagement
	FirebaseAuthClient  internal_auth_tenant.TenantClient
	UnleashClient       unleashclient.ClientInstance
	ConfigurationClient clients.ConfigurationClientInterface
	Env                 string

	TaskQueue interface {
		Add(msg *taskq.Message) error
	}
	StudentCommentRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, comment *entity.StudentComment) error
		DeleteStudentComments(ctx context.Context, db database.QueryExecer, cmtIDs []string) error
		RetrieveByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, fields ...string) ([]entity.StudentComment, error)
	}
	StudentRepo interface {
		Retrieve(context.Context, database.QueryExecer, pgtype.TextArray) ([]repository.StudentProfile, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entity.LegacyStudent) error
	}
	UserRepo interface {
		Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.LegacyUser, error)
		GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entity.LegacyUser, error)
		GetByEmailInsensitiveCase(ctx context.Context, db database.QueryExecer, emails []string) ([]*entity.LegacyUser, error)
		GetByPhone(ctx context.Context, db database.QueryExecer, phones pgtype.TextArray) ([]*entity.LegacyUser, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser) error
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entity.LegacyUser, error)
	}
	UsrEmailRepo interface {
		Create(ctx context.Context, db database.QueryExecer, usrID pgtype.Text, email pgtype.Text) (*entity.UsrEmail, error)
		CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser) ([]*entity.UsrEmail, error)
	}
	UserGroupRepo interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entity.UserGroup) error
	}
	UserGroupV2Repo interface {
		FindUserGroupByRoleName(ctx context.Context, db database.QueryExecer, roleName string) (*entity.UserGroupV2, error)
	}
	UserGroupsMemberRepo interface {
		AssignWithUserGroup(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser, userGroupID pgtype.Text) error
	}
	OrganizationRepo interface {
		GetTenantIDByOrgID(ctx context.Context, db database.QueryExecer, orgID string) (string, error)
	}
	UserAccessPathRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userAccessPaths []*entity.UserAccessPath) error
	}
	ImportUserEventRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, importUserEvents []*entity.ImportUserEvent) ([]*entity.ImportUserEvent, error)
	}
	UserModifierService interface {
		GetLocations(ctx context.Context, locationIDsReq []string) ([]*domain.Location, error)
		CreateUsersInIdentityPlatform(ctx context.Context, tenantID string, users []*entity.LegacyUser, resourcePath int64) error
		GetLocationsByPartnerInternalIDs(ctx context.Context, partnerInternalIDs []string) ([]*domain.Location, error)
		GetGradeMaster(ctx context.Context, gradeID string) (map[entity.DomainGrade]field.Int32, error)
		UpsertTaggedUsers(ctx context.Context, db database.QueryExecer, userWithTags map[entity.User][]entity.DomainTag, existedTaggedUsers []entity.DomainTaggedUser) error
	}
	GradeOrganizationRepo interface {
		GetByGradeValues(ctx context.Context, db database.QueryExecer, gradeValues []int32) ([]*repository.GradeOrganization, error)
	}
	SchoolInfoRepo interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.SchoolInfo, error)
		GetBySchoolPartnerIDs(ctx context.Context, db database.QueryExecer, schoolPartnerIds pgtype.TextArray) ([]*entity.SchoolInfo, error)
	}

	SchoolCourseRepo interface {
		GetByIDsAndSchoolIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, schoolIDs pgtype.TextArray) ([]*entity.SchoolCourse, error)
		GetBySchoolCoursePartnerIDsAndSchoolIDs(ctx context.Context, db database.QueryExecer, schoolCoursePartnerIds pgtype.TextArray, schoolIDs pgtype.TextArray) ([]*entity.SchoolCourse, error)
	}
	UserAddressRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userAddresses []*entity.UserAddress) error
		SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) error
		GetByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserAddress, error)
	}
	SchoolHistoryRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, schoolHistories []*entity.SchoolHistory) error
		GetSchoolHistoriesByGradeIDAndStudentID(ctx context.Context, db database.QueryExecer, gradeID pgtype.Text, studentID pgtype.Text, isCurrent pgtype.Bool) ([]*entity.SchoolHistory, error)
		SetCurrentSchoolByStudentIDAndSchoolID(ctx context.Context, db database.QueryExecer, schoolID pgtype.Text, studentID pgtype.Text) error
	}
	PrefectureRepo interface {
		GetByPrefectureID(ctx context.Context, db database.QueryExecer, prefectureID pgtype.Text) (*entity.Prefecture, error)
		GetByPrefectureCode(ctx context.Context, db database.QueryExecer, prefectureCode pgtype.Text) (*entity.Prefecture, error)
		GetByPrefectureIDs(ctx context.Context, db database.QueryExecer, prefectureIDs pgtype.TextArray) ([]*entity.Prefecture, error)
	}
	UserPhoneNumberRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userPhoneNumbers []*entity.UserPhoneNumber) error
	}
	DomainLocationRepo interface {
		GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) (entity.DomainLocations, error)
	}
	DomainStudentService interface {
		ValidateUpdateSystemAndExternalUserID(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error
		UpsertMultiple(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToCreate ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error)
		GetEmailWithStudentID(ctx context.Context, studentIDs []string) (map[string]entity.User, error)
		GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error)
		GetGradesByExternalIDs(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error)
		GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error)
		GetLocationsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainLocations, error)
		GetSchoolsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchools, error)
		GetSchoolCoursesByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error)
		GetPrefecturesByCodes(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error)
		IsFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled(organization valueobj.HasOrganizationID) bool
		UpsertMultipleWithErrorCollection(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error)
		IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool
		IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization valueobj.HasOrganizationID) bool
		IsDisableAutoDeactivateStudents(organization valueobj.HasOrganizationID) bool
		IsExperimentalBulkInsertEnrollmentStatusHistories(organization valueobj.HasOrganizationID) bool
		IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error)
	}
	FeatureManager interface {
		FeatureUsernameToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
	}
	DomainTagRepo               DomainTagRepo
	EnrollmentStatusHistoryRepo DomainEnrollmentStatusHistoryRepo
	DomainUserAccessPathRepo    DomainUserAccessPathRepo
}

func (s *StudentService) UpsertStudentComment(ctx context.Context, req *pb.UpsertStudentCommentRequest) (*pb.UpsertStudentCommentResponse, error) {
	req.StudentComment.CoachId = interceptors.UserIDFromContext(ctx)
	studentCommentModel, err := toStudentCommentEntity(req.StudentComment)
	if err != nil {
		return nil, err
	}

	err = s.StudentCommentRepo.Upsert(ctx, s.DB, studentCommentModel)
	if err != nil {
		return nil, err
	}

	return &pb.UpsertStudentCommentResponse{
		Successful: true,
	}, nil
}

func (s *StudentService) DeleteStudentComments(ctx context.Context, req *pb.DeleteStudentCommentsRequest) (*pb.DeleteStudentCommentsResponse, error) {
	if req.CommentIds == nil {
		return nil, status.Error(codes.InvalidArgument, "comment ids must not nil")
	}
	if len(req.CommentIds) == 0 {
		return &pb.DeleteStudentCommentsResponse{
			Successful: true,
		}, nil
	}

	err := s.StudentCommentRepo.DeleteStudentComments(ctx, s.DB, req.CommentIds)
	if err != nil {
		switch err.Error() {
		case repository.ErrUnAffected.Error():
			return &pb.DeleteStudentCommentsResponse{
				Successful: false,
			}, nil
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &pb.DeleteStudentCommentsResponse{
		Successful: true,
	}, nil
}

const (
	featureToggleBulkUpdateStudentCSV = "User_StudentManagement_BulkUpdateStudentCSV"
)

func (s *StudentService) GenerateImportStudentTemplate(ctx context.Context, _ *pb.GenerateImportStudentTemplateRequest) (*pb.GenerateImportStudentTemplateResponse, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	templateCSVHeader := StableTemplateImportStudentHeaders
	templateCSVValue := StableTemplateImportStudentValues

	featureUserNameStudentParentEnabled := unleash.IsFeatureUserNameStudentParentEnabled(s.UnleashClient, s.Env, organization)
	if featureUserNameStudentParentEnabled {
		templateCSVHeader, templateCSVValue = prependBeforeColumn(
			templateCSVHeader, templateCSVValue,
			"last_name",
			"username", "username",
		)
	}

	templateCSV := templateCSVHeader + "\n" + templateCSVValue
	return &pb.GenerateImportStudentTemplateResponse{
		Data: []byte(convertDataTemplateCSVToBase64(templateCSV)),
	}, nil
}

func (s *StudentService) RetrieveStudentComment(ctx context.Context, req *pb.RetrieveStudentCommentRequest) (*pb.RetrieveStudentCommentResponse, error) {
	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, "studentId is cannot empty or nil")
	}
	commentListModel, err := s.StudentCommentRepo.RetrieveByStudentID(ctx, s.DB, database.Text(req.StudentId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "student id invalid")
	}
	comments := make([]*pb.CommentInfo, 0, len(commentListModel))

	for idx := range commentListModel {
		studentComment := toStudentCommentPb(&commentListModel[idx])
		coaches, err := s.UserRepo.Retrieve(ctx, s.DB, database.TextArray([]string{studentComment.CoachId}))
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrap(err, "cannot find coach").Error())
		}
		if len(coaches) == 0 {
			return nil, status.Error(codes.InvalidArgument, "the coach id is non-existing")
		}
		comments = append(comments, &pb.CommentInfo{
			CoachName:      coaches[0].GetName(),
			StudentComment: studentComment,
		})
	}
	return &pb.RetrieveStudentCommentResponse{
		Comment: comments,
	}, nil
}

func toStudentCommentPb(c *entity.StudentComment) *pb.StudentComment {
	return &pb.StudentComment{
		CommentId:      c.CommentID.String,
		StudentId:      c.StudentID.String,
		CommentContent: c.CommentContent.String,
		CoachId:        c.CoachID.String,
		UpdatedAt:      &timestamppb.Timestamp{Seconds: c.UpdatedAt.Time.Unix()},
		CreatedAt:      &timestamppb.Timestamp{Seconds: c.UpdatedAt.Time.Unix()},
	}
}

func toStudentCommentEntity(src *pb.StudentComment) (*entity.StudentComment, error) {
	if src.CommentId == "" {
		src.CommentId = idutil.ULIDNow()
	}
	e := new(entity.StudentComment)
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.CommentID.Set(src.CommentId),
		e.StudentID.Set(src.StudentId),
		e.CoachID.Set(src.CoachId),
		e.CommentContent.Set(src.CommentContent),
		e.UpdatedAt.Set(time.Now()),
		e.CreatedAt.Set(time.Now()),
	); err != nil {
		return nil, err
	}

	return e, nil
}

func (s *StudentService) GetStudentProfile(ctx context.Context, req *pb.GetStudentProfileRequest) (*pb.GetStudentProfileResponse, error) {
	if n := len(req.StudentIds); n > 200 {
		return nil, status.Error(codes.InvalidArgument, "number of ID in validStudentIDs request must be less than 200")
	} else if n == 0 {
		req.StudentIds = []string{interceptors.UserIDFromContext(ctx)}
	}

	dbProfiles, err := s.StudentRepo.Retrieve(ctx, s.DB, database.TextArray(req.StudentIds))
	if err != nil {
		return nil, toStatusError(err)
	}

	userID := interceptors.UserIDFromContext(ctx)
	outProfiles := make([]*pb.StudentProfile, len(dbProfiles))

	for i, dbProfile := range dbProfiles {
		outProfile := studentToBasicProfile(dbProfile.Student)

		isProfileOwner := dbProfile.Student.ID.String == userID
		if isProfileOwner {
			outProfile = studentToProfile(dbProfile.Student)
		}

		if dbProfile.Student.SchoolID.Status == pgtype.Present && isProfileOwner {
			outProfile.School = toSchoolPb(&dbProfile.School)
		}

		outProfile.GradeName = dbProfile.Grade.Name.String

		outProfiles[i] = outProfile
	}

	return &pb.GetStudentProfileResponse{Profiles: outProfiles}, nil
}

func (s *StudentService) UpsertStudent(ctx context.Context, req *pb.UpsertStudentRequest) (*pb.UpsertStudentResponse, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, errorx.GRPCErr(err, nil)
	}

	option := unleash.DomainStudentFeatureOption{}
	option = s.FeatureManager.FeatureUsernameToStudentFeatureOption(ctx, organization, option)
	option = s.FeatureManager.FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx, organization, option)
	option = s.FeatureManager.FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(ctx, organization, option)
	option = s.FeatureManager.FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(ctx, organization, option)

	studentProfile := grpc.ToDomainStudents(req.GetStudentProfiles(), option.EnableUsername)

	students, err := s.DomainStudentService.UpsertMultiple(ctx, option, studentProfile...)
	if err != nil {
		switch e := err.(type) {
		case errcode.Error:
			return nil, errorx.GRPCErr(err, errorx.PbErrorMessage(e))
		case errcode.DomainError:
			return nil, errorx.GRPCErr(e, grpc.ToPbErrorMessageBackOffice(e))
		}
	}
	return &pb.UpsertStudentResponse{StudentProfiles: grpc.UpsertStudentProfiles(students)}, nil
}

func studentToBasicProfile(student entity.LegacyStudent) *pb.StudentProfile {
	return &pb.StudentProfile{
		Id:     student.ID.String,
		Name:   student.FullName.String,
		Avatar: student.Avatar.String,
	}
}

func studentToProfile(student entity.LegacyStudent) *pb.StudentProfile {
	birthDay := timestamppb.New(student.Birthday.Time)
	createdAt := timestamppb.New(student.CreatedAt.Time)
	country := pbc.Country(pbc.Country_value[student.Country.String])
	grade, _ := i18n.ConvertIntGradeToStringV1(country, int(student.CurrentGrade.Int))

	var divs []int64
	data, _ := student.GetStudentAdditionalData()
	if data != nil {
		divs = data.JprefDivs
	}

	return &pb.StudentProfile{
		Id:        student.ID.String,
		Name:      student.FullName.String,
		Country:   country,
		Phone:     student.PhoneNumber.String,
		Email:     student.Email.String,
		Grade:     grade,
		Avatar:    student.Avatar.String,
		Birthday:  birthDay,
		CreatedAt: createdAt,
		Divs:      divs,
	}
}
