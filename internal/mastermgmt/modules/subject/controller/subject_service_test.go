package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/subject/infrastructure/repo"
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

func TestSubjectService_ExportSubjects(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	subjectRepo := new(mock_repo.MockSubjectRepo)
	subjects := []*domain.Subject{
		{
			SubjectID: "ID 1",
			Name:      "Subject 1",
		},
		{
			SubjectID: "ID 2",
			Name:      "Subject 2",
		},
		{
			SubjectID: "ID 3",
			Name:      "Subject 3",
		},
	}

	resStr := `"subject_id","name"` + "\n" +
		`"ID 1","Subject 1"` + "\n" +
		`"ID 2","Subject 2"` + "\n" +
		`"ID 3","Subject 3"` + "\n"

	s := &SubjectService{
		ExportSubjectsQueryHandler: queries.ExportSubjectsQueryHandler{
			DB:          db,
			SubjectRepo: subjectRepo,
		},
	}

	t.Run("export all data in db with correct column", func(t *testing.T) {
		// arrange
		subjectRepo.On("GetAll", ctx, db).Once().Return(subjects, nil)

		byteData := []byte(resStr)

		// act
		resp, err := s.ExportSubjects(ctx, &mpb.ExportSubjectsRequest{})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when retrieve data failed", func(t *testing.T) {
		// arrange
		subjectRepo.On("GetAll", ctx, db).Once().Return(nil, errors.New("sample error"))

		// act
		resp, err := s.ExportSubjects(ctx, &mpb.ExportSubjectsRequest{})

		// assert
		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}

func TestSubjectService_ImportSubjects(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	subjectRepo := new(mock_repo.MockSubjectRepo)
	randID := idutil.ULIDNow()
	s := &SubjectService{
		ImportSubjectsCommandHandler: commands.ImportSubjectsCommandHandler{
			DB:          db,
			SubjectRepo: subjectRepo,
		},
	}

	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportSubjectsRequest{},
		},
		{
			name:        "invalid file - number of column != 2",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 2, got 3"),
			req: &mpb.ImportSubjectsRequest{
				Payload: []byte(`partner_internal_id,name,sequence
				1,Subject 1,2
				2,Subject 2,3
				3,Subject 3,4`),
			},
		},
		{
			name:        "invalid file - column 1 name (toLowerCase) != subject_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be subject_id, got Number"),
			req: &mpb.ImportSubjectsRequest{
				Payload: []byte(`Number,name
				1,p1
				2,p2
				3,p3`),
			},
		},
		{
			name:        "invalid file - column 2 name (toLowerCase) != name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be name, got namez"),
			req: &mpb.ImportSubjectsRequest{
				Payload: []byte(`subject_id,namez
				1,p1
				2,p2
				3,p3`),
			},
		},
		{
			name: "parsing valid file (with validation error lines in response)",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportSubjectsRequest{
				Payload: []byte(fmt.Sprintf(`subject_id,name
				,Subject 1
				ID2,Subject 2
				ID3,Subject 3
				ID4,Subject 4
				ID5,Subject 5
				%s`, fmt.Sprintf("ID6,%s", string([]byte{0xff, 0xfe, 0xfd})))),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 7",
						Description: "subject name is not a valid UTF8 string",
					},
				},
			},
			setup: func(ctx context.Context) {
				ids := []string{"ID2", "ID4", "ID4", "ID5", "ID6"}
				subjectRepo.On("GetByNames", ctx, db, ids).Once().Return(nil)
				subjectRepo.On("Import", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "invalid file - wrong values case 2",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportSubjectsRequest{
				Payload: []byte(fmt.Sprintf(`subject_id,name
				subject id 1,
				%s,name_x
				%s,name_x
				%s,name_1
				%s,name_2`, idutil.ULIDNow(), idutil.ULIDNow(), randID, randID),
				),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "subject name can not be empty",
					},
					{
						Field:       "Row Number: 6",
						Description: "id " + randID + " is duplicated",
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
			resp, err := s.ImportSubjects(testCase.ctx, testCase.req.(*mpb.ImportSubjectsRequest))
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
				mock.AssertExpectationsForObjects(t, subjectRepo)
			}
		})
	}
}
