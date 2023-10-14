package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

type mockForTestRepo struct{}

func (s *mockForTestRepo) Insert(ctx context.Context, db database.QueryExecer, e database.Entity, excludedFields []string) error {
	return nil
}

func TestImportAllForTest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	forTestRepo := &mockForTestRepo{}

	s := &ImportMasterDataForTestService{
		DB:          db,
		ForTestRepo: forTestRepo,
	}

	testcases := []TestCase{
		{
			name: "one and only one ut (happy case)",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &pb.ImportAllForTestRequest{
				Payload: []byte(
					`
					-- accounting_category
					accounting_category_id,name,remarks,is_archived
					1,Cat 1,Remarks 1,0
					
					-- tax
					tax_id,name,tax_percentage,tax_category,default_flag,is_archived
					1,Tax 1,11,1,0,0
					
					-- billing_schedule
					billing_schedule_id,name,remarks,is_archived
					1,Cat 1,Remarks 1,1
					
					-- billing_schedule_period
					billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
					1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
					
					-- billing_ratio
					billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
					1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,1,1,2,0
					
					-- discount
					discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
					1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,
					
					-- fee
					fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
					1,Cat 1,1,1,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,1,1,Remarks 1,0,0
					
					-- material
					material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
					2,Cat 1,1,1,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,1,1,Remarks 2,0,0
					
					-- package
					package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
					3,Package 3,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
					
					-- leaving_reason
					leaving_reason_id,name,leaving_reason_type,remark,is_archived
					1,Cat 1,1,Remarks 1,1
					
					-- product_accounting_category
					product_id,accounting_category_id
					1,1
					
					-- product_discount
					product_id,discount_id
					1,1
					
					-- product_grade
					product_id,grade_id
					1,1
					
					-- product_location
					product_id,location_id
					1,1
					
					-- product_price
					product_id,billing_schedule_period_id,quantity,price,price_type
					1,1,3,12.25,DEFAULT_PRICE
					
					-- product_setting
					product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
					1,true,true,false,false
					
					-- package_course
					package_id,course_id,mandatory_flag,max_slots_per_course,course_weight
					3,1,0,2,3
					
					-- package_course_fee
					package_id,course_id,fee_id,available_from,available_until,is_added_by_default
					3,1,1,2021-12-07,2022-12-07,false
					
					-- package_course_material
					package_id,course_id,material_id,available_from,available_until,is_added_by_default
					3,1,2,2021-12-07,2022-12-07,true

					-- notification_date
					notification_date_id,order_type,notification_date,is_archived
					1,ORDER_TYPE_NEW,10,0
					`),
			},
			expectedResp: &pb.ImportAllForTestResponse{
				Errors: []*pb.ImportAllForTestResponse_ImportAllForTestError{},
			},
			setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.ImportAllForTest(testCase.ctx, testCase.req.(*pb.ImportAllForTestRequest))
			assert.Nil(t, err)

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*pb.ImportAllForTestResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
		})
	}
}
