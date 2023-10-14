package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

var (
	acID       = idutil.ULIDNow()
	taxID      = idutil.ULIDNow()
	bsID       = idutil.ULIDNow()
	bspID      = idutil.ULIDNow()
	brID       = idutil.ULIDNow()
	discountID = idutil.ULIDNow()
	feeID      = idutil.ULIDNow()
	materialID = idutil.ULIDNow()
	packageID  = idutil.ULIDNow()
	lrID       = idutil.ULIDNow()
	courseID   string
)

func (s *suite) aValidRequestPayloadForImportingAllInOneCsvFileForTest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	locationID := constants.ManabieOrgLocation
	courseIDs, err := s.insertCourses(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	courseID = courseIDs[0]

	fileData := fmt.Sprintf(
		`
					-- accounting_category
					accounting_category_id,name,remarks,is_archived
					%s,Cat 1,Remarks 1,0
					
					-- tax
					tax_id,name,tax_percentage,tax_category,default_flag,is_archived
					%s,Tax 1,11,1,0,0
					
					-- billing_schedule
					billing_schedule_id,name,remarks,is_archived
					%s,Cat 1,Remarks 1,1
					
					-- billing_schedule_period
					billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
					%s,Cat 1,%s,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
					
					-- billing_ratio
					billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
					%s,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,%s,1,2,0
					
					-- discount
					discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
					%s,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,
					
					-- fee
					fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
					%s,Cat 1,1,%s,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,%s,1,Remarks 1,0,0
					
					-- material
					material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
					%s,Cat 1,1,%s,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,%s,1,Remarks 2,0,0
					
					-- package
					package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
					%s,Package 3,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
					
					-- leaving_reason
					leaving_reason_id,name,leaving_reason_type,remark,is_archived
					%s,Cat 1,1,Remarks 1,1
					
					-- product_accounting_category
					product_id,accounting_category_id
					%s,%s
					
					-- product_discount
					product_id,discount_id
					%s,%s
					
					-- product_grade
					product_id,grade_id
					%s,1
					
					-- product_location
					product_id,location_id
					%s,%s
					
					-- product_price
					product_id,billing_schedule_period_id,quantity,price,price_type
					%s,%s,3,12.25,DEFAULT_PRICE
					
					-- product_setting
					product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
					%s,true,true,false,false
					
					-- package_course
					package_id,course_id,mandatory_flag,max_slots_per_course,course_weight
					%s,%s,0,2,3
					
					-- package_course_fee
					package_id,course_id,fee_id,available_from,available_until,is_added_by_default
					%s,%s,%s,2021-12-07,2022-12-07,true
					
					-- package_course_material
					package_id,course_id,material_id,available_from,available_until,is_added_by_default
					%s,%s,%s,2021-12-07,2022-12-07,true
					`,
		acID,
		taxID,
		bsID,
		bspID, bsID,
		brID, bspID,
		discountID,
		feeID, taxID, bsID,
		materialID, taxID, bsID,
		packageID,
		lrID,
		materialID, acID,
		feeID, discountID,
		packageID,
		packageID, locationID,
		materialID, bspID,
		feeID,
		packageID, courseID,
		packageID, courseID, feeID,
		packageID, courseID, materialID,
	)

	stepState.Request = &pb.ImportAllForTestRequest{
		Payload: []byte(fileData),
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importAllInOneCsvFileForTest(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataForTestServiceClient(s.PaymentConn).ImportAllForTest(
		contextWithToken(ctx), stepState.Request.(*pb.ImportAllForTestRequest),
	)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidAllInOneCsvFileForTestIsImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var importSuccessfully bool

	stmt := `
		SELECT exists
			(SELECT 
				ac.accounting_category_id,
				t.tax_id,
				bs.billing_schedule_id,
				bsp.billing_schedule_period_id,
				br.billing_ratio_id,
				d.discount_id,
				f.fee_id,
				m.material_id,
				p.package_id,
				lr.leaving_reason_id,
				pac.product_id,
				pd.product_id,
				pg.product_id,
				pl.product_id,
				pp.product_id,
				ps.product_id,
				pc.package_id,
				pcf.package_id,
				pcm.package_id
			FROM
				accounting_category ac,
				tax t,
				billing_schedule bs,
				billing_schedule_period bsp,
				billing_ratio br,
				discount d,
				fee f,
				material m,
				package p,
				leaving_reason lr,
				product_accounting_category pac,
				product_discount pd,
				product_grade pg,
				product_location pl,
				product_price pp,
				product_setting ps,
				package_course pc,
				package_course_fee pcf,
				package_course_material pcm
			WHERE
				ac.accounting_category_id = $1 AND
				t.tax_id = $2 AND
				bs.billing_schedule_id = $3 AND
				bsp.billing_schedule_period_id = $4 AND
				br.billing_ratio_id = $5 AND
				d.discount_id = $6 AND
				f.fee_id = $7 AND
				m.material_id = $8 AND
				p.package_id = $9 AND
				lr.leaving_reason_id = $10 AND
				pac.product_id = $11 AND
				pd.product_id = $12 AND
				pg.product_id = $13 AND
				pl.product_id = $14 AND
				pp.product_id = $15 AND
				ps.product_id = $16 AND
				pc.package_id = $17 AND
				pcf.package_id = $18 AND
				pcm.package_id = $19
			LIMIT 1);
	`

	row := s.FatimaDBTrace.QueryRow(ctx, stmt,
		acID,
		taxID,
		bsID,
		bspID,
		brID,
		discountID,
		feeID,
		materialID,
		packageID,
		lrID,
		materialID,
		feeID,
		packageID,
		packageID,
		materialID,
		feeID,
		packageID,
		packageID,
		packageID,
	)
	err := row.Scan(&importSuccessfully)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("query (to check if all records insert correctly) err: %s", err)
	}

	if !importSuccessfully {
		return StepStateToContext(ctx, stepState), fmt.Errorf("import unsuccessfully")
	}

	return StepStateToContext(ctx, stepState), nil
}
