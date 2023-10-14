package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LessonReportReaderService struct {
	DB              database.Ext
	Cfg             *configurations.Config
	SchoolAdminRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities.SchoolAdmin, error)
	}
	UserRepo interface {
		UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
	}
	TeacherRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Teacher, error)
	}
	ConfigRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, country pgtype.Text, group pgtype.Text, keys pgtype.TextArray) ([]*entities.Config, error)
	}
	StudentRepo interface {
		Find(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Student, error)
	}
}

func (s *LessonReportReaderService) getSchoolIdFromContext(ctx context.Context) (int32, error) {
	userID := interceptors.UserIDFromContext(ctx)
	group, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(userID))
	if err != nil {
		return 0, status.Error(codes.Internal, fmt.Errorf("UserRepo.UserGroup: %w", err).Error())
	}

	switch group {
	case constant.UserGroupTeacher:
		teacher, err := s.TeacherRepo.FindByID(ctx, s.DB, database.Text(userID))
		if err != nil {
			return 0, status.Error(codes.Internal, fmt.Errorf("TeacherRepo.FindByID: %w", err).Error())
		}
		if teacher.SchoolIDs.Status != pgtype.Present || len(teacher.SchoolIDs.Elements) == 0 {
			return 0, status.Error(codes.Internal, fmt.Errorf("can't detect school of teacher").Error())
		}
		return teacher.SchoolIDs.Elements[0].Int, nil
	case constant.UserGroupSchoolAdmin:
		schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DB, database.Text(userID))
		if err != nil {
			return 0, status.Error(codes.Internal, fmt.Errorf("SchoolAdminRepo.Get: %w", err).Error())
		}
		return schoolAdmin.SchoolID.Int, nil
	case constant.UserGroupAdmin:
		return 0, nil
	case constant.UserGroupStudent:
		student, err := s.StudentRepo.Find(ctx, s.DB, database.Text(userID))
		if err != nil {
			return 0, status.Error(codes.Internal, fmt.Errorf("StudentRepo.FindByID: %w", err).Error())
		}
		return student.SchoolID.Int, nil
	default:
		return 0, status.Error(codes.Internal, fmt.Errorf("unknown user's group").Error())
	}
}

func (s *LessonReportReaderService) RetrievePartnerDomain(ctx context.Context, req *bpb.GetPartnerDomainRequest) (*bpb.GetPartnerDomainResponse, error) {
	if len(req.Type.String()) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("domain type must not be emptied").Error())
	}

	schoolId, err := s.getSchoolIdFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("getSchoolIdFromContext.Get: %w", err).Error())
	}
	if schoolId == 0 {
		return nil, status.Error(codes.Internal, fmt.Errorf("Can not get schoolID").Error())
	}
	configKey, url, err := s.ConfigByDomainType(req.Type, schoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("configKeyByDomainType: %w", err).Error())
	}
	// Currently, Configs use for jprep staging only. It use staging manabie, but different domain.
	domains, err := s.ConfigRepo.Retrieve(ctx, s.DB, database.Text(pb.COUNTRY_MASTER.String()), database.Text("lesson"), database.TextArray([]string{configKey}))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConfigRepo.Find: %w", err).Error())
	}
	if len(domains) > 0 {
		url = domains[0].Value.String
	}

	return &bpb.GetPartnerDomainResponse{
		Domain: url,
	}, nil
}

func (s *LessonReportReaderService) ConfigByDomainType(domainType bpb.DomainType, schoolID int32) (configKey string, url string, err error) {
	switch domainType {
	case bpb.DomainType_DOMAIN_TYPE_BO:
		return fmt.Sprintf(`domain_%s_bo_%d`, s.Cfg.Common.Environment, uint32(schoolID)), s.Cfg.Partner.DomainBo, nil
	case bpb.DomainType_DOMAIN_TYPE_TEACHER:
		return fmt.Sprintf(`domain_%s_teacher_%d`, s.Cfg.Common.Environment, uint32(schoolID)), s.Cfg.Partner.DomainTeacher, nil
	case bpb.DomainType_DOMAIN_TYPE_LEARNER:
		return fmt.Sprintf(`domain_%s_learner_%d`, s.Cfg.Common.Environment, uint32(schoolID)), s.Cfg.Partner.DomainLearner, nil
	default:
		return "", "", fmt.Errorf("LessonReportReaderService: invalid domain type")
	}
}
