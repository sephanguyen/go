package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/grade/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name             string
	ctx              context.Context
	req              interface{}
	expectedResp     interface{}
	expectedErr      error
	setup            func(ctx context.Context)
	expectedErrModel *errdetails.BadRequest
}

func TestGradeService_ImportGrades(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	gradeRepo := new(mock_repo.MockGradeRepo)
	gradeRandID := idutil.ULIDNow()
	s := &GradeService{
		ImportGradesCommandHandler: commands.ImportGradesCommandHandler{
			DB:        db,
			GradeRepo: gradeRepo,
		},
	}

	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportGradesRequest{},
		},
		{
			name:        "invalid file - number of column != 6",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 5, got 3"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(`partner_internal_id,name,sequence
				1,Grade 1,2
				2,Grade 2,3
				3,Grade 3,4`),
			},
		},
		{
			name:        "invalid file - column 1 name (toLowerCase) != grade_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be grade_id, got Number"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(`Number,grade_partner_id,name,sequence,remarks
				1,p1,Grade 1,1,note
				2,p2,Grade 2,2,
				3,p3,Grade 3,3,note2`),
			},
		},
		{
			name:        "invalid file - column 2 name (toLowerCase) != grade_partner_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be grade_partner_id, got partner_id"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(`grade_id,partner_id,name,sequence,remarks
				1,p1,Grade 1,1,note
				2,p2,Grade 2,2,
				3,p3,Grade 3,3,note2`),
			},
		},
		{
			name:        "invalid file - column 3 name (toLowerCase) != name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be name, got namez"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(`grade_id,grade_partner_id,namez,sequence,remarks
				1,p1,Grade 1,1,note
				2,p2,Grade 2,2,
				3,p3,Grade 3,3,note2`),
			},
		},
		{
			name:        "invalid file - column 4 name (toLowerCase) != sequence",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 4 should be sequence, got sequencez"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(`grade_id,grade_partner_id,name,sequencez,remarks
				1,p1,Grade 1,1,note
				2,p2,Grade 2,2,
				3,p3,Grade 3,3,note2`),
			},
		},
		{
			name:        "invalid file - column5 name (toLowerCase) != remarks",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 5 should be remarks, got remarksz"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(`grade_id,grade_partner_id,name,sequence,remarksz
				1,p1,Grade 1,1,note
				2,p2,Grade 2,2,
				3,p3,Grade 3,3,note2`),
			},
		},
		{
			name: "parsing valid file (with validation error lines in response)",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(fmt.Sprintf(`grade_id,grade_partner_id,name,sequence,remarks
				,1,Grade 1 x,bool,1,note
				ID2,2,,2,note
				ID3,,Grade 1 sample,1,3,note
				ID4,4,Grade 2,1,note
				ID5,4,Grade x,5,note
				%s`, fmt.Sprintf("ID6,6,%s,0,20,notes", string([]byte{0xff, 0xfe, 0xfd})))),
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
						Description: "grade name can not be empty",
					},
					{
						Field:       "Row Number: 4",
						Description: "column grade_partner_id is required",
					},
					{
						Field:       "Row Number: 5",
						Description: "sequence 1 is duplicated",
					},
					{
						Field:       "Row Number: 6",
						Description: "grade partner id 4 is duplicated",
					},
					{
						Field:       "Row Number: 7",
						Description: "grade name is not a valid UTF8 string",
					},
				},
			},
			setup: func(ctx context.Context) {
				gradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil)
				gradeRepo.On("Import", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "invalid file - wrong values case 2",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportGradesRequest{
				Payload: []byte(fmt.Sprintf(`grade_id,grade_partner_id,name,sequence,remarks
				grade id 1,1234,,2345,updated remarks
				%s,1235,new name %s,2346,updated remarks
				%s,1236,new name-1 %s,2345,remarks`, gradeRandID, gradeRandID, gradeRandID, gradeRandID)),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "grade name can not be empty",
					},
					{
						Field:       "Row Number: 4",
						Description: "sequence 2345 is duplicated",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportGrades(testCase.ctx, testCase.req.(*mpb.ImportGradesRequest))
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
				mock.AssertExpectationsForObjects(t, gradeRepo)
			}
		})
	}
}

func TestGradeService_ExportGrades(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	gradeRepo := new(mock_repo.MockGradeRepo)
	mark1 := "marks 1"
	grades := []*domain.Grade{
		{
			ID:                "ID 1",
			Name:              "Grade 1",
			PartnerInternalID: "Partner 1",
			Sequence:          1,
			Remarks:           mark1,
		},
		{
			ID:                "ID 2",
			Name:              "Grade 2",
			PartnerInternalID: "Partner 2",
			Sequence:          2,
		},
		{
			ID:                "ID 3",
			Name:              "Grade 3",
			PartnerInternalID: "Partner 3",
			Sequence:          3,
		},
	}

	gradeStr := `"grade_id","grade_partner_id","name","sequence","remarks"` + "\n" +
		`"ID 1","Partner 1","Grade 1","1","marks 1"` + "\n" +
		`"ID 2","Partner 2","Grade 2","2",""` + "\n" +
		`"ID 3","Partner 3","Grade 3","3",""` + "\n"

	s := &GradeService{
		ExportGradesQueryHandler: queries.ExportGradesQueryHandler{
			DB:        db,
			GradeRepo: gradeRepo,
		},
	}

	t.Run("export all data in db with correct column", func(t *testing.T) {
		// arrange
		gradeRepo.On("GetAll", ctx, db).Once().Return(grades, nil)

		byteData := []byte(gradeStr)

		// act
		resp, err := s.ExportGrades(ctx, &mpb.ExportGradesRequest{})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when retrieve data failed", func(t *testing.T) {
		// arrange
		gradeRepo.On("GetAll", ctx, db).Once().Return(nil, errors.New("sample error"))

		// act
		resp, err := s.ExportGrades(ctx, &mpb.ExportGradesRequest{})

		// assert
		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}
