package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_eureka_service "github.com/manabie-com/backend/mock/eureka/services"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

// Fix after LMS version 3
// func TestCalculateWeeklyPlan_StartWeekIsNot1(t *testing.T) {
// 	p1 := new(entities.PresetStudyPlanWeekly)
// 	p1.ID.Set("1")
// 	p1.PresetStudyPlanID.Set("ps1")
// 	p1.TopicID.Set("topic1")
// 	p1.Week.Set(1)

// 	p2 := new(entities.PresetStudyPlanWeekly)
// 	p2.ID.Set("2")
// 	p2.PresetStudyPlanID.Set("ps1")
// 	p2.TopicID.Set("topic1")
// 	p2.Week.Set(1)

// 	p3 := new(entities.PresetStudyPlanWeekly)
// 	p3.ID.Set("3")
// 	p3.PresetStudyPlanID.Set("ps1")
// 	p3.TopicID.Set("topic1")
// 	p3.Week.Set(2)

// 	p4 := new(entities.PresetStudyPlanWeekly)
// 	p4.ID.Set("4")
// 	p4.PresetStudyPlanID.Set("ps1")
// 	p4.TopicID.Set("topic1")
// 	p4.Week.Set(3)

// 	p5 := new(entities.PresetStudyPlanWeekly)
// 	p5.ID.Set("5")
// 	p5.PresetStudyPlanID.Set("ps1")
// 	p5.TopicID.Set("topic1")
// 	p5.Week.Set(4)

// 	weeklies := []*entities.PresetStudyPlanWeekly{p1, p2, p3, p4, p5}

// 	startDate := time.Date(2019, time.June, 17, 0, 0, 0, 0, time.UTC) // Monday
// 	req := &pb.AssignPresetStudyPlansRequest_PlanDetail{
// 		PresetStudyPlanId: "ps1",
// 		StartDate:         &types.Timestamp{Seconds: startDate.Unix()},
// 		StartWeek:         3,
// 	}
// 	plans := (&StudentService{}).calculateWeeklyPlan(req, weeklies)
// 	if expected, got := time.Date(2019, time.June, 3, 0, 0, 0, 0, time.UTC), plans[0].StartDate.Time; !expected.Equal(got) {
// 		t.Errorf("expected plan start date: %v, got: %v", expected, got)
// 	}
// 	if expected, got := time.Date(2019, time.June, 3, 0, 0, 0, 0, time.UTC), plans[1].StartDate.Time; !expected.Equal(got) {
// 		t.Errorf("expected plan start date: %v, got: %v", expected, got)
// 	}
// 	if expected, got := time.Date(2019, time.June, 10, 0, 0, 0, 0, time.UTC), plans[2].StartDate.Time; !expected.Equal(got) {
// 		t.Errorf("expected plan start date: %v, got: %v", expected, got)
// 	}
// 	if expected, got := startDate, plans[3].StartDate.Time; !expected.Equal(got) {
// 		t.Errorf("expected plan start date: %v, got: %v", expected, got)
// 	}
// 	if expected, got := time.Date(2019, time.June, 24, 0, 0, 0, 0, time.UTC), plans[4].StartDate.Time; !expected.Equal(got) {
// 		t.Errorf("expected plan start date: %v, got: %v", expected, got)
// 	}
// }

