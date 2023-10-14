package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	export_entities "github.com/manabie-com/backend/internal/payment/export_entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/export_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportService_ExportStudentBilling(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db              *mockDb.Ext
		billItemService *mockServices.IBillItemServiceForExportService
	)

	studentBillingRecords := []*export_entities.StudentBillingExport{
		{
			StudentName:     "student_name_1",
			StudentID:       "student_id_1",
			Grade:           "grade_id_1",
			Location:        "location_id_1",
			CreatedDate:     database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
			Status:          pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			BillingItemName: "billing_item_name_1",
			Courses:         "couse_1, couse_2, couse_3",
			DiscountName:    "discount_name_1",
			DiscountAmount:  float32(2),
			TaxAmount:       float32(2),
			BillingAmount:   float32(200),
		},
		{
			StudentName:     "student_name_2",
			StudentID:       "student_id_2",
			Grade:           "grade_id_2",
			Location:        "location_id_2",
			CreatedDate:     database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
			Status:          pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			BillingItemName: "billing_item_name_2",
			Courses:         "couse_1, couse_2, couse_3",
			DiscountName:    "discount_name_2",
			DiscountAmount:  float32(3),
			TaxAmount:       float32(3),
			BillingAmount:   float32(300),
		},
		{
			StudentName:     "student_name_3",
			StudentID:       "student_id_3",
			Grade:           "grade_id_3",
			Location:        "location_id_3",
			CreatedDate:     database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
			Status:          pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			BillingItemName: "billing_item_name_3",
			DiscountName:    "discount_name_3",
			DiscountAmount:  float32(4),
			TaxAmount:       float32(4),
			BillingAmount:   float32(400),
		},
	}

	testcases := []utils.TestCase{
		{
			Name: "Error on BankBranchRepo.FindExportableBankBranches",
			Ctx:  ctx,
			Req: &pb.ExportStudentBillingRequest{
				LocationIds: []string{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemService.On("GetExportStudentBilling", ctx, db, mock.Anything).Once().Return([]*export_entities.StudentBillingExport{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy Case export with bank branch records",
			Ctx:  ctx,
			Req: &pb.ExportStudentBillingRequest{
				LocationIds: []string{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemService.On("GetExportStudentBilling", ctx, db, mock.Anything).Once().Return(studentBillingRecords, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemService = new(mockServices.IBillItemServiceForExportService)
			testCase.Setup(testCase.Ctx)
			s := &ExportService{
				DB:              db,
				BillItemService: billItemService,
			}

			response, err := s.ExportStudentBilling(testCase.Ctx, testCase.Req.(*pb.ExportStudentBillingRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, response)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, response)
			}

			mock.AssertExpectationsForObjects(t, db, billItemService)

		})
	}

}
