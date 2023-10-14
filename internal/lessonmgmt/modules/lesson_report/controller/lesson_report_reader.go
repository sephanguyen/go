package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LessonReportReaderService struct {
	wrapperConnection *support.WrapperDBConnection
	cfg               *configurations.Config
	configRepo        infrastructure.ConfigRepo
}

func NewLessonReportReaderService(
	wrapperConnection *support.WrapperDBConnection,
	cfg *configurations.Config,
	configRepo infrastructure.ConfigRepo,
) *LessonReportReaderService {
	return &LessonReportReaderService{
		wrapperConnection: wrapperConnection,
		cfg:               cfg,
		configRepo:        configRepo,
	}
}

func (s *LessonReportReaderService) GetConfigByDomainType(domainType lpb.DomainType, schoolID string) (configKey string, url string, err error) {
	switch domainType {
	case lpb.DomainType_DOMAIN_TYPE_BO:
		return fmt.Sprintf(`domain_%s_bo_%s`, s.cfg.Common.Environment, schoolID), s.cfg.Partner.DomainBo, nil
	case lpb.DomainType_DOMAIN_TYPE_TEACHER:
		return fmt.Sprintf(`domain_%s_teacher_%s`, s.cfg.Common.Environment, schoolID), s.cfg.Partner.DomainTeacher, nil
	case lpb.DomainType_DOMAIN_TYPE_LEARNER:
		return fmt.Sprintf(`domain_%s_learner_%s`, s.cfg.Common.Environment, schoolID), s.cfg.Partner.DomainLearner, nil
	default:
		return "", "", fmt.Errorf("LessonReportReaderService: invalid domain type")
	}
}

func (s *LessonReportReaderService) RetrievePartnerDomain(ctx context.Context, req *lpb.GetPartnerDomainRequest) (*lpb.GetPartnerDomainResponse, error) {
	if len(req.Type.String()) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("domain type must not be emptied").Error())
	}
	resourcePathID := golibs.ResourcePathFromCtx(ctx)
	if resourcePathID == "" {
		return nil, status.Error(codes.Internal, fmt.Errorf("Can not get resourcePathID").Error())
	}
	configKey, url, err := s.GetConfigByDomainType(req.Type, resourcePathID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("configKeyByDomainType: %w", err).Error())
	}
	conn, err := s.wrapperConnection.GetDB(resourcePathID)
	if err != nil {
		return nil, err
	}
	domains, err := s.configRepo.Retrieve(ctx, conn, database.Text(pb.COUNTRY_MASTER.String()), database.Text("lesson"), database.TextArray([]string{configKey}))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConfigRepo.Find: %w", err).Error())
	}
	if len(domains) > 0 {
		url = domains[0].Value.String
	}

	return &lpb.GetPartnerDomainResponse{
		Domain: url,
	}, nil
}
