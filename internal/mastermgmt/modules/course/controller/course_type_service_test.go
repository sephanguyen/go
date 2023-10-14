package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_course_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/course/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCourseTypeService_ImportCourseTypes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	courseTypeRepo := &mock_course_repo.MockCourseTypeRepo{}
	s := &CourseTypeService{
		DB: db,
		CourseTypeCommandHandler: commands.CourseTypeCommandHandler{
			DB:             db,
			CourseTypeRepo: courseTypeRepo,
		},
	}

	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportCourseTypesRequest{},
		},
		{
			name:        "invalid file - number of column != 4",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 4, got 3"),
			req: &mpb.ImportCourseTypesRequest{
				Payload: []byte(`course_type_id,course_type_name,is_archived
				1,Course 1,0`),
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != course_type_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be course_type_id, got Number"),
			req: &mpb.ImportCourseTypesRequest{
				Payload: []byte(`Number,course_type_name,is_archived,remarks
				1,Course 1,1,m`),
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != course_type_name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be course_type_name, got name"),
			req: &mpb.ImportCourseTypesRequest{
				Payload: []byte(`course_type_id,name,is_archived,remarks
				1,Course 1,1,m`),
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != is_archived",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be is_archived, got type_id"),
			req: &mpb.ImportCourseTypesRequest{
				Payload: []byte(`course_type_id,course_type_name,type_id,remarks
				1,Course 1,1,m`),
			},
		},
		{
			name:        "invalid file - fourth column name (toLowerCase) != remarks",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 4 should be remarks, got marks"),
			req: &mpb.ImportCourseTypesRequest{
				Payload: []byte(`course_type_id,course_type_name,is_archived,marks
				1,Course 1,1,m`),
			},
		},
		{
			name: "parsing valid file with invalid values",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportCourseTypesRequest{
				Payload: []byte(fmt.Sprintf(`course_type_id,course_type_name,is_archived,remarks
				1,Course 1,bool,m
				99,,1,m
				22,Course xyz,0,
				%s`, fmt.Sprintf("12,%s,1,m", string([]byte{0xff, 0xfe, 0xfd})))),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "bool is not a valid boolean: strconv.ParseBool: parsing \"bool\": invalid syntax",
					},
					{
						Field:       "Row Number: 3",
						Description: "name can not be empty",
					},
					{
						Field:       "Row Number: 5",
						Description: `name is not a valid UTF8 string`,
					},
				},
			},
		},
		{
			name: "valid file with valid values should be imported",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportCourseTypesRequest{
				Payload: []byte(`course_type_id,course_type_name,is_archived,remarks
				1,Course type 1,1,
				2,Course type 2,0,note`),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				expectedCourseTypes := []*domain.CourseType{
					{
						CourseTypeID: "1",
						Name:         "Course type 1",
						IsArchived:   true,
						Remarks:      "",
					},
					{
						CourseTypeID: "2",
						Name:         "Course type 2",
						IsArchived:   false,
						Remarks:      "note",
					},
				}
				db.On("Begin", ctx).Once().Return(tx, nil)
				courseTypeRepo.On("Import", ctx, db, expectedCourseTypes).Return(nil)
				tx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportCourseTypes(testCase.ctx, testCase.req.(*mpb.ImportCourseTypesRequest))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, err, testCase.expectedErr)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, courseTypeRepo)
			}
		})
	}
}
