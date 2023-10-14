package services

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockStudentReaderService struct {
	retrieveStudentAssociatedToParentAccount func(ctx context.Context, in *upb.RetrieveStudentAssociatedToParentAccountRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentAssociatedToParentAccountResponse, error)
}

func (m *mockStudentReaderService) RetrieveStudentAssociatedToParentAccount(ctx context.Context, in *upb.RetrieveStudentAssociatedToParentAccountRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentAssociatedToParentAccountResponse, error) {
	return m.retrieveStudentAssociatedToParentAccount(ctx, in, opts...)
}

func TestStudentReader_FindStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	studentRepository := &mock_repositories.MockStudentRepo{}
	s := &StudentReaderService{
		studentRepository: studentRepository,
		db:                db,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.FindStudentRequest{},
			expectedErr: status.Error(codes.Unimplemented, fmt.Sprintln("Method has not implemented yet")),
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.FindStudentRequest)
			_, err := s.FindStudent(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestStudentReader_RetrieveStudentProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	s := &StudentReaderService{
		studentRepository: studentRepo,
		db:                db,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	now := time.Now()
	timeProto := timestamppb.New(now)

	testCases := []TestCase{
		{
			name: "invalid argument student_id, too many student_ids",
			ctx:  ctx,
			req: &bpb.RetrieveStudentProfileRequest{
				StudentIds: make([]string, 201),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "number of ID in validStudentIDsrequest must be less than 200"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "get my profile err",
			ctx:          ctx,
			req:          &bpb.RetrieveStudentProfileRequest{StudentIds: []string{userId}},
			expectedResp: nil,
			expectedErr:  toStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userId})).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "get my profile",
			ctx:  ctx,
			req:  &bpb.RetrieveStudentProfileRequest{StudentIds: []string{userId}},
			expectedResp: &bpb.RetrieveStudentProfileResponse{
				Items: []*bpb.RetrieveStudentProfileResponse_Data{
					{
						Profile: &bpb.StudentProfile{Id: userId, Birthday: timeProto, CreatedAt: timeProto, LastLoginDate: timeProto},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				student := new(entities.Student)
				student.ID.Set(userId)
				student.CreatedAt.Set(now)
				student.UpdatedAt.Set(now)
				student.BillingDate.Set(now)
				student.Birthday.Set(now)
				student.LastLoginDate.Set(now)

				students := []repositories.StudentProfile{{
					Student: *student,
				}}
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userId})).Once().Return(students, nil)
			},
		},
		{
			name: "get my profile with empty student ID",
			ctx:  ctx,
			req:  &bpb.RetrieveStudentProfileRequest{},
			expectedResp: &bpb.RetrieveStudentProfileResponse{
				Items: []*bpb.RetrieveStudentProfileResponse_Data{
					{
						Profile: &bpb.StudentProfile{Id: userId, Birthday: timeProto, CreatedAt: timeProto, LastLoginDate: timeProto},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				student := new(entities.Student)
				student.ID.Set(userId)
				student.CreatedAt.Set(now)
				student.UpdatedAt.Set(now)
				student.BillingDate.Set(now)
				student.Birthday.Set(now)
				student.LastLoginDate.Set(now)

				students := []repositories.StudentProfile{{
					Student: *student,
				}}
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userId})).Once().Return(students, nil)
			},
		},
		{
			name: "get profile that has null last login date",
			ctx:  ctx,
			req:  &bpb.RetrieveStudentProfileRequest{StudentIds: []string{userId}},
			expectedResp: &bpb.RetrieveStudentProfileResponse{
				Items: []*bpb.RetrieveStudentProfileResponse_Data{
					{
						Profile: &bpb.StudentProfile{Id: userId, Birthday: timeProto, CreatedAt: timeProto},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				student := new(entities.Student)
				student.ID.Set(userId)
				student.CreatedAt.Set(now)
				student.UpdatedAt.Set(now)
				student.BillingDate.Set(now)
				student.Birthday.Set(now)

				students := []repositories.StudentProfile{{
					Student: *student,
				}}
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userId})).Once().Return(students, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveStudentProfile(testCase.ctx, testCase.req.(*bpb.RetrieveStudentProfileRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudentReader_RetrieveLearningProgress(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	studentRepository := &mock_repositories.MockStudentRepo{}
	s := &StudentReaderService{
		studentRepository: studentRepository,
		db:                db,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.RetrieveLearningProgressRequest{},
			expectedErr: status.Error(codes.Unimplemented, fmt.Sprintln("Method has not implemented yet")),
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.RetrieveLearningProgressRequest)
			_, err := s.RetrieveLearningProgress(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestStudentReader_RetrieveStat(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	studentRepository := &mock_repositories.MockStudentRepo{}
	s := &StudentReaderService{
		studentRepository: studentRepository,
		db:                db,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.RetrieveStatRequest{},
			expectedErr: status.Error(codes.Unimplemented, fmt.Sprintln("Method has not implemented yet")),
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.RetrieveStatRequest)
			_, err := s.RetrieveStat(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestStudentReader_RetrieveStudentAssociatedToParentAccount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	student := generateStudent()
	studentRepository := &mock_repositories.MockStudentRepo{}
	s := &StudentReaderService{
		studentRepository: studentRepository,
		db:                db,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req:  &bpb.RetrieveStudentAssociatedToParentAccountRequest{},
			setup: func(ctx context.Context) {
				s.UserReaderStudentSvc = &mockStudentReaderService{
					retrieveStudentAssociatedToParentAccount: func(ctx context.Context, in *upb.RetrieveStudentAssociatedToParentAccountRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentAssociatedToParentAccountResponse, error) {
						return &upb.RetrieveStudentAssociatedToParentAccountResponse{
							Profiles: []*cpb.BasicProfile{student},
						}, nil
					},
				}
			},
			expectedResp: []*cpb.BasicProfile{
				{
					Name:    student.Name,
					Group:   student.Group,
					Country: student.Country,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "error query",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.RetrieveStudentAssociatedToParentAccountRequest{},
			expectedErr: fmt.Errorf("[StudentReaderService]:[retrieve student associated to parent account]:%v", fmt.Errorf("error query")),
			setup: func(ctx context.Context) {
				s.UserReaderStudentSvc = &mockStudentReaderService{
					retrieveStudentAssociatedToParentAccount: func(ctx context.Context, in *upb.RetrieveStudentAssociatedToParentAccountRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentAssociatedToParentAccountResponse, error) {
						return nil, fmt.Errorf("error query")
					},
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.RetrieveStudentAssociatedToParentAccountRequest)
			resp, err := s.RetrieveStudentAssociatedToParentAccount(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				expectedResp := testCase.expectedResp.([]*cpb.BasicProfile)
				assert.Equal(t, len(expectedResp), len(resp.Profiles))
				assert.Equal(t, expectedResp[0].GivenName, resp.Profiles[0].GivenName)
				assert.Equal(t, expectedResp[0].Group, resp.Profiles[0].Group)
				assert.Equal(t, expectedResp[0].Country, resp.Profiles[0].Country)
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, studentRepository)
		})
	}
}

func TestStudentReader_GetListSchoolIDsByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	s := &StudentReaderService{
		studentRepository: studentRepo,
		db:                db,
	}

	userId := idutil.ULIDNow()
	userId2 := idutil.ULIDNow()
	schoolId := 1
	ctx = interceptors.ContextWithUserID(ctx, userId)

	now := time.Now()

	testCases := []TestCase{
		{
			name:         "get my profile err",
			ctx:          ctx,
			req:          &bpb.GetListSchoolIDsByStudentIDsRequest{StudentIds: []string{userId}},
			expectedResp: nil,
			expectedErr:  toStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userId})).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "get my profile",
			ctx:  ctx,
			req:  &bpb.GetListSchoolIDsByStudentIDsRequest{StudentIds: []string{userId}},
			expectedResp: &bpb.GetListSchoolIDsByStudentIDsResponse{
				SchoolIds: []*bpb.SchoolIDWithStudentIDs{
					{
						SchoolId:   fmt.Sprint(schoolId),
						StudentIds: []string{userId},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				student := new(entities.Student)
				student.ID.Set(userId)
				student.CreatedAt.Set(now)
				student.UpdatedAt.Set(now)
				student.BillingDate.Set(now)
				student.Birthday.Set(now)
				student.LastLoginDate.Set(now)
				school := new(entities.School)
				school.ID.Set(schoolId)

				students := []repositories.StudentProfile{{
					Student: *student,
					School:  *school,
				}}
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userId})).Once().Return(students, nil)
			},
		},
		{
			name: "get 2 user profile",
			ctx:  ctx,
			req:  &bpb.GetListSchoolIDsByStudentIDsRequest{StudentIds: []string{userId}},
			expectedResp: &bpb.GetListSchoolIDsByStudentIDsResponse{
				SchoolIds: []*bpb.SchoolIDWithStudentIDs{
					{
						SchoolId:   fmt.Sprint(schoolId),
						StudentIds: []string{userId, userId2},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				student := new(entities.Student)
				student.ID.Set(userId)
				student.CreatedAt.Set(now)
				student.UpdatedAt.Set(now)
				student.BillingDate.Set(now)
				student.Birthday.Set(now)
				student.LastLoginDate.Set(now)
				student2 := new(entities.Student)
				student2.ID.Set(userId2)
				student2.CreatedAt.Set(now)
				student2.UpdatedAt.Set(now)
				student2.BillingDate.Set(now)
				student2.Birthday.Set(now)
				student2.LastLoginDate.Set(now)
				school := new(entities.School)
				school.ID.Set(schoolId)

				students := []repositories.StudentProfile{{
					Student: *student,
					School:  *school,
				}, {
					Student: *student2,
					School:  *school,
				}}
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userId})).Once().Return(students, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.GetListSchoolIDsByStudentIDs(testCase.ctx, testCase.req.(*bpb.GetListSchoolIDsByStudentIDsRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func generateStudent() *cpb.BasicProfile {
	rand.Seed(time.Now().UnixNano())
	e := new(cpb.BasicProfile)
	e.Avatar = fmt.Sprintf("http://avatar-%d", rand.Int())
	e.Group = cpb.UserGroup_USER_GROUP_STUDENT
	e.Name = fmt.Sprintf("student %d", rand.Int())
	e.Country = cpb.Country_COUNTRY_VN
	e.CreatedAt = timestamppb.Now()

	return e
}

func TestStudentReaderService_RetrieveStudentSchoolHistory(t *testing.T) {
	db := &mock_database.Ext{}

	schoolHistoryRepo := &mock_repositories.MockSchoolHistoryRepo{}

	s := &StudentReaderService{
		SchoolHistoryRepository: schoolHistoryRepo,
		db:                      db,
	}

	testCases := []TestCase{
		{
			name: "Happy case",
			ctx:  context.Background(),
			req: &bpb.RetrieveStudentSchoolHistoryRequest{
				StudentIds: []string{"whatever"},
			},
			expectedResp: &bpb.RetrieveStudentSchoolHistoryResponse{
				Schools: map[string]*bpb.RetrieveStudentSchoolHistoryResponse_School{
					"school_id_1": {
						SchoolId: "school_id_1",
					},
				},
			},
			setup: func(ctx context.Context) {
				studentInfos := []*repositories.StudentSchoolInfo{
					{
						SchoolID:   database.Text("school_id_1"),
						SchoolName: database.Text("school_name"),
						StudentID:  database.Text("whatever"),
					},
				}

				schoolHistoryRepo.On("GetCurrentSchoolInfoByStudentIDs", mock.Anything, db, mock.Anything).Once().Return(studentInfos, nil)
			},
		},
		{
			name: "Error case",
			ctx:  context.Background(),
			req:  &bpb.RetrieveStudentSchoolHistoryRequest{},
			setup: func(ctx context.Context) {
				schoolHistoryRepo.On("GetCurrentSchoolInfoByStudentIDs", mock.Anything, db, mock.Anything).Once().Return([]*repositories.StudentSchoolInfo{}, fmt.Errorf("rows.Err:"))
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(fmt.Errorf("rows.Err:"), "s.SchoolHistoryRepository.GetCurrentSchoolInfoByStudentIDs ").Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			req := testCase.req.(*bpb.RetrieveStudentSchoolHistoryRequest)

			resp, err := s.RetrieveStudentSchoolHistory(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				expectedResp := testCase.expectedResp.(*bpb.RetrieveStudentSchoolHistoryResponse)
				assert.Equal(t, len(expectedResp.GetSchools()), len(resp.GetSchools()))
			}

			mock.AssertExpectationsForObjects(t, db, schoolHistoryRepo)
		})
	}
}
