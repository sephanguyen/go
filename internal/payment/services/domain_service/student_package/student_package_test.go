package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pmpb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var mapVal interface{}

func TestStudentPackage_getStartTimeFromOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	now := time.Now().UTC()

	testcases := []utils.TestCase{
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				&pmpb.OrderItem{
					StartDate:     timestamppb.New(now),
					EffectiveDate: timestamppb.New(now.AddDate(0, 1, 0)),
				},
			},
			ExpectedResp: now,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				&pmpb.OrderItem{
					EffectiveDate: timestamppb.New(now.AddDate(0, 1, 0)),
				},
			},
			ExpectedResp: now.AddDate(0, 1, 0),
			Setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			orderItemReq := testCase.Req.([]interface{})[0].(*pmpb.OrderItem)

			resp := getStartTimeFromOrder(orderItemReq)
			assert.Equal(t, testCase.ExpectedResp, resp)

		})
	}
}

func TestStudentPackage_writeStudentPackageLog(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentPackageLogRepo *mockRepositories.MockStudentPackageLogRepo
	)
	now := time.Now().UTC()
	packageProperties := entities.PackageProperties{
		AllCourseInfo: []entities.CourseInfo{
			{
				CourseID:      constant.CourseID,
				Name:          constant.CourseName,
				NumberOfSlots: 1,
				Weight:        6,
			},
		},
		CanWatchVideo:     []string{constant.CourseID},
		CanViewStudyGuide: []string{constant.CourseID},
		CanDoQuiz:         []string{constant.CourseID},
		LimitOnlineLesson: 0,
		AskTutor: &entities.AskTutorCfg{
			TotalQuestionLimit: 0,
			LimitDuration:      "",
		},
	}
	packagePropertiesJson, _ := json.Marshal(packageProperties)
	studentPackageObject := entities.StudentPackages{
		ID: pgtype.Text{
			String: constant.StudentPackageID,
			Status: pgtype.Present,
		},
		StudentID: pgtype.Text{
			String: constant.StudentID,
			Status: pgtype.Present,
		},
		PackageID: pgtype.Text{
			String: constant.PackageID,
			Status: pgtype.Present,
		},
		StartAt: pgtype.Timestamptz{
			Time:   now,
			Status: pgtype.Present,
		},
		EndAt: pgtype.Timestamptz{
			Time:   now.AddDate(0, 4, 0),
			Status: pgtype.Present,
		},
		Properties: pgtype.JSONB{
			Bytes:  packagePropertiesJson,
			Status: pgtype.Present,
		},
		IsActive: pgtype.Bool{
			Bool:   false,
			Status: pgtype.Present,
		},
		LocationIDs: pgtype.TextArray{
			Elements: []pgtype.Text{
				{
					String: constant.LocationID,
					Status: pgtype.Present,
				},
			},
			Status: pgtype.Present,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:             now,
			Status:           pgtype.Present,
			InfinityModifier: 0,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:   now,
			Status: pgtype.Present,
		},
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
	}
	var testcases = []utils.TestCase{
		{
			Name: "Fail case: error when create student package log",
			Ctx:  ctx,
			Req: []interface{}{
				&studentPackageObject,
				constant.CourseID,
				"upsert student package",
				"flow upsert student package",
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageService{
				StudentPackageLogRepo: studentPackageLogRepo,
			}
			studentPackage := testCase.Req.([]interface{})[0].(*entities.StudentPackages)
			courseID := testCase.Req.([]interface{})[1].(string)
			action := testCase.Req.([]interface{})[2].(string)
			flow := testCase.Req.([]interface{})[3].(string)

			err := s.writeStudentPackageLog(testCase.Ctx, db, studentPackage, courseID, action, flow)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageLogRepo)
		})
	}
}
