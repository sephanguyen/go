package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserMgmtService interface {
	SearchBasicProfile(ctx context.Context, in *upb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error)
}

type StudentLearningTimeDaiyRepo interface {
	Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
	RetrieveV2(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*repositories.StudentLearningTimeDailyV2, error)
}

type StudentLearningTimeReaderService struct {
	DB                          database.Ext
	UserMgmtService             UserMgmtService
	StudentLearningTimeDaiyRepo StudentLearningTimeDaiyRepo
}

func (s *StudentLearningTimeReaderService) normalizeTimeRetrieveLearningProgress(ctx context.Context, req *pb.RetrieveLearningProgressRequest) (*pgtype.Timestamptz, *pgtype.Timestamptz, error) {
	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	resp, err := s.UserMgmtService.SearchBasicProfile(mdCtx, &upb.SearchBasicProfileRequest{
		UserIds: []string{req.StudentId},
		Paging: &cpb.Paging{
			Limit: 1,
		},
	})
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "RetrieveUserProfile: %v", err.Error())
	}
	if len(resp.Profiles) == 0 {
		return nil, nil, status.Error(codes.NotFound, codes.NotFound.String())
	}
	country := pb.Country(resp.Profiles[0].Country)

	// For this api, the client should send:
	//  - req.From is Monday 00:00:00 on student's local time
	//  - req.To is Sunday 23:59:59 on student's local time
	// and because both req.From and req.To are using google.protobuf.Timestamp,
	// so if the student's country is VN, which is UTC+07, the expected data are:
	//  - req.From is Sunday 17:00:00 +00
	//  - req.To is next Sunday 16:59:59 +00
	//  but currently the data client send are:
	//  - req.From is Monday 00:00:00 +00
	//  - req.To is Sunday 23:59:59 +00
	// so we must change the req.From and req.To to match with expected data above.
	// TODO: remove this when the client send same with expected data.
	tFrom := req.From.AsTime()
	if tFrom.Hour() == 0 && tFrom.Minute() == 0 && tFrom.Second() == 0 {
		if country == pb.Country(cpb.Country_COUNTRY_VN) {
			tFrom = tFrom.Add(-7 * time.Hour)
		}
	}
	from := new(pgtype.Timestamptz)
	from.Set(tFrom)

	tTo := req.To.AsTime()
	if tTo.Hour() == 23 && tTo.Minute() == 59 && tTo.Second() == 59 {
		if country == pb.Country(cpb.Country_COUNTRY_VN) {
			tTo = tTo.Add(-7 * time.Hour)
		}
	}
	to := new(pgtype.Timestamptz)
	to.Set(tTo)
	return from, to, nil
}

func validateRetrieveLearningProgressRequest(ctx context.Context, req *pb.RetrieveLearningProgressRequest) error {
	if !dateRangeValid(req.From, req.To) {
		return status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	if !canProcessStudentData(ctx, req.StudentId) {
		return status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return nil
}

func (s *StudentLearningTimeReaderService) RetrieveLearningProgress(ctx context.Context, req *pb.RetrieveLearningProgressRequest) (*pb.RetrieveLearningProgressResponse, error) {
	if err := validateRetrieveLearningProgressRequest(ctx, req); err != nil {
		return nil, err
	}
	from, to, err := s.normalizeTimeRetrieveLearningProgress(ctx, req)
	if err != nil {
		return nil, err
	}

	learningTimeByDailies, err := s.StudentLearningTimeDaiyRepo.Retrieve(ctx, s.DB, database.Text(req.StudentId), from, to)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.StudentLearningTimeDaiyRepo.Retrieve %w", err).Error())
	}

	var ret []*pb.RetrieveLearningProgressResponse_DailyLearningTime
	for from.Time.Before(to.Time) {
		lt := &pb.RetrieveLearningProgressResponse_DailyLearningTime{
			Day: timestamppb.New(from.Time),
		}
		for _, d := range learningTimeByDailies {
			if d.Day.Time.Equal(from.Time) {
				lt.TotalTimeSpentInDay = int64(d.LearningTime.Int)
				break
			}
		}
		ret = append(ret, lt)
		from.Time = from.Time.Add(24 * time.Hour)
	}

	return &pb.RetrieveLearningProgressResponse{Dailies: ret}, nil
}

func dateRangeValid(from, to *timestamppb.Timestamp) bool {
	if from == nil || to == nil {
		return false
	}
	return from.AsTime().Before(to.AsTime())
}

func canProcessStudentData(ctx context.Context, studentID string) bool {
	currentUserID := interceptors.UserIDFromContext(ctx)
	uGroup := interceptors.UserGroupFromContext(ctx)

	switch uGroup {
	case cpb.UserGroup_USER_GROUP_STUDENT.String(), constant.RoleStudent:
		// only allows student get his owned plans
		return currentUserID == studentID
	default:
		return true
	}
}
