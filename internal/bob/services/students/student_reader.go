package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	bsvc "github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IStudentRepository interface {
	GetStudentsByParentID(ctx context.Context, db database.QueryExecer, parentID pgtype.Text) ([]*entities.User, error)
	Retrieve(context.Context, database.QueryExecer, pgtype.TextArray) ([]repositories.StudentProfile, error)
}

type StudentReaderServiceClient interface {
	FindStudent(context.Context, *bpb.FindStudentRequest) (*bpb.FindStudentResponse, error)
	RetrieveStudentProfile(context.Context, *bpb.RetrieveStudentProfileRequest) (*bpb.RetrieveStudentProfileResponse, error)
	RetrieveLearningProgress(context.Context, *bpb.RetrieveLearningProgressRequest) (*bpb.RetrieveLearningProgressResponse, error)
	RetrieveStat(context.Context, *bpb.RetrieveStatRequest) (*bpb.RetrieveStatResponse, error)
	RetrieveStudentAssociatedToParentAccount(context.Context, *bpb.RetrieveStudentAssociatedToParentAccountRequest) (*bpb.RetrieveStudentAssociatedToParentAccountResponse, error)
	GetListSchoolIDsByStudentIDs(context.Context, *bpb.GetListSchoolIDsByStudentIDsRequest) (*bpb.GetListSchoolIDsByStudentIDsResponse, error)
	RetrieveStudentSchoolHistory(context.Context, *bpb.RetrieveStudentSchoolHistoryRequest) (*bpb.RetrieveStudentSchoolHistoryResponse, error)
}

type StudentReaderService struct {
	studentRepository    IStudentRepository
	db                   database.Ext
	UserReaderStudentSvc bsvc.UserReaderStudentSvc

	SchoolHistoryRepository interface {
		GetCurrentSchoolInfoByStudentIDs(context.Context, database.QueryExecer, pgtype.TextArray) ([]*repositories.StudentSchoolInfo, error)
	}
}

func NewStudentReaderService(
	db database.Ext,
	userReaderStudentSvc bsvc.UserReaderStudentSvc,
) bpb.StudentReaderServiceServer {
	return &StudentReaderService{
		studentRepository:       &repositories.StudentRepo{},
		db:                      db,
		UserReaderStudentSvc:    userReaderStudentSvc,
		SchoolHistoryRepository: &repositories.SchoolHistoryRepo{},
	}
}

const (
	emailAlreadyExist = "emailAlreadyExist"
	phoneAlreadyExist = "phoneAlreadyExist"
)

var (
	emailAlreadyExistMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        emailAlreadyExist,
				Subject:     "registration",
				Description: "email already exist",
			},
		},
	}

	phoneAlreadyExistMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        phoneAlreadyExist,
				Subject:     "registration",
				Description: "phone number already exist",
			},
		},
	}
)

func (s *StudentReaderService) FindStudent(ctx context.Context, in *bpb.FindStudentRequest) (*bpb.FindStudentResponse, error) {
	return nil, status.Error(codes.Unimplemented, fmt.Sprintln("Method has not implemented yet"))
}

func (s *StudentReaderService) RetrieveStudentProfile(ctx context.Context, in *bpb.RetrieveStudentProfileRequest) (*bpb.RetrieveStudentProfileResponse, error) {
	if len(in.StudentIds) == 0 {
		in.StudentIds = []string{interceptors.UserIDFromContext(ctx)}
	}

	if n := len(in.StudentIds); n > 200 {
		return nil, status.Error(codes.InvalidArgument, "number of ID in validStudentIDsrequest must be less than 200")
	}
	students, err := s.studentRepository.Retrieve(ctx, s.db, database.TextArray(in.StudentIds))
	if err != nil {
		return nil, toStatusError(err)
	}

	resp := make([]*bpb.RetrieveStudentProfileResponse_Data, len(students))

	for i, s := range students {
		profile := student2Profile(&s.Student)

		if s.Student.SchoolID.Status == pgtype.Present {
			profile.School = toSchoolPb(&s.School)
		}

		resp[i] = &bpb.RetrieveStudentProfileResponse_Data{
			Profile: profile,
		}
	}
	return &bpb.RetrieveStudentProfileResponse{
		Items: resp,
	}, nil
}