func TestStudentService_RetrieveProfile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentRepo := new(mock_repositories.MockStudentRepo)

	s := &StudentService{
		StudentRepo: studentRepo,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	now := time.Now()
	timeProto, _ := types.TimestampProto(now)
	nowProtoResp := timestamppb.New(now)
	testCases := []TestCase{
		{
			name:         "get my profile err",
			ctx:          ctx,
			req:          &pb.GetStudentProfileRequest{StudentIds: []string{userId}},
			expectedResp: nil,
			expectedErr:  toStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					getStudentProfile: func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
						return nil, toStatusError(pgx.ErrNoRows)
					},
				}
			},
		},
		{
			name: "get another student assigned profile",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{StudentIds: []string{"1"}},
			expectedResp: &pb.GetStudentProfileResponse{
				Datas: []*pb.GetStudentProfileResponse_Data{
					{
						Profile: &pb.StudentProfile{
							Id:        userId,
							Birthday:  timeProto,
							CreatedAt: timeProto,
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					getStudentProfile: func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
						return &upb.GetStudentProfileResponse{
							Profiles: []*upb.StudentProfile{
								{
									Id:        userId,
									Birthday:  nowProtoResp,
									CreatedAt: nowProtoResp,
								},
							},
						}, nil
					},
				}
			},
		},
		{
			name: "get my profile",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{StudentIds: []string{userId}},
			expectedResp: &pb.GetStudentProfileResponse{
				Datas: []*pb.GetStudentProfileResponse_Data{
					{
						Profile: &pb.StudentProfile{
							Id:        userId,
							Birthday:  timeProto,
							CreatedAt: timeProto,
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					getStudentProfile: func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
						return &upb.GetStudentProfileResponse{
							Profiles: []*upb.StudentProfile{
								{
									Id:        userId,
									Birthday:  nowProtoResp,
									CreatedAt: nowProtoResp,
								},
							},
						}, nil
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.GetStudentProfile(testCase.ctx, testCase.req.(*pb.GetStudentProfileRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudentService_FindStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentRepo := new(mock_repositories.MockStudentRepo)
	userRepo := new(mock_repositories.MockUserRepo)

	s := &StudentServiceABAC{
		&StudentService{
			StudentRepo: studentRepo,
			UserRepo:    userRepo,
		},
	}

	student := &entities.Student{}
	_ = student.PhoneNumber.Set("validPhone")

	testCases := []TestCase{
		{
			name:        "valid student phone number",
			ctx:         interceptors.ContextWithUserID(ctx, "staff"),
			req:         &pb.FindStudentRequest{Phone: "validPhone"},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentRepo.On("FindByPhone", ctx, mock.Anything, database.Text("validPhone")).Once().Return(student, nil)
			},
		},
		{
			name:        "invalid student id",
			ctx:         interceptors.ContextWithUserID(ctx, "staff"),
			req:         &pb.FindStudentRequest{Phone: "invalidPhone"},
			expectedErr: status.Error(codes.NotFound, "student not found"),
			setup: func(ctx context.Context) {
				studentRepo.On("FindByPhone", ctx, mock.Anything, database.Text("invalidPhone")).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.FindStudent(testCase.ctx, testCase.req.(*pb.FindStudentRequest))
			assert.Equal(t, testCase.expectedErr, err)

			if err == nil {
				assert.Equal(t, resp.Profile.Phone, testCase.req.(*pb.FindStudentRequest).Phone)
			}
		})
	}
}

func TestStudentService_RetrieveLearningProgress(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEurekaSvc := &mock_eureka_service.MockStudentLearningTimeReaderClient{}

	s := &StudentServiceABAC{
		&StudentService{
			StudentLearningTimeSvc: mockEurekaSvc,
		},
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	start := timeutil.StartWeekIn(pb.COUNTRY_VN)
	from, _ := types.TimestampProto(start)
	to, _ := types.TimestampProto(timeutil.EndWeekIn(pb.COUNTRY_VN))

	testCases := []TestCase{
		{
			name:         "invalid request",
			ctx:          ctx,
			req:          &pb.RetrieveLearningProgressRequest{StudentId: userId, From: nil, To: nil},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "date time not valid",
			ctx:          ctx,
			req:          &pb.RetrieveLearningProgressRequest{StudentId: userId, From: to, To: from},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "student can't get others",
			ctx:          interceptors.ContextWithUserGroup(ctx, cpb.UserGroup_USER_GROUP_STUDENT.String()),
			req:          &pb.RetrieveLearningProgressRequest{StudentId: "hachs kow", From: from, To: to},
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, codes.PermissionDenied.String()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "call eureka error",
			ctx:          interceptors.NewIncomingContext(ctx),
			req:          &pb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedResp: nil,
			expectedErr:  grpc.ErrServerStopped,
			setup: func(ctx context.Context) {
				mockEurekaSvc.On("RetrieveLearningProgress", mock.Anything, mock.Anything).Once().Return(nil, grpc.ErrServerStopped)
			},
		},
		{
			name: "happy case",
			ctx:  interceptors.NewIncomingContext(ctx),
			req:  &pb.RetrieveLearningProgressRequest{StudentId: userId, From: from, To: to},
			expectedResp: &pb.RetrieveLearningProgressResponse{
				Dailies: []*pb.RetrieveLearningProgressResponse_DailyLearningTime{
					{
						TotalTimeSpentInDay: 42,
					},
					{
						TotalTimeSpentInDay: 0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockEurekaSvc.On("RetrieveLearningProgress", mock.Anything, mock.Anything).Once().Return(
					&epb.RetrieveLearningProgressResponse{
						Dailies: []*epb.RetrieveLearningProgressResponse_DailyLearningTime{
							{
								TotalTimeSpentInDay: 42,
							},
							{
								TotalTimeSpentInDay: 0,
							},
						},
					},
					nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveLearningProgress(testCase.ctx, testCase.req.(*pb.RetrieveLearningProgressRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudentService_RetrievePresetStudyPlans(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)

	s := &StudentService{
		PresetStudyPlanRepo: presetStudyPlanRepo,
	}

	now := time.Now()
	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	testCases := []TestCase{
		{
			name:         "err db RetrievePresetStudyPlans",
			ctx:          ctx,
			req:          &pb.RetrievePresetStudyPlansRequest{Country: pb.COUNTRY_VN, Subject: pb.SUBJECT_CHEMISTRY},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, errors.Wrapf(pgx.ErrNoRows, "c.PresetStudyPlanRepo.RetrievePresetStudyPlans: grade: %v", "").Error()),
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("RetrievePresetStudyPlans", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:         "err convert grade",
			ctx:          ctx,
			req:          &pb.RetrievePresetStudyPlansRequest{Country: pb.COUNTRY_NONE, Subject: pb.SUBJECT_CHEMISTRY, Grade: "G13"},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "cannot find country grade map"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.RetrievePresetStudyPlansRequest{Country: pb.COUNTRY_VN, Subject: pb.SUBJECT_CHEMISTRY},
			expectedResp: &pb.RetrievePresetStudyPlansResponse{
				PresetStudyPlans: []*pb.PresetStudyPlan{
					{
						Id:        "1",
						Name:      "p",
						Grade:     "Lớp 12",
						Country:   pb.COUNTRY_VN,
						Subject:   pb.SUBJECT_CHEMISTRY,
						CreatedAt: &types.Timestamp{Seconds: now.Unix()},
						UpdatedAt: &types.Timestamp{Seconds: now.Unix()},
						StartDate: &types.Timestamp{Seconds: timeutil.DefaultPSPStartDate(pb.COUNTRY_VN).Unix()},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				p := new(entities.PresetStudyPlan)
				p.ID.Set("1")
				p.Name.Set("p")
				p.Country.Set("COUNTRY_VN")
				p.Grade.Set(12)
				p.Subject.Set("SUBJECT_CHEMISTRY")
				p.CreatedAt.Set(now)
				p.UpdatedAt.Set(now)
				p.StartDate.Set(timeutil.DefaultPSPStartDate(pb.COUNTRY_VN))
				presetStudyPlanRepo.On("RetrievePresetStudyPlans", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.PresetStudyPlan{p}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrievePresetStudyPlans(testCase.ctx, testCase.req.(*pb.RetrievePresetStudyPlansRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudentService_RetrievePresetStudyPlanWeeklies(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)

	s := &StudentService{
		PresetStudyPlanRepo: presetStudyPlanRepo,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	testCases := []TestCase{
		{
			name:         "err db RetrievePresetStudyPlans",
			ctx:          ctx,
			req:          &pb.RetrievePresetStudyPlanWeekliesRequest{PresetStudyPlanId: "PresetStudyPlanId"},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, errors.Wrapf(pgx.ErrNoRows, "c.PresetStudyPlanRepo.RetrievePresetStudyPlanWeeklies").Error()),
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("RetrievePresetStudyPlanWeeklies", ctx, mock.Anything, database.Text("PresetStudyPlanId")).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.RetrievePresetStudyPlanWeekliesRequest{PresetStudyPlanId: "PresetStudyPlanId"},
			expectedResp: &pb.RetrievePresetStudyPlanWeekliesResponse{
				PresetStudyPlanWeeklies: []*pb.PresetStudyPlanWeekly{
					toPresetStudyPlanWeeklyPb(&entities.PresetStudyPlanWeekly{Week: pgtype.Int2{Int: 1, Status: 2}}),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("RetrievePresetStudyPlanWeeklies", ctx, mock.Anything, database.Text("PresetStudyPlanId")).Once().Return([]*entities.PresetStudyPlanWeekly{{Week: pgtype.Int2{Int: 1, Status: 2}}}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrievePresetStudyPlanWeeklies(testCase.ctx, testCase.req.(*pb.RetrievePresetStudyPlanWeekliesRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

// TO-DO Rewrite unit test
// func TestStudentService_AssignPresetStudyPlans(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)
// 	studentRepo := new(mock_repositories.MockStudentRepo)
// 	userRepo := new(mock_repositories.MockUserRepo)
// 	loComplRepo := new(mock_repositories.MockStudentsLearningObjectivesCompletenessRepo)
// 	topicComplRepo := new(mock_repositories.MockStudentTopicCompletenessRepo)
// 	studentTopicOverdueRepo := new(mock_repositories.MockStudentTopicOverdueRepo)

// 	s := &StudentService{
// 		PresetStudyPlanRepo:          presetStudyPlanRepo,
// 		UserRepo:                     userRepo,
// 		StudentRepo:                  studentRepo,
// 		StudentLOCompletenessRepo:    loComplRepo,
// 		StudentTopicCompletenessRepo: topicComplRepo,
// 		StudentTopicOverdueRepo:      studentTopicOverdueRepo,
// 	}

// 	userId := idutil.ULIDNow()
// 	ctx = interceptors.ContextWithUserId(ctx, userId)

// 	testCases := []TestCase{
// 		{
// 			name:         "err db get user group",
// 			ctx:          ctx,
// 			req:          &pb.AssignPresetStudyPlansRequest{StudentId: userId},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(errors.Wrapf(pgx.ErrNoRows, "s.UserRepo.UserGroup: userID: %q", userId), "canProcessStudentData").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return("", pgx.ErrNoRows)
// 			},
// 		},
// 		{
// 			name:         "err permission denied",
// 			ctx:          ctx,
// 			req:          &pb.AssignPresetStudyPlansRequest{StudentId: "another-userid"},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.PermissionDenied, codes.PermissionDenied.String()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 			},
// 		},
// 		{
// 			name: "err db when RetrievePresetStudyPlanWeeklies",
// 			ctx:  ctx,
// 			req: &pb.AssignPresetStudyPlansRequest{
// 				StudentId: userId,
// 				PlanDetails: []*pb.AssignPresetStudyPlansRequest_PlanDetail{
// 					{
// 						PresetStudyPlanId: "",
// 						StartWeek:         0,
// 						StartDate:         nil,
// 					},
// 				},
// 			},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(errors.New("c.PresetStudyPlanRepo.RetrievePresetStudyPlanWeeklies: no rows in result set"), "eg.Wait").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 				presetStudyPlanRepo.On("RetrievePresetStudyPlanWeeklies", mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
// 			},
// 		},
// 		{
// 			name: "err db when assign preset student plan",
// 			ctx:  ctx,
// 			req: &pb.AssignPresetStudyPlansRequest{
// 				StudentId: userId,
// 				PlanDetails: []*pb.AssignPresetStudyPlansRequest_PlanDetail{
// 					{
// 						PresetStudyPlanId: "",
// 						StartWeek:         0,
// 						StartDate:         types.TimestampNow(),
// 					},
// 				},
// 			},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(pgx.ErrTxClosed, "c.PresetStudyPlanRepo.AssignPresetStudyPlan").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 				presetStudyPlanRepo.On("RetrievePresetStudyPlanWeeklies", mock.Anything, mock.Anything).Once().Return([]*entities.PresetStudyPlanWeekly{}, nil)
// 				presetStudyPlanRepo.On("AssignPresetStudyPlan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
// 				studentTopicOverdueRepo.On("AssignOverdueTopic", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
// 				studentTopicOverdueRepo.On("RetrieveStudentTopicOverdue", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
// 			},
// 		},
// 		{
// 			name:         "err db when assign preset student plan",
// 			ctx:          ctx,
// 			req:          &pb.AssignPresetStudyPlansRequest{StudentId: userId},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(pgx.ErrTxClosed, "c.PresetStudyPlanRepo.AssignPresetStudyPlan").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 				presetStudyPlanRepo.On("AssignPresetStudyPlan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
// 			},
// 		},
// 		{
// 			name:         "success",
// 			ctx:          ctx,
// 			req:          &pb.AssignPresetStudyPlansRequest{StudentId: userId},
// 			expectedResp: &pb.AssignPresetStudyPlansResponse{Successful: true},
// 			expectedErr:  nil,
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 				presetStudyPlanRepo.On("RetrieveStudentPresetStudyPlanWeeklies", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
// 				presetStudyPlanRepo.On("AssignPresetStudyPlan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
// 				loComplRepo.On("RetrieveFinishedLOs", mock.Anything, mock.Anything).Once().Return(nil, nil)
// 				topicComplRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			testCase.setup(testCase.ctx)
// 			resp, err := s.AssignPresetStudyPlans(testCase.ctx, testCase.req.(*pb.AssignPresetStudyPlansRequest))
// 			assert.Equal(t, testCase.expectedErr, err)
// 			if testCase.expectedResp == nil {
// 				assert.Nil(t, testCase.expectedResp, resp)
// 			} else {
// 				assert.Equal(t, testCase.expectedResp, resp)
// 			}
// 		})
// 	}
// }

// TO-DO Fix after LMS version 3
// func TestStudentService_RetrieveStudentStudyPlans(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)
// 	studentRepo := new(mock_repositories.MockStudentRepo)
// 	userRepo := new(mock_repositories.MockUserRepo)

// 	s := &StudentService{
// 		PresetStudyPlanRepo: presetStudyPlanRepo,
// 		UserRepo:            userRepo,
// 		StudentRepo:         studentRepo,
// 	}

// 	userId := idutil.ULIDNow()
// 	ctx = interceptors.ContextWithUserId(ctx, userId)

// 	tfrom := time.Now().Add(-time.Hour)
// 	from, _ := types.TimestampProto(tfrom)
// 	now := types.TimestampNow()
// 	timeNow := time.Now()

// 	testCases := []TestCase{
// 		{
// 			name:         "invalid from/to",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlansRequest{StudentId: userId, From: now, To: from},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
// 			setup: func(ctx context.Context) {
// 			},
// 		},
// 		{
// 			name:         "err permission denied",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlansRequest{StudentId: userId},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(errors.Wrapf(pgx.ErrNoRows, "s.UserRepo.UserGroup: userID: %q", userId), "canProcessStudentData").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return("", pgx.ErrNoRows)
// 			},
// 		},
// 		{
// 			name:         "err permission denied with another userId",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlansRequest{StudentId: "another-userid"},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.PermissionDenied, codes.PermissionDenied.String()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 			},
// 		},
// 		{
// 			name:         "err db RetrieveStudentPresetStudyPlans",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlansRequest{StudentId: userId, From: from, To: now},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(pgx.ErrNoRows, "s.PresetStudyPlanRepo.RetrieveStudentPresetStudyPlans").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 				presetStudyPlanRepo.On("RetrieveStudentPresetStudyPlans", ctx, database.Text(userId), mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
// 			},
// 		},
// 		{
// 			name: "err db RetrieveStudentPresetStudyPlans",
// 			ctx:  ctx,
// 			req:  &pb.RetrieveStudentStudyPlansRequest{StudentId: userId, From: from, To: now},
// 			expectedResp: &pb.RetrieveStudentStudyPlansResponse{PlanWithStartDates: []*pb.RetrieveStudentStudyPlansResponse_PlanWithStartDate{
// 				{
// 					Plan: &pb.PresetStudyPlan{
// 						Id:        "1",
// 						Name:      "p",
// 						Grade:     "G12",
// 						Country:   pb.COUNTRY_VN,
// 						Subject:   pb.SUBJECT_CHEMISTRY,
// 						CreatedAt: &types.Timestamp{Seconds: timeNow.Unix()},
// 						UpdatedAt: &types.Timestamp{Seconds: timeNow.Unix()},
// 						StartDate: &types.Timestamp{Seconds: timeutil.DefaultPSPStartDate(pb.COUNTRY_VN).Unix()},
// 					},
// 					Week:      0,
// 					StartDate: &types.Timestamp{Seconds: tfrom.Unix()},
// 				},
// 			}},
// 			expectedErr: nil,
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)

// 				p := new(entities.PresetStudyPlan)
// 				p.ID.Set("1")
// 				p.Name.Set("p")
// 				p.Country.Set("COUNTRY_VN")
// 				p.Grade.Set(12)
// 				p.Subject.Set("SUBJECT_CHEMISTRY")
// 				p.CreatedAt.Set(timeNow)
// 				p.UpdatedAt.Set(timeNow)
// 				p.StartDate.Set(timeutil.DefaultPSPStartDate(pb.COUNTRY_VN))
// 				planWithDate := &repositories.PlanWithStartDate{
// 					PresetStudyPlan: p,
// 					Week:            &pgtype.Int2{Int: 0, Status: 2},
// 					StartDate:       &pgtype.Timestamptz{Time: tfrom, Status: 2},
// 				}

// 				presetStudyPlanRepo.On("RetrieveStudentPresetStudyPlans", ctx, database.Text(userId), mock.Anything, mock.Anything).Once().Return([]*repositories.PlanWithStartDate{planWithDate}, nil)
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			testCase.setup(testCase.ctx)
// 			resp, err := s.RetrieveStudentStudyPlans(testCase.ctx, testCase.req.(*pb.RetrieveStudentStudyPlansRequest))
// 			assert.Equal(t, testCase.expectedErr, err)
// 			if testCase.expectedResp == nil {
// 				assert.Nil(t, testCase.expectedResp, resp)
// 			} else {
// 				assert.Equal(t, testCase.expectedResp, resp)
// 			}
// 		})
// 	}
// }

// TO-DO Fix after LMS version 3
// func TestStudentService_RetrieveStudentStudyPlanWeeklies(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)
// 	studentRepo := new(mock_repositories.MockStudentRepo)
// 	userRepo := new(mock_repositories.MockUserRepo)

// 	s := &StudentService{
// 		PresetStudyPlanRepo: presetStudyPlanRepo,
// 		UserRepo:            userRepo,
// 		StudentRepo:         studentRepo,
// 	}

// 	userId := idutil.ULIDNow()
// 	ctx = interceptors.ContextWithUserId(ctx, userId)

// 	tfrom := time.Now().Add(-time.Hour)
// 	from, _ := types.TimestampProto(tfrom)
// 	now := types.TimestampNow()

// 	testCases := []TestCase{
// 		{
// 			name:         "invalid from/to",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlanWeekliesRequest{StudentId: userId, From: now, To: from, RetrieveAll: false},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
// 			setup: func(ctx context.Context) {
// 			},
// 		},
// 		{
// 			name:         "invalid from/to",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlanWeekliesRequest{StudentId: userId, From: now, To: from, RetrieveAll: false},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
// 			setup: func(ctx context.Context) {
// 			},
// 		},
// 		{
// 			name:         "err permission denied",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlanWeekliesRequest{StudentId: userId},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(errors.Wrapf(pgx.ErrNoRows, "s.UserRepo.UserGroup: userID: %q", userId), "canProcessStudentData").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return("", pgx.ErrNoRows)
// 			},
// 		},
// 		{
// 			name:         "err permission denied with another userId",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlanWeekliesRequest{StudentId: "another-userid", RetrieveAll: false},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.PermissionDenied, codes.PermissionDenied.String()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 			},
// 		},
// 		{
// 			name:         "err db RetrieveStudentPresetStudyPlanWeeklies",
// 			ctx:          ctx,
// 			req:          &pb.RetrieveStudentStudyPlanWeekliesRequest{StudentId: userId, From: from, To: now, RetrieveAll: false},
// 			expectedResp: nil,
// 			expectedErr:  status.Error(codes.Unknown, errors.Wrap(pgx.ErrNoRows, "s.PresetStudyPlanRepo.RetrieveStudentPresetStudyPlanWeeklies").Error()),
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 				presetStudyPlanRepo.On("RetrieveStudentPresetStudyPlanWeeklies", ctx, database.Text(userId), mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
// 			},
// 		},
// 		{
// 			name: "err db RetrieveStudentPresetStudyPlanWeeklies",
// 			ctx:  ctx,
// 			req:  &pb.RetrieveStudentStudyPlanWeekliesRequest{StudentId: userId, From: from, To: now, RetrieveAll: false},
// 			expectedResp: &pb.RetrieveStudentStudyPlanWeekliesResponse{
// 				TopicWithStartDates: []*pb.RetrieveStudentStudyPlanWeekliesResponse_TopicWithStartDate{
// 					{
// 						TopicId:   "topicId",
// 						TopicName: "topicName",
// 						StartDate: &types.Timestamp{Seconds: tfrom.Unix()},
// 					},
// 				},
// 			},
// 			expectedErr: nil,
// 			setup: func(ctx context.Context) {
// 				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupStudent, nil)
// 				topic := repositories.Topic{
// 					Topic: entities.Topic{
// 						ID:   pgtype.Text{String: "topicId", Status: 2},
// 						Name: pgtype.Text{String: "topicName", Status: 2},
// 					},
// 					StartDate: pgtype.Timestamptz{Time: tfrom, Status: 2},
// 				}

// 				presetStudyPlanRepo.On("RetrieveStudentPresetStudyPlanWeeklies", ctx, database.Text(userId), mock.Anything, mock.Anything, mock.Anything).Once().Return([]repositories.Topic{topic}, nil)
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			testCase.setup(testCase.ctx)
// 			resp, err := s.RetrieveStudentStudyPlanWeeklies(testCase.ctx, testCase.req.(*pb.RetrieveStudentStudyPlanWeekliesRequest))
// 			assert.Equal(t, testCase.expectedErr, err)
// 			if testCase.expectedResp == nil {
// 				assert.Nil(t, testCase.expectedResp, resp)
// 			} else {
// 				assert.Equal(t, testCase.expectedResp, resp)
// 			}
// 		})
// 	}
// }

type mockUserMgmtStudentService struct {
	getStudentProfile      func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error)
	upsertStudentComment   func(ctx context.Context, in *upb.UpsertStudentCommentRequest, opts ...grpc.CallOption) (*upb.UpsertStudentCommentResponse, error)
	deleteStudentComments  func(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error)
	retrieveStudentComment func(ctx context.Context, in *upb.RetrieveStudentCommentRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentCommentResponse, error)
}

func (m *mockUserMgmtStudentService) GetStudentProfile(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
	return m.getStudentProfile(ctx, in, opts...)
}

func (m *mockUserMgmtStudentService) UpsertStudentComment(ctx context.Context, in *upb.UpsertStudentCommentRequest, opts ...grpc.CallOption) (*upb.UpsertStudentCommentResponse, error) {
	return m.upsertStudentComment(ctx, in, opts...)
}

func (m *mockUserMgmtStudentService) DeleteStudentComments(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error) {
	return m.deleteStudentComments(ctx, in, opts...)
}

func (m *mockUserMgmtStudentService) RetrieveStudentComment(ctx context.Context, in *upb.RetrieveStudentCommentRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentCommentResponse, error) {
	return m.retrieveStudentComment(ctx, in, opts...)
}

func TestStudentService_UpsertStudentComment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCommentRepo := new(mock_repositories.MockStudentCommentRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	userRepo := new(mock_repositories.MockUserRepo)

	s := &StudentService{
		StudentCommentRepo: studentCommentRepo,
		UserRepo:           userRepo,
		StudentRepo:        studentRepo,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	// tfrom := time.Now().Add(-time.Hour)
	// from, _ := types.TimestampProto(tfrom)
	now := types.TimestampNow()

	testCases := []TestCase{
		{
			name: "err: usermgmt conn failed",
			ctx:  ctx,
			req: &pb.UpsertStudentCommentRequest{
				StudentComment: &pb.StudentComment{
					CommentId:      "commentId",
					StudentId:      "studentId",
					CommentContent: "content",
					UpdatedAt:      now,
					CreatedAt:      now,
				},
			},
			expectedResp: nil,
			expectedErr:  errors.New("s.UserMgmtStudentSvc.UpsertStudentComment: rpc error: code = Unknown desc = usermgmt conn failed"),
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					upsertStudentComment: func(ctx context.Context, in *upb.UpsertStudentCommentRequest, opts ...grpc.CallOption) (*upb.UpsertStudentCommentResponse, error) {
						return nil, fmt.Errorf("usermgmt conn failed")
					},
					deleteStudentComments: func(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error) {
						return nil, fmt.Errorf("usermgmt conn failed")
					},
				}
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req: &pb.UpsertStudentCommentRequest{
				StudentComment: &pb.StudentComment{
					CommentId:      "commentId",
					StudentId:      "studentId",
					CommentContent: "content",
					UpdatedAt:      now,
					CreatedAt:      now,
				},
			},
			expectedResp: &pb.UpsertStudentCommentResponse{
				Successful: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					upsertStudentComment: func(ctx context.Context, in *upb.UpsertStudentCommentRequest, opts ...grpc.CallOption) (*upb.UpsertStudentCommentResponse, error) {
						return &upb.UpsertStudentCommentResponse{
							Successful: true,
						}, nil
					},
					deleteStudentComments: func(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error) {
						return &upb.DeleteStudentCommentsResponse{
							Successful: true,
						}, nil
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.UpsertStudentComment(testCase.ctx, testCase.req.(*pb.UpsertStudentCommentRequest))
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudentService_RetrieveStudentComment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCommentRepo := new(mock_repositories.MockStudentCommentRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	userRepo := new(mock_repositories.MockUserRepo)

	s := &StudentService{
		StudentCommentRepo: studentCommentRepo,
		UserRepo:           userRepo,
		StudentRepo:        studentRepo,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	tnow := time.Now()

	testCases := []TestCase{
		{
			name:         "invalid student id",
			ctx:          ctx,
			req:          &pb.RetrieveStudentCommentRequest{StudentId: "invalid-student-id"},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "s.UserMgmtStudentSvc.RetrieveStudentComment: rpc error: code = InvalidArgument desc = s.UserMgmtStudentSvc.RetrieveStudentComment student id invalid"),
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					retrieveStudentComment: func(ctx context.Context, in *upb.RetrieveStudentCommentRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentCommentResponse, error) {
						return nil, status.Error(codes.InvalidArgument, "s.UserMgmtStudentSvc.RetrieveStudentComment: rpc error: code = InvalidArgument desc = s.UserMgmtStudentSvc.RetrieveStudentComment student id invalid")
					},
				}
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.RetrieveStudentCommentRequest{StudentId: "student-id"},
			expectedResp: &pb.RetrieveStudentCommentResponse{
				Comment: []*pb.CommentInfo{
					{
						StudentComment: toStudentCommentPb(&entities.StudentComment{
							UpdatedAt: pgtype.Timestamptz{Time: tnow, Status: 2},
							CreatedAt: pgtype.Timestamptz{Time: tnow, Status: 2},
						}),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					retrieveStudentComment: func(ctx context.Context, in *upb.RetrieveStudentCommentRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentCommentResponse, error) {
						return &upb.RetrieveStudentCommentResponse{
							Comment: []*upb.CommentInfo{
								{
									StudentComment: &upb.StudentComment{
										UpdatedAt: &timestamppb.Timestamp{Seconds: tnow.Unix()},
										CreatedAt: &timestamppb.Timestamp{Seconds: tnow.Unix()},
									},
								},
							},
						}, nil
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveStudentComment(testCase.ctx, testCase.req.(*pb.RetrieveStudentCommentRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudent2Profile(t *testing.T) {
	t.Parallel()
	now := time.Now()

	user := entities.User{
		ID:                pgtype.Text{String: "user-id"},
		Avatar:            pgtype.Text{String: "avatar"},
		Group:             pgtype.Text{String: "group"},
		LastName:          pgtype.Text{String: "name"},
		Country:           pgtype.Text{String: pb.COUNTRY_VN.String()},
		PhoneNumber:       pgtype.Text{String: "phone"},
		Email:             pgtype.Text{String: "email"},
		DeviceToken:       pgtype.Text{String: "device-token"},
		AllowNotification: pgtype.Bool{Bool: true},
		UpdatedAt:         pgtype.Timestamptz{Time: now},
		CreatedAt:         pgtype.Timestamptz{Time: now},
	}

	testCases := []TestCase{
		{
			name: "student on trial",
			req: &entities.Student{
				User:             user,
				ID:               user.ID,
				CurrentGrade:     pgtype.Int2{Int: 12},
				TargetUniversity: pgtype.Text{String: "target-university"},
				Biography:        pgtype.Text{String: "biography"},
				Birthday:         pgtype.Date{Time: now},
				OnTrial:          pgtype.Bool{Bool: true},
				BillingDate:      pgtype.Timestamptz{Time: now.Add(24 * time.Hour)},
				UpdatedAt:        user.UpdatedAt,
				CreatedAt:        user.CreatedAt,
			},
			expectedResp: &pb.StudentProfile{
				Id:               user.ID.String,
				Name:             user.GetName(),
				Country:          pb.Country(pb.Country_value[user.Country.String]),
				Phone:            user.PhoneNumber.String,
				Email:            user.Email.String,
				Grade:            "Lớp 12",
				TargetUniversity: "target-university",
				Avatar:           "avatar",
				Birthday:         &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				Biography:        "biography",
				PaymentStatus:    pb.PAYMENT_STATUS_ON_TRIAL,
				BillingDate:      &types.Timestamp{Seconds: int64(now.Add(24 * time.Hour).Unix()), Nanos: int32(now.Add(24 * time.Hour).Nanosecond())},
				CreatedAt:        &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				PlanId:           "",
			},
		},
		{
			name: "student paid",
			req: &entities.Student{
				User:             user,
				ID:               user.ID,
				CurrentGrade:     pgtype.Int2{Int: 12},
				TargetUniversity: pgtype.Text{String: "target-university"},
				Biography:        pgtype.Text{String: "biography"},
				Birthday:         pgtype.Date{Time: now},
				OnTrial:          pgtype.Bool{Bool: false},
				BillingDate:      pgtype.Timestamptz{Time: now.Add(24 * time.Hour)},
				UpdatedAt:        user.UpdatedAt,
				CreatedAt:        user.CreatedAt,
			},
			expectedResp: &pb.StudentProfile{
				Id:               user.ID.String,
				Name:             user.GetName(),
				Country:          pb.Country(pb.Country_value[user.Country.String]),
				Phone:            user.PhoneNumber.String,
				Email:            user.Email.String,
				Grade:            "Lớp 12",
				TargetUniversity: "target-university",
				Avatar:           "avatar",
				Birthday:         &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				Biography:        "biography",
				PaymentStatus:    pb.PAYMENT_STATUS_PAID,
				BillingDate:      &types.Timestamp{Seconds: int64(now.Add(24 * time.Hour).Unix()), Nanos: int32(now.Add(24 * time.Hour).Nanosecond())},
				CreatedAt:        &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				PlanId:           "",
			},
		},
		{
			name: "student expired trial",
			req: &entities.Student{
				User:             user,
				ID:               user.ID,
				CurrentGrade:     pgtype.Int2{Int: 12},
				TargetUniversity: pgtype.Text{String: "target-university"},
				Biography:        pgtype.Text{String: "biography"},
				Birthday:         pgtype.Date{Time: now},
				OnTrial:          pgtype.Bool{Bool: true},
				BillingDate:      pgtype.Timestamptz{Time: now.Add(-24 * time.Hour)},
				UpdatedAt:        user.UpdatedAt,
				CreatedAt:        user.CreatedAt,
			},
			expectedResp: &pb.StudentProfile{
				Id:               user.ID.String,
				Name:             user.GetName(),
				Country:          pb.Country(pb.Country_value[user.Country.String]),
				Phone:            user.PhoneNumber.String,
				Email:            user.Email.String,
				Grade:            "Lớp 12",
				TargetUniversity: "target-university",
				Avatar:           "avatar",
				Birthday:         &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				Biography:        "biography",
				PaymentStatus:    pb.PAYMENT_STATUS_EXPIRED_TRIAL,
				BillingDate:      &types.Timestamp{Seconds: int64(now.Add(-24 * time.Hour).Unix()), Nanos: int32(now.Add(-24 * time.Hour).Nanosecond())},
				CreatedAt:        &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				PlanId:           "",
			},
		},
		{
			name: "student late payment",
			req: &entities.Student{
				User:             user,
				ID:               user.ID,
				CurrentGrade:     pgtype.Int2{Int: 12},
				TargetUniversity: pgtype.Text{String: "target-university"},
				Biography:        pgtype.Text{String: "biography"},
				Birthday:         pgtype.Date{Time: now},
				OnTrial:          pgtype.Bool{Bool: false},
				BillingDate:      pgtype.Timestamptz{Time: now.Add(-24 * time.Hour)},
				UpdatedAt:        user.UpdatedAt,
				CreatedAt:        user.CreatedAt,
			},
			expectedResp: &pb.StudentProfile{
				Id:               user.ID.String,
				Name:             user.GetName(),
				Country:          pb.Country(pb.Country_value[user.Country.String]),
				Phone:            user.PhoneNumber.String,
				Email:            user.Email.String,
				Grade:            "Lớp 12",
				TargetUniversity: "target-university",
				Avatar:           "avatar",
				Birthday:         &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				Biography:        "biography",
				PaymentStatus:    pb.PAYMENT_STATUS_LATE_PAYMENT,
				BillingDate:      &types.Timestamp{Seconds: int64(now.Add(-24 * time.Hour).Unix()), Nanos: int32(now.Add(-24 * time.Hour).Nanosecond())},
				CreatedAt:        &types.Timestamp{Seconds: int64(now.Unix()), Nanos: int32(now.Nanosecond())},
				PlanId:           "",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := student2Profile(testCase.req.(*entities.Student))
			assert.Equal(t, testCase.expectedResp, result)
		})
	}
}

func TestStudentService_StudentPermission(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userID := "student-id"
	ctx = interceptors.ContextWithUserID(ctx, userID)

	userRepo := new(mock_repositories.MockUserRepo)

	s := &StudentService{
		UserRepo: userRepo,
	}
	n := timeutil.Now()

	updatedAt := pgtype.Timestamptz{}
	updatedAt.Set(n)

	result := map[int32]*pb.Permission{
		int32(12): allowAll,
		int32(11): allowAll,
		int32(10): allowAll,
		int32(9):  allowAll,
		int32(8):  allowAll,
		int32(7):  allowAll,
		int32(6):  allowAll,
		int32(5):  allowAll,
	}
	testCases := []TestCase{
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.StudentPermissionRequest{},
			expectedResp: &pb.StudentPermissionResponse{
				Permissions: result,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.ctx)

			resp, err := s.StudentPermission(tc.ctx, tc.req.(*pb.StudentPermissionRequest))
			assert.Equal(t, tc.expectedErr, err)
			if tc.expectedResp == nil {
				assert.Nil(t, tc.expectedResp, resp)
			} else {
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}

func TestStudentService_GetStudentProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userID := "student-id"
	ctx = interceptors.ContextWithUserID(ctx, userID)

	userRepo := new(mock_repositories.MockUserRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)

	s := &StudentServiceABAC{
		&StudentService{
			UserRepo:    userRepo,
			StudentRepo: studentRepo,
		},
	}

	now := timeutil.Now()
	nowProto, _ := types.TimestampProto(now)
	nowProtoResp := timestamppb.New(now)
	updatedAt := pgtype.Timestamptz{}
	updatedAt.Set(now)

	student := new(entities.Student)
	err := multierr.Combine(
		student.ID.Set(userID),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.BillingDate.Set(now),
		student.Birthday.Set(now),
		student.OnTrial.Set(true),
	)
	if err != nil {
		t.Fatalf("error creating student entity: %s", err)
	}

	testCases := []TestCase{
		{
			name: "invalid argument student_id, too many student_ids",
			ctx:  ctx,
			req: &pb.GetStudentProfileRequest{
				StudentIds: make([]string, 201),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "number of ID in validStudentIDsrequest must be less than 200"),
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					getStudentProfile: func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
						return nil, status.Error(codes.InvalidArgument, "number of ID in validStudentIDsrequest must be less than 200")
					},
				}
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{StudentIds: []string{userID}},
			expectedResp: &pb.GetStudentProfileResponse{
				Datas: []*pb.GetStudentProfileResponse_Data{
					{
						Profile: &pb.StudentProfile{
							Id:        userID,
							Birthday:  nowProto,
							CreatedAt: nowProto,
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					getStudentProfile: func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
						return &upb.GetStudentProfileResponse{
							Profiles: []*upb.StudentProfile{
								{
									Id:        userID,
									Birthday:  nowProtoResp,
									CreatedAt: nowProtoResp,
								},
							},
						}, nil
					},
				}
			},
		},
		{
			name: "success with empty studentID",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{},
			expectedResp: &pb.GetStudentProfileResponse{
				Datas: []*pb.GetStudentProfileResponse_Data{
					{
						Profile: &pb.StudentProfile{
							Id:        userID,
							Birthday:  nowProto,
							CreatedAt: nowProto,
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtStudentSvc = &mockUserMgmtStudentService{
					getStudentProfile: func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
						return &upb.GetStudentProfileResponse{
							Profiles: []*upb.StudentProfile{
								{
									Id:        userID,
									Birthday:  nowProtoResp,
									CreatedAt: nowProtoResp,
								},
							},
						}, nil
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.ctx)
			resp, err := s.GetStudentProfile(tc.ctx, tc.req.(*pb.GetStudentProfileRequest))
			assert.Equal(t, tc.expectedErr, err)
			if tc.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				for i, v := range resp.Datas {
					assert.Equal(t, tc.expectedResp.(*pb.GetStudentProfileResponse).Datas[i].Profile.Id, v.Profile.Id)
					assert.Equal(t, tc.expectedResp.(*pb.GetStudentProfileResponse).Datas[i].Profile.PaymentStatus, v.Profile.PaymentStatus)
					assert.Equal(t, tc.expectedResp.(*pb.GetStudentProfileResponse).Datas[i].Profile.PlanId, v.Profile.PlanId)
					assert.Equal(t, tc.expectedResp.(*pb.GetStudentProfileResponse).Datas[i].Profile.BillingDate, (*types.Timestamp)(nil))
					assert.Equal(t, tc.expectedResp.(*pb.GetStudentProfileResponse).Datas[i].Profile.BillingAt, (*types.Timestamp)(nil))
				}
			}
		})
	}
}

type MockLearningTimeCalculator struct{}

func (m *MockLearningTimeCalculator) CalculateLearningTimeByEventLogs(ctx context.Context, studentID string, logs []*pb.StudentEventLog) error {
	return nil
}

func TestStudentService_CountTotalLOsFinished(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userID := "student-id"
	ctx = interceptors.ContextWithUserID(ctx, userID)

	invalidfrom, err := types.TimestampProto(time.Now().Add(time.Hour))
	assert.NoError(t, err)
	invalidto, err := types.TimestampProto(time.Now())
	assert.NoError(t, err)

	from := time.Now()
	to := time.Now().Add(time.Hour)

	fromTimestamp, err := types.TimestampProto(from)
	assert.NoError(t, err)
	toTimestamp, err := types.TimestampProto(to)
	assert.NoError(t, err)

	fromTimestamptz, err := database.TimestamptzFromProto(fromTimestamp)
	assert.NoError(t, err)
	toTimestamptz, err := database.TimestamptzFromProto(toTimestamp)
	assert.NoError(t, err)

	userRepo := new(mock_repositories.MockUserRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	StudentLOCompletenessRepo := new(mock_repositories.MockStudentsLearningObjectivesCompletenessRepo)

	s := &StudentServiceABAC{
		&StudentService{
			UserRepo:                  userRepo,
			StudentRepo:               studentRepo,
			StudentLOCompletenessRepo: StudentLOCompletenessRepo,
		},
	}
	testcases := []TestCase{
		{
			name:        "error invalid from and to (type date)",
			ctx:         ctx,
			req:         &pb.CountTotalLOsFinishedRequest{StudentId: userID, From: invalidfrom, To: invalidto},
			expectedErr: status.Error(codes.InvalidArgument, codes.InvalidArgument.String()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "error when count total lo finish",
			ctx:         ctx,
			req:         &pb.CountTotalLOsFinishedRequest{StudentId: userID, From: fromTimestamp, To: toTimestamp},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.StudentLOCompletenessRepo.TotalLOFinished: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				StudentLOCompletenessRepo.On("CountTotalLOsFinished", mock.Anything, mock.Anything, database.Text(userID), fromTimestamptz, toTimestamptz).Once().Return(0, pgx.ErrTxClosed)
			},
		},
		{
			name:         "happy case",
			ctx:          ctx,
			expectedResp: &pb.CountTotalLOsFinishedResponse{TotalLosFinished: 5},
			req:          &pb.CountTotalLOsFinishedRequest{StudentId: userID, From: fromTimestamp, To: toTimestamp},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				StudentLOCompletenessRepo.On("CountTotalLOsFinished", mock.Anything, mock.Anything, database.Text(userID), fromTimestamptz, toTimestamptz).Once().Return(5, nil)
			},
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup(testcase.ctx)
			resp, err := s.CountTotalLOsFinished(testcase.ctx, testcase.req.(*pb.CountTotalLOsFinishedRequest))
			assert.Equal(t, testcase.expectedErr, err)
			if testcase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				expectedResponse := testcase.expectedResp.(*pb.CountTotalLOsFinishedResponse)
				assert.Equal(t, expectedResponse, resp)
			}
		})
	}
}

func toStudentCommentPb(c *entities.StudentComment) *pb.StudentComment {
	return &pb.StudentComment{
		CommentId:      c.CommentID.String,
		StudentId:      c.StudentID.String,
		CommentContent: c.CommentContent.String,
		CoachId:        c.CoachID.String,
		UpdatedAt:      &types.Timestamp{Seconds: c.UpdatedAt.Time.Unix()},
		CreatedAt:      &types.Timestamp{Seconds: c.UpdatedAt.Time.Unix()},
	}
}
