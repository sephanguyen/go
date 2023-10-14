package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStudentLearningTimeReader_RetrieveLearningProgress(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	//tx := &mock_database.Tx{}

	mockStudentLearningTimeDailyRepo := &mock_repositories.MockStudentLearningTimeDailyRepo{}
	mockUserMgmtService := &mock_services.MockUserMgmtService{}

	srv := &StudentLearningTimeReaderService{
		DB:                          db,
		StudentLearningTimeDaiyRepo: mockStudentLearningTimeDailyRepo,
		UserMgmtService:             mockUserMgmtService,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	start := timeutil.StartWeekIn(bpb.COUNTRY_VN)
	from := timestamppb.New(start)
	to := timestamppb.New(timeutil.EndWeekIn(bpb.COUNTRY_VN))

	tests := []TestCase{
		{
			name:         "null from/to",
			ctx:          ctx,
			req:          &epb.RetrieveLearningProgressRequest{StudentId: userId, From: nil, To: nil},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
		},
		{
			name:         "invalid from/to",
			ctx:          ctx,
			req:          &epb.RetrieveLearningProgressRequest{StudentId: userId, From: to, To: from},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
		},
		{
			name:        "out going context error",
			ctx:         ctx,
			req:         &epb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedErr: status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String()),
		},
		{
			name: "student can't retrieve other student",
			ctx: interceptors.ContextWithUserGroup(
				interceptors.NewIncomingContext(ctx), cpb.UserGroup_USER_GROUP_STUDENT.String(),
			),
			req:         &epb.RetrieveLearningProgressRequest{StudentId: "not me", From: from, To: to},
			expectedErr: status.Errorf(codes.PermissionDenied, codes.PermissionDenied.String()),
		},
		{
			name: "search profile error",
			ctx: interceptors.ContextWithUserGroup(
				interceptors.NewIncomingContext(ctx), cpb.UserGroup_USER_GROUP_STUDENT.String(),
			),
			req:         &epb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedErr: status.Errorf(codes.Internal, "RetrieveUserProfile: %v", grpc.ErrServerStopped),
			setup: func(ctx context.Context) {
				mockUserMgmtService.On("SearchBasicProfile", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, grpc.ErrServerStopped)
			},
		},
		{
			name: "search profile empty",
			ctx: interceptors.ContextWithUserGroup(
				interceptors.NewIncomingContext(ctx), cpb.UserGroup_USER_GROUP_STUDENT.String(),
			),
			req:         &epb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedErr: status.Errorf(codes.NotFound, codes.NotFound.String()),
			setup: func(ctx context.Context) {
				mockUserMgmtService.On("SearchBasicProfile", mock.Anything, mock.Anything, mock.Anything).Once().Return(&upb.SearchBasicProfileResponse{
					Profiles: []*cpb.BasicProfile{},
				}, nil)
			},
		},
		{
			name: "StudentLearningTimeDaiyRepo error",
			ctx: interceptors.ContextWithUserGroup(
				interceptors.NewIncomingContext(ctx), cpb.UserGroup_USER_GROUP_STUDENT.String(),
			),
			req:         &epb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.StudentLearningTimeDaiyRepo.Retrieve %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				mockUserMgmtService.On("SearchBasicProfile", mock.Anything, mock.Anything, mock.Anything).Once().Return(&upb.SearchBasicProfileResponse{
					Profiles: []*cpb.BasicProfile{
						{
							Country: cpb.Country_COUNTRY_VN,
						},
					},
				}, nil)
				mockStudentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, database.Text(userId), mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "not founds log",
			ctx: interceptors.ContextWithUserGroup(
				interceptors.NewIncomingContext(ctx), cpb.UserGroup_USER_GROUP_STUDENT.String(),
			),
			req: &epb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedResp: &epb.RetrieveLearningProgressResponse{Dailies: []*epb.RetrieveLearningProgressResponse_DailyLearningTime{
				{
					TotalTimeSpentInDay: 0,
					Day:                 timestamppb.New(start),
				},
				{
					TotalTimeSpentInDay: 0,
					Day:                 timestamppb.New(start.Add(24 * time.Hour)),
				},
				{
					TotalTimeSpentInDay: 0,
					Day:                 timestamppb.New(start.Add(24 * 2 * time.Hour)),
				},
				{
					TotalTimeSpentInDay: 0,
					Day:                 timestamppb.New(start.Add(24 * 3 * time.Hour)),
				},
				{
					TotalTimeSpentInDay: 0,
					Day:                 timestamppb.New(start.Add(24 * 4 * time.Hour)),
				},
				{
					TotalTimeSpentInDay: 0,
					Day:                 timestamppb.New(start.Add(24 * 5 * time.Hour)),
				},
				{
					TotalTimeSpentInDay: 0,
					Day:                 timestamppb.New(start.Add(24 * 6 * time.Hour)),
				},
			}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUserMgmtService.On("SearchBasicProfile", mock.Anything, mock.Anything, mock.Anything).Once().Return(&upb.SearchBasicProfileResponse{
					Profiles: []*cpb.BasicProfile{
						{
							Country: cpb.Country_COUNTRY_VN,
						},
					},
				}, nil)
				mockStudentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, database.Text(userId), mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudentLearningTimeDaily{}, nil)
			},
		},
		{
			name: "success",
			ctx:  interceptors.NewIncomingContext(ctx),
			req:  &epb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedResp: &epb.RetrieveLearningProgressResponse{
				Dailies: []*epb.RetrieveLearningProgressResponse_DailyLearningTime{
					{
						TotalTimeSpentInDay: 180,
						Day:                 timestamppb.New(start),
					},
					{
						TotalTimeSpentInDay: 300,
						Day:                 timestamppb.New(start.Add(24 * time.Hour)),
					},
					{
						TotalTimeSpentInDay: 0,
						Day:                 timestamppb.New(start.Add(24 * 2 * time.Hour)),
					},
					{
						TotalTimeSpentInDay: 0,
						Day:                 timestamppb.New(start.Add(24 * 3 * time.Hour)),
					},
					{
						TotalTimeSpentInDay: 0,
						Day:                 timestamppb.New(start.Add(24 * 4 * time.Hour)),
					},
					{
						TotalTimeSpentInDay: 0,
						Day:                 timestamppb.New(start.Add(24 * 5 * time.Hour)),
					},
					{
						TotalTimeSpentInDay: 0,
						Day:                 timestamppb.New(start.Add(24 * 6 * time.Hour)),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				learningTime := new(entities.StudentLearningTimeDaily)
				learningTime.ID.Set(1)
				learningTime.StudentID.Set("s1")
				learningTime.LearningTime.Set(180)
				learningTime.Day.Set(start)

				learningTime2 := new(entities.StudentLearningTimeDaily)
				learningTime2.ID.Set(1)
				learningTime2.StudentID.Set("s1")
				learningTime2.LearningTime.Set(300)
				learningTime2.Day.Set(start.Add(24 * time.Hour))

				mockUserMgmtService.On("SearchBasicProfile", mock.Anything, mock.Anything, mock.Anything).Once().Return(&upb.SearchBasicProfileResponse{
					Profiles: []*cpb.BasicProfile{
						{
							Country: cpb.Country_COUNTRY_VN,
						},
					},
				}, nil)
				mockStudentLearningTimeDailyRepo.On("Retrieve", mock.Anything, mock.Anything, database.Text(userId), mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudentLearningTimeDaily{
					learningTime, learningTime2,
				}, nil)
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(ctx)
			}
			_, err := srv.RetrieveLearningProgress(test.ctx, test.req.(*epb.RetrieveLearningProgressRequest))
			assert.Equal(t, test.expectedErr, err)
			if err != nil {
				t.Logf("err = %v", err)
			}
		})
	}
}