func (s *StudentReaderService) GetListSchoolIDsByStudentIDs(ctx context.Context, in *bpb.GetListSchoolIDsByStudentIDsRequest) (*bpb.GetListSchoolIDsByStudentIDsResponse, error) {
	students, err := s.studentRepository.Retrieve(ctx, s.db, database.TextArray(in.StudentIds))
	if err != nil {
		return nil, toStatusError(err)
	}
	schoolIds := []*bpb.SchoolIDWithStudentIDs{}
	mapSchoolStudentIds := make(map[string][]string)
	for _, student := range students {
		schoolIDStr := fmt.Sprint(student.School.ID.Int)
		if mapStudentIDs, ok := mapSchoolStudentIds[schoolIDStr]; ok {
			mapStudentIDs = append(mapStudentIDs, student.Student.ID.String)
			mapSchoolStudentIds[schoolIDStr] = mapStudentIDs
		} else {
			mapSchoolStudentIds[schoolIDStr] = []string{student.Student.ID.String}
		}
	}
	for schoolID, studentIDs := range mapSchoolStudentIds {
		schoolIds = append(schoolIds, &bpb.SchoolIDWithStudentIDs{SchoolId: schoolID, StudentIds: studentIDs})
	}

	return &bpb.GetListSchoolIDsByStudentIDsResponse{SchoolIds: schoolIds}, nil
}

func toStatusError(err error) error {
	switch e := errors.Cause(err).(type) {
	case *pgconn.PgError:
		switch e.Code {
		case pgerrcode.ForeignKeyViolation: // foreign_key_violation
			return status.Error(codes.InvalidArgument, e.Message)
		case pgerrcode.UniqueViolation: // unique_violation
			stt := status.New(codes.AlreadyExists, e.Message)
			if e.ConstraintName == "users_phone_un" {
				stt, _ = stt.WithDetails(phoneAlreadyExistMsg)
			} else if e.ConstraintName == "users_email_un" {
				stt, _ = stt.WithDetails(emailAlreadyExistMsg)
			}

			return stt.Err()
		}
	}

	return status.Convert(err).Err()
}

func toSchoolPb(se *entities.School) *bpb.School {
	s := &bpb.School{
		Id:      se.ID.Int,
		Name:    se.Name.String,
		Country: cpb.Country(cpb.Country_value[se.Country.String]),
	}
	if se.City == nil {
		s.City = &bpb.City{
			Id: se.CityID.Int,
		}
	} else {
		s.City = toCityPb(se.City)
	}
	if se.District == nil {
		s.District = &bpb.District{
			Id: se.DistrictID.Int,
		}
	} else {
		s.District = toDistrictPb(se.District)
	}
	if se.Point.Status == pgtype.Present {
		s.Point = &bpb.Point{
			Lat:  se.Point.P.X,
			Long: se.Point.P.Y,
		}
	}
	return s
}

func toCityPb(ce *entities.City) *bpb.City {
	if ce == nil {
		return nil
	}
	c := &bpb.City{
		Id:      ce.ID.Int,
		Name:    ce.Name.String,
		Country: cpb.Country(cpb.Country_value[ce.Country.String]),
	}
	return c
}

func toDistrictPb(de *entities.District) *bpb.District {
	if de == nil {
		return nil
	}
	d := &bpb.District{
		Id:      de.ID.Int,
		Name:    de.Name.String,
		Country: cpb.Country(cpb.Country_value[de.Country.String]),
	}
	if de.City == nil {
		d.City = &bpb.City{
			Id: de.CityID.Int,
		}
	} else {
		d.City = toCityPb(de.City)
	}
	return d
}

