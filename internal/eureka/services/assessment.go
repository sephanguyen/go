package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	lrni "github.com/manabie-com/backend/internal/golibs/learnosity/init"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AssessmentService struct {
	sspb.UnimplementedAssessmentServer
	DB               database.Ext
	LearnosityConfig *configurations.LearnosityConfig
}

func NewAssessmentService(db database.Ext, learnosityConfig *configurations.LearnosityConfig) sspb.AssessmentServer {
	return &AssessmentService{
		DB:               db,
		LearnosityConfig: learnosityConfig,
	}
}

func (s *AssessmentService) generateLearnositySecurity(ctx context.Context, domain string, timestamp time.Time) learnosity.Security {
	return learnosity.Security{
		ConsumerKey:    s.LearnosityConfig.ConsumerKey,
		Domain:         domain,
		Timestamp:      learnosity.FormatUTCTime(timestamp),
		UserID:         interceptors.UserIDFromContext(ctx),
		ConsumerSecret: s.LearnosityConfig.ConsumerSecret,
	}
}

func (s *AssessmentService) validateGetSignedRequest(req *sspb.GetSignedRequestRequest) error {
	if req.RequestData == "" {
		return errors.New("req must have RequestData")
	}

	return nil
}

func (s *AssessmentService) GetSignedRequest(ctx context.Context, req *sspb.GetSignedRequestRequest) (*sspb.GetSignedRequestResponse, error) {
	if err := s.validateGetSignedRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateGetSignedRequest: %w", err).Error())
	}

	domain := req.Domain

	// Get origin domain from request header
	if domain == "" {
		md, valid := metadata.FromIncomingContext(ctx)
		if !valid {
			return nil, status.Error(codes.Unknown, "can't get metadata info from incoming context")
		}

		values := md.Get("origin")
		if len(values) > 0 {
			domain = strings.TrimPrefix(values[0], "https://")
		}
	}

	now := time.Now()
	security := s.generateLearnositySecurity(ctx, domain, now)

	init := lrni.New(learnosity.ServiceItems, security, learnosity.RequestString(req.RequestData))

	signedRequest, err := init.Generate(true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("init.Generate: %w", err).Error())
	}

	return &sspb.GetSignedRequestResponse{
		SignedRequest: signedRequest.(string),
	}, nil
}
