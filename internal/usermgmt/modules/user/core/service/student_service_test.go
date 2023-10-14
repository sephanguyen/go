package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestCase struct {
	name          string
	ctx           context.Context
	req           interface{}
	expectedErr   error
	setup         func(ctx context.Context)
	setupWithMock func(ctx context.Context, mockInterface interface{})
	expectedResp  interface{}
	option        interface{}
}

func TestStudentService_UpsertStudentComment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	studentCommentRepo := new(mock_repositories.MockStudentCommentRepo)

	now := timestamppb.Now()

	s := &StudentService{
		StudentCommentRepo: studentCommentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case: create comment",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			expectedResp: &pb.UpsertStudentCommentResponse{
				Successful: true,
			},
			req: &pb.UpsertStudentCommentRequest{
				StudentComment: &pb.StudentComment{
					CommentId:      "",
					StudentId:      "studentId",
					CommentContent: "content",
					UpdatedAt:      now,
					CreatedAt:      now,
				},
			},
			setup: func(ctx context.Context) {
				studentCommentRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case: update content comment",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			expectedResp: &pb.UpsertStudentCommentResponse{
				Successful: true,
			},
			req: &pb.UpsertStudentCommentRequest{
				StudentComment: &pb.StudentComment{
					CommentId:      "commentId",
					StudentId:      "studentId",
					CommentContent: "",
					UpdatedAt:      now,
					CreatedAt:      now,
				},
			},
			setup: func(ctx context.Context) {
				studentCommentRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "err: upsert failed",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: errors.New("cannot insert new student_comments"),
			req: &pb.UpsertStudentCommentRequest{
				StudentComment: &pb.StudentComment{
					CommentId:      "commentId",
					StudentId:      "studentId",
					CommentContent: "",
					UpdatedAt:      now,
					CreatedAt:      now,
				},
			},
			setup: func(ctx context.Context) {
				studentCommentRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(errors.New("cannot insert new student_comments"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.UpsertStudentComment(testCase.ctx, testCase.req.(*pb.UpsertStudentCommentRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStudentService_DeleteStudentComments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	studentCommentRepo := new(mock_repositories.MockStudentCommentRepo)

	s := &StudentService{
		StudentCommentRepo: studentCommentRepo,
	}

	testCases := []TestCase{
		{
			name:        "delete student comments success",
			ctx:         ctx,
			expectedErr: nil,
			req: &pb.DeleteStudentCommentsRequest{
				CommentIds: []string{"comment-1", "comment-2"},
			},
			setup: func(ctx context.Context) {
				studentCommentRepo.On("DeleteStudentComments", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "delete student comments success with comment ids length = 0",
			ctx:         ctx,
			expectedErr: nil,
			req: &pb.DeleteStudentCommentsRequest{
				CommentIds: []string{},
			},
		},
		{
			name:        "error student comment ids empty",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "comment ids must not nil"),
			req: &pb.DeleteStudentCommentsRequest{
				CommentIds: nil,
			},
		},
		{
			name:        "error query database",
			ctx:         ctx,
			expectedErr: status.Errorf(codes.Internal, "error database"),
			req: &pb.DeleteStudentCommentsRequest{
				CommentIds: []string{"comment-1"},
			},
			setup: func(ctx context.Context) {
				studentCommentRepo.On("DeleteStudentComments", ctx, mock.Anything, mock.Anything).Once().Return(errors.New("error database"))
			},
		},
		{
			name:        "error database row affected",
			ctx:         ctx,
			expectedErr: nil,
			req: &pb.DeleteStudentCommentsRequest{
				CommentIds: []string{"comment-1"},
			},
			setup: func(ctx context.Context) {
				studentCommentRepo.On("DeleteStudentComments", ctx, mock.Anything, mock.Anything).Once().Return(repository.ErrUnAffected)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.DeleteStudentComments(testCase.ctx, testCase.req.(*pb.DeleteStudentCommentsRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}

}

func TestStudentService_GenerateImportStudentTemplate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := &StudentService{UnleashClient: unleashClient}

	templateImportStudentStableHeaders := "user_id,external_user_id,last_name,first_name,last_name_phonetic,first_name_phonetic,email,grade,birthday,gender,remarks,student_tag,student_phone_number,home_phone_number,contact_preference,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date,enrollment_status,location,status_start_date"
	templateImportStudentStableValues := "uuid,externaluserid,lastname,firstname,lastname_phonetic,firstname_phonetic,student@email.com,1,2000/01/12,1,remarks,tag_partner_id_1;tag_partner_id_2,123456789,123456789,1,7000000,prefecture value,city value,street1 value,street2 value,school_partner_id_1;school_partner_id_2,school_course_partner_id_1;school_course_partner_id_2,2000/01/08;2000/01/09,2000/01/10;2000/01/11,1;2,location_partner_id_1;location_partner_id_2,2000/01/12;2000/01/13"

	templateImportStudentWithUserNameHeaders := "user_id,external_user_id,username,last_name,first_name,last_name_phonetic,first_name_phonetic,email,grade,birthday,gender,remarks,student_tag,student_phone_number,home_phone_number,contact_preference,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date,enrollment_status,location,status_start_date"
	templateImportStudentWithUserNameValues := "uuid,externaluserid,username,lastname,firstname,lastname_phonetic,firstname_phonetic,student@email.com,1,2000/01/12,1,remarks,tag_partner_id_1;tag_partner_id_2,123456789,123456789,1,7000000,prefecture value,city value,street1 value,street2 value,school_partner_id_1;school_partner_id_2,school_course_partner_id_1;school_course_partner_id_2,2000/01/08;2000/01/09,2000/01/10;2000/01/11,1;2,location_partner_id_1;location_partner_id_2,2000/01/12;2000/01/13"

	type expectedResp struct {
		header string
		values string
	}

	testCases := []TestCase{
		{
			name: "stable template bulk update students",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(false, nil)
			},
			req: &pb.GenerateImportStudentTemplateRequest{},
			expectedResp: expectedResp{
				header: templateImportStudentStableHeaders,
				values: templateImportStudentStableValues,
			},
			expectedErr: nil,
		},
		{
			name: "template bulk update students with username",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			req: &pb.GenerateImportStudentTemplateRequest{},
			expectedResp: expectedResp{
				header: templateImportStudentWithUserNameHeaders,
				values: templateImportStudentWithUserNameValues,
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{Manabie: &interceptors.ManabieClaims{ResourcePath: fmt.Sprint(constants.ManabieSchool)}}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.GenerateImportStudentTemplate(testCase.ctx, testCase.req.(*pb.GenerateImportStudentTemplateRequest))
			assert.Equal(t, testCase.expectedErr, err)

			data, _ := base64.StdEncoding.DecodeString(string(resp.Data))
			expectedCSV := testCase.expectedResp.(expectedResp)
			assert.Equal(t, expectedCSV.header+"\n"+expectedCSV.values, string(data))
		})
	}

}

func TestStudentService_RetrieveStudentComment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	studentCommentRepo := new(mock_repositories.MockStudentCommentRepo)
	userRepo := new(mock_repositories.MockUserRepo)

	tNow := time.Now()

	s := &StudentService{
		StudentCommentRepo: studentCommentRepo,
		UserRepo:           userRepo,
	}
	studentId := "student-id-1"

	testCases := []TestCase{
		{
			name: "happy case: retrieve student comment",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &pb.RetrieveStudentCommentRequest{
				StudentId: studentId,
			},
			expectedResp: &pb.RetrieveStudentCommentResponse{
				Comment: []*pb.CommentInfo{
					{
						StudentComment: &pb.StudentComment{
							CreatedAt: &timestamppb.Timestamp{Seconds: tNow.Unix()},
							UpdatedAt: &timestamppb.Timestamp{Seconds: tNow.Unix()},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentCommentRepo.On("RetrieveByStudentID", ctx, mock.Anything, database.Text("student-id-1"), mock.Anything).Once().Return([]entity.StudentComment{{
					UpdatedAt: pgtype.Timestamptz{Time: tNow, Status: 2},
					CreatedAt: pgtype.Timestamptz{Time: tNow, Status: 2},
				}}, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{
					{},
				}, nil)
			},
		},
		{
			name: "error student id nil or empty",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id-1"),
			req: &pb.RetrieveStudentCommentRequest{
				StudentId: "",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "studentId is cannot empty or nil"),
		},
		{
			name: "error invalid student id",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id-1"),
			req: &pb.RetrieveStudentCommentRequest{
				StudentId: studentId,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "student id invalid"),
			setup: func(ctx context.Context) {
				studentCommentRepo.On("RetrieveByStudentID", ctx, mock.Anything, database.Text("student-id-1"), mock.Anything).Once().Return(nil,
					errors.New("error retrieve comments by student id"))
			},
		},
		{
			name: "error cannot find coach",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id-1"),
			req: &pb.RetrieveStudentCommentRequest{
				StudentId: studentId,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "cannot find coach: error from user repo"),
			setup: func(ctx context.Context) {
				studentCommentRepo.On("RetrieveByStudentID", ctx, mock.Anything, database.Text("student-id-1"), mock.Anything).Once().Return([]entity.StudentComment{{
					UpdatedAt: pgtype.Timestamptz{Time: tNow, Status: 2},
					CreatedAt: pgtype.Timestamptz{Time: tNow, Status: 2},
				}}, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, errors.New("error from user repo"))
			},
		},
		{
			name: "error length coaches is 0",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id-1"),
			req: &pb.RetrieveStudentCommentRequest{
				StudentId: studentId,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "the coach id is non-existing"),
			setup: func(ctx context.Context) {
				studentCommentRepo.On("RetrieveByStudentID", ctx, mock.Anything, database.Text("student-id-1"), mock.Anything).Once().Return([]entity.StudentComment{{
					UpdatedAt: pgtype.Timestamptz{Time: tNow, Status: 2},
					CreatedAt: pgtype.Timestamptz{Time: tNow, Status: 2},
				}}, nil)
				var retrieveUserResponse []*entity.LegacyUser
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(retrieveUserResponse, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
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

func TestStudentService_GetStudentProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userID := "student-id"
	ctx = interceptors.ContextWithUserID(ctx, userID)

	studentRepo := new(mock_repositories.MockStudentRepo)

	s := &StudentService{
		StudentRepo: studentRepo,
	}

	now := timeutil.Now()
	nowProto := timestamppb.New(now)

	student := new(entity.LegacyStudent)
	err := multierr.Combine(
		student.ID.Set(userID),
		student.Birthday.Set(now),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
	)
	if err != nil {
		t.Fatalf("error creating student entity: %s", err)
	}

	gradeName := "Grade 1"
	grade := &repository.GradeEntity{
		ID:   database.Text("grad-id"),
		Name: database.Text(gradeName),
	}
	if err != nil {
		t.Fatalf("error creating student entity: %s", err)
	}
	students := []repository.StudentProfile{
		{
			Student: *student,
		},
	}

	studentsWithGradeInfo := []repository.StudentProfile{
		{
			Student: *student,
			Grade:   *grade,
		},
	}

	testCases := []TestCase{
		{
			name: "success: get another student assigned profile",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{StudentIds: []string{"1"}},
			expectedResp: &pb.GetStudentProfileResponse{
				Profiles: []*pb.StudentProfile{
					{
						Id:        userID,
						Birthday:  nowProto,
						CreatedAt: nowProto,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{"1"})).Once().Return(students, nil)
			},
		},
		{
			name: "success: get my profile",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{StudentIds: []string{userID}},
			expectedResp: &pb.GetStudentProfileResponse{
				Profiles: []*pb.StudentProfile{
					{
						Id:        userID,
						Birthday:  nowProto,
						CreatedAt: nowProto,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userID})).Once().Return(students, nil)
			},
		},
		{
			name: "success: with empty studentID",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{},
			expectedResp: &pb.GetStudentProfileResponse{
				Profiles: []*pb.StudentProfile{
					{
						Id:        userID,
						Birthday:  nowProto,
						CreatedAt: nowProto,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userID})).Once().Return(students, nil)
			},
		},
		{
			name:         "failed: invalid argument student_id, too many student_ids",
			ctx:          ctx,
			req:          &pb.GetStudentProfileRequest{StudentIds: make([]string, 201)},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "number of ID in validStudentIDs request must be less than 200"),
			setup:        func(ctx context.Context) {},
		},
		{
			name:         "failed: get my profile err",
			ctx:          ctx,
			req:          &pb.GetStudentProfileRequest{StudentIds: []string{userID}},
			expectedResp: nil,
			expectedErr:  toStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userID})).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "success: get student profile with correct grade info",
			ctx:  ctx,
			req:  &pb.GetStudentProfileRequest{StudentIds: []string{userID}},
			expectedResp: &pb.GetStudentProfileResponse{
				Profiles: []*pb.StudentProfile{
					{
						Id:        userID,
						Birthday:  nowProto,
						CreatedAt: nowProto,
						GradeName: gradeName,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{userID})).Once().Return(studentsWithGradeInfo, nil)
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
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}