func student2Profile(student *entities.Student) *bpb.StudentProfile {
	birthDay := timestamppb.New(student.Birthday.Time)
	createdAt := timestamppb.New(student.CreatedAt.Time)
	country := cpb.Country(cpb.Country_value[student.Country.String])
	grade, _ := i18n.ConvertIntGradeToStringV1(country, int(student.CurrentGrade.Int))

	var divs []int64
	data, _ := student.GetStudentAdditionalData()
	if data != nil {
		divs = data.JprefDivs
	}

	studentProfile := &bpb.StudentProfile{
		Id:               student.ID.String,
		Name:             student.GetName(),
		FullNamePhonetic: student.FullNamePhonetic.String,
		Country:          country,
		Phone:            student.PhoneNumber.String,
		Email:            student.Email.String,
		Grade:            grade,
		TargetUniversity: student.TargetUniversity.String,
		Avatar:           student.Avatar.String,
		Birthday:         birthDay,
		Biography:        student.Biography.String,
		CreatedAt:        createdAt,
		IsTester:         student.IsTester.Bool,
		FacebookId:       student.FacebookID.String,
		Divs:             divs,
	}
	if !student.LastLoginDate.Time.IsZero() {
		studentProfile.LastLoginDate = timestamppb.New(student.LastLoginDate.Time)
	}
	return studentProfile
}

func (s *StudentReaderService) RetrieveLearningProgress(ctx context.Context, in *bpb.RetrieveLearningProgressRequest) (*bpb.RetrieveLearningProgressResponse, error) {
	return nil, status.Error(codes.Unimplemented, fmt.Sprintln("Method has not implemented yet"))
}

func (s *StudentReaderService) RetrieveStat(ctx context.Context, in *bpb.RetrieveStatRequest) (*bpb.RetrieveStatResponse, error) {
	return nil, status.Error(codes.Unimplemented, fmt.Sprintln("Method has not implemented yet"))
}

func (s *StudentReaderService) RetrieveStudentAssociatedToParentAccount(ctx context.Context, in *bpb.RetrieveStudentAssociatedToParentAccountRequest) (*bpb.RetrieveStudentAssociatedToParentAccountResponse, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}

	resp, err := s.UserReaderStudentSvc.RetrieveStudentAssociatedToParentAccount(
		metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token),
		&upb.RetrieveStudentAssociatedToParentAccountRequest{},
	)

	if err != nil {
		return nil, fmt.Errorf("[StudentReaderService]:[retrieve student associated to parent account]:%v", err)
	}

	return &bpb.RetrieveStudentAssociatedToParentAccountResponse{
		Profiles: resp.Profiles,
	}, nil
}

func (s *StudentReaderService) RetrieveStudentSchoolHistory(ctx context.Context, req *bpb.RetrieveStudentSchoolHistoryRequest) (*bpb.RetrieveStudentSchoolHistoryResponse, error) {
	studentInfos, err := s.SchoolHistoryRepository.GetCurrentSchoolInfoByStudentIDs(ctx, s.db, database.TextArray(req.GetStudentIds()))
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "s.SchoolHistoryRepository.GetCurrentSchoolInfoByStudentIDs ").Error())
	}

	schools := make(map[string]*bpb.RetrieveStudentSchoolHistoryResponse_School, 0)
	for _, si := range studentInfos {
		if school, ok := schools[si.SchoolID.String]; ok {
			school.StudentIds = append(school.StudentIds, si.StudentID.String)
		} else {
			schools[si.SchoolID.String] = &bpb.RetrieveStudentSchoolHistoryResponse_School{
				SchoolId:   si.SchoolID.String,
				SchoolName: si.SchoolName.String,
				StudentIds: []string{si.StudentID.String},
			}
		}
	}

	return &bpb.RetrieveStudentSchoolHistoryResponse{
		Schools: schools,
	}, nil
}
