package openapisvc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_common "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/common"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateBillingAddressEventInfo struct {
	StudentID   string
	PayerName   string
	UserAddress *upb.UserAddress
}

type UpdateBillingProcessResult struct {
	StudentPaymentDetailEntity         *entities.StudentPaymentDetail
	BillingAddressEntity               *entities.BillingAddress
	NeedBillingAddressUpsert           bool
	NeedStudentPaymentDetailUpsert     bool
	StudentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType
}

func (s *OpenAPIModifierService) AutoUpdateBillingAddressInfoAndPaymentDetail(ctx context.Context, info *UpdateBillingAddressEventInfo) error {
	err := validateUpdateBillingAddressEventInfo(info)
	if err != nil {
		s.logger.Warnf("validation error for student %s err: %v", info.StudentID, err)
		return nil
	}

	studentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, s.DB, info.StudentID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return status.Error(codes.Internal, fmt.Sprintf("StudentPaymentDetailRepo.FindByStudentID: %v", err))
	}

	billingAddress, err := s.BillingAddressRepo.FindByUserID(ctx, s.DB, info.StudentID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return status.Error(codes.Internal, fmt.Sprintf("BillingAddressRepo.FindByUserID: %v", err))
	}

	return database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		return s.updateBillingInfoAndPaymentDetailFromEvent(ctx, tx, studentPaymentDetail, billingAddress, info)
	})
}

func (s *OpenAPIModifierService) updateBillingInfoAndPaymentDetailFromEvent(
	ctx context.Context,
	db database.QueryExecer,
	studentPaymentDetail *entities.StudentPaymentDetail,
	billingAddress *entities.BillingAddress,
	info *UpdateBillingAddressEventInfo,
) error {
	var (
		res *UpdateBillingProcessResult
		err error
	)

	studentPaymentActionLogBillingInfo := genActionLogInfoWithBillingAddress()

	hasEmptyBillingEventMessage := info.UserAddress == nil || hasEmptyField([]string{
		info.UserAddress.PostalCode,
		info.UserAddress.City,
		info.UserAddress.Prefecture,
	})

	switch {
	case billingAddress == nil && studentPaymentDetail == nil: // if billing address and student payment detail is null, create the billing address and payment detail
		// if user's home address is not set or has empty fields, do nothing
		if hasEmptyBillingEventMessage {
			s.logger.Infof("update skipped for student %s due to empty home address", info.StudentID)
			return nil
		}

		res, err = s.createNewBillingAndPaymentDetailEntity(ctx, db, info)
		if err != nil {
			return err
		}
	case studentPaymentDetail != nil && billingAddress == nil: // if only the billing address is nil, create billing address and update payer name
		// if user's home address is not set or has empty fields, do nothing
		if hasEmptyBillingEventMessage {
			s.logger.Infof("update skipped for student %s due to empty home address", info.StudentID)
			return nil
		}

		res, err = s.createNewBillingAndUpdatePayerNameEntity(ctx, db, studentPaymentDetail, info, studentPaymentActionLogBillingInfo)
		if err != nil {
			return err
		}

		err = s.resetThePaymentMethodIfEmpty(ctx, db, studentPaymentDetail, res)
		if err != nil {
			return err
		}
	case studentPaymentDetail != nil && billingAddress != nil && hasEmptyBillingEventMessage: // if the event message has null UserAddress or empty billing address fields, remove the payment method and billing info
		res, err = s.removeBillingAndPaymentMethodEntity(studentPaymentDetail, billingAddress, studentPaymentActionLogBillingInfo)
		if err != nil {
			return err
		}
	default: // student payment detail and billing address exists, update payment detail and billing address
		// added safety null check though it is impossible to have existing billing address and null student payment detail
		// the null billing address is already checked in the above cases
		if studentPaymentDetail == nil {
			return status.Error(codes.Internal, "unexpected null student payment detail")
		}

		res, err = s.updateBillingAndPayerNameEntity(ctx, db, studentPaymentDetail, billingAddress, info, studentPaymentActionLogBillingInfo)
		if err != nil {
			return err
		}

		err = s.resetThePaymentMethodIfEmpty(ctx, db, studentPaymentDetail, res)
		if err != nil {
			return err
		}
	}

	if res.NeedStudentPaymentDetailUpsert && res.StudentPaymentDetailEntity != nil {
		err = s.StudentPaymentDetailRepo.Upsert(ctx, db, res.StudentPaymentDetailEntity)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("StudentPaymentDetailRepo.Upsert: %v", err))
		}
	}

	if res.NeedBillingAddressUpsert && res.BillingAddressEntity != nil {
		err = s.BillingAddressRepo.Upsert(ctx, db, res.BillingAddressEntity)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("BillingAddressRepo.Upsert: %v", err))
		}
	}

	// Check also if there are changes happened on either payment detail and billing address
	if reflect.DeepEqual(studentPaymentActionLogBillingInfo, genActionLogInfoWithBillingAddress()) || (!res.NeedStudentPaymentDetailUpsert && !res.NeedBillingAddressUpsert) {
		return nil
	}

	// Create action log
	err = s.createPaymentDetailActionLog(ctx, db, studentPaymentDetail.StudentPaymentDetailID.String, invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_DETAILS.String(), res.StudentPaymentActionLogBillingInfo)
	if err != nil {
		return err
	}

	return nil
}

func genRemovedBillingAddressAndPaymentMethod(
	sp *entities.StudentPaymentDetail,
	ba *entities.BillingAddress,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (*entities.StudentPaymentDetail, *entities.BillingAddress, *StudentPaymentActionDetailLogType, error) {
	// Set action log for remove payment method update
	studentPaymentActionLogBillingInfo = setActionLogFromPaymentDetailAndEvent(sp, &BillingAddressInfo{
		PayerName: sp.PayerName.String,
	}, studentPaymentActionLogBillingInfo, "")

	// Set action log for remove billing address update
	studentPaymentActionLogBillingInfo = setActionLogFromBillingAddressAndEvent(ba, "", &BillingAddressInfo{
		PostalCode: "",
		City:       "",
		Street1:    "",
		Street2:    "",
	}, studentPaymentActionLogBillingInfo)

	// Set student default payment method to empty
	err := multierr.Combine(
		sp.PayerName.Set(""),
		sp.PaymentMethod.Set(""),
		ba.City.Set(""),
		ba.PostalCode.Set(""),
		ba.Street1.Set(""),
		ba.Street2.Set(""),
		ba.PrefectureCode.Set(""),
	)
	if err != nil {
		return nil, nil, studentPaymentActionLogBillingInfo, status.Error(codes.Internal, err.Error())
	}

	return sp, ba, studentPaymentActionLogBillingInfo, nil
}

func genUpdatedStudentPayerName(
	payerName string,
	sp *entities.StudentPaymentDetail,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (*entities.StudentPaymentDetail, *StudentPaymentActionDetailLogType, error) {
	studentPaymentActionLogBillingInfo = setActionLogFromPaymentDetailAndEvent(sp, &BillingAddressInfo{
		PayerName: payerName,
	}, studentPaymentActionLogBillingInfo, sp.PaymentMethod.String)

	err := sp.PayerName.Set(payerName)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return sp, studentPaymentActionLogBillingInfo, nil
}

func (s *OpenAPIModifierService) setBillingAddressFieldToUpdate(
	ctx context.Context,
	db database.QueryExecer,
	ba *entities.BillingAddress,
	info *UpdateBillingAddressEventInfo,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (hasChanges bool, actionLogBillingInfo *StudentPaymentActionDetailLogType, err error) {
	errors := []error{}

	existingPrefecture, err := s.PrefectureRepo.FindByPrefectureID(ctx, db, info.UserAddress.Prefecture)
	if err != nil {
		return false, nil, status.Error(codes.Internal, fmt.Sprintf("PrefectureRepo.FindByPrefectureID: %v", err))
	}

	actionLogBillingInfo = setActionLogFromBillingAddressAndEvent(ba, existingPrefecture.PrefectureCode.String, &BillingAddressInfo{
		PostalCode: info.UserAddress.PostalCode,
		City:       info.UserAddress.City,
		Street1:    info.UserAddress.FirstStreet,
		Street2:    info.UserAddress.SecondStreet,
	}, studentPaymentActionLogBillingInfo)

	if ba.City.String != info.UserAddress.City {
		hasChanges = true
		errors = append(errors, ba.City.Set(info.UserAddress.City))
	}

	if ba.PostalCode.String != info.UserAddress.PostalCode {
		hasChanges = true
		errors = append(errors, ba.PostalCode.Set(info.UserAddress.PostalCode))
	}

	if ba.PrefectureCode.String != existingPrefecture.PrefectureCode.String {
		hasChanges = true
		errors = append(errors, ba.PrefectureCode.Set(existingPrefecture.PrefectureCode.String))
	}

	if ba.Street1.String != info.UserAddress.FirstStreet {
		hasChanges = true
		errors = append(errors, ba.Street1.Set(info.UserAddress.FirstStreet))
	}

	if ba.Street2.String != info.UserAddress.SecondStreet {
		hasChanges = true
		errors = append(errors, ba.Street2.Set(info.UserAddress.SecondStreet))
	}

	if err := multierr.Combine(errors...); err != nil {
		return false, nil, status.Error(codes.Internal, err.Error())
	}

	return hasChanges, actionLogBillingInfo, nil
}

func generateNewStudentPaymentDetail(info *UpdateBillingAddressEventInfo) (*entities.StudentPaymentDetail, error) {
	now := time.Now()

	sp := new(entities.StudentPaymentDetail)
	database.AllNullEntity(sp)

	if err := multierr.Combine(
		sp.StudentPaymentDetailID.Set(database.Text(idutil.ULIDNow())),
		sp.StudentID.Set(info.StudentID),
		sp.PayerName.Set(info.PayerName),
		sp.PaymentMethod.Set(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		sp.CreatedAt.Set(now),
		sp.UpdatedAt.Set(now),
	); err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *OpenAPIModifierService) generateNewBillingAddress(ctx context.Context, db database.QueryExecer, info *UpdateBillingAddressEventInfo, studentPaymentDetailID string) (*entities.BillingAddress, error) {
	now := time.Now()

	existingPrefecture, err := s.PrefectureRepo.FindByPrefectureID(ctx, db, info.UserAddress.Prefecture)
	if err != nil {
		return nil, fmt.Errorf("PrefectureRepo.FindByPrefectureID: %v", err)
	}

	ba := new(entities.BillingAddress)
	database.AllNullEntity(ba)

	if err := multierr.Combine(
		ba.BillingAddressID.Set(database.Text(idutil.ULIDNow())),
		ba.StudentPaymentDetailID.Set(studentPaymentDetailID),
		ba.UserID.Set(info.StudentID),
		ba.PostalCode.Set(info.UserAddress.PostalCode),
		ba.PrefectureCode.Set(existingPrefecture.PrefectureCode),
		ba.City.Set(info.UserAddress.City),
		ba.Street1.Set(info.UserAddress.FirstStreet),
		ba.Street2.Set(info.UserAddress.SecondStreet),
		ba.CreatedAt.Set(now),
		ba.UpdatedAt.Set(now),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return ba, nil
}

func (s *OpenAPIModifierService) createNewBillingAndPaymentDetailEntity(
	ctx context.Context,
	db database.QueryExecer,
	info *UpdateBillingAddressEventInfo,
) (*UpdateBillingProcessResult, error) {
	studentPaymentDetailEntity, err := generateNewStudentPaymentDetail(info)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	billingAddressEntity, err := s.generateNewBillingAddress(ctx, db, info, studentPaymentDetailEntity.StudentPaymentDetailID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &UpdateBillingProcessResult{
		StudentPaymentDetailEntity:         studentPaymentDetailEntity,
		BillingAddressEntity:               billingAddressEntity,
		NeedStudentPaymentDetailUpsert:     true,
		NeedBillingAddressUpsert:           true,
		StudentPaymentActionLogBillingInfo: nil,
	}, nil
}

func (s *OpenAPIModifierService) createNewBillingAndUpdatePayerNameEntity(
	ctx context.Context,
	db database.QueryExecer,
	studentPaymentDetail *entities.StudentPaymentDetail,
	info *UpdateBillingAddressEventInfo,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (*UpdateBillingProcessResult, error) {
	res := &UpdateBillingProcessResult{}

	if studentPaymentDetail.PayerName.String != info.PayerName {
		studentPaymentDetailEntity, actionLogInfo, err := genUpdatedStudentPayerName(info.PayerName, studentPaymentDetail, studentPaymentActionLogBillingInfo)
		if err != nil {
			return nil, err
		}

		res.NeedStudentPaymentDetailUpsert = true
		res.StudentPaymentDetailEntity = studentPaymentDetailEntity
		res.StudentPaymentActionLogBillingInfo = actionLogInfo
	}

	billingAddressEntity, err := s.generateNewBillingAddress(ctx, db, info, studentPaymentDetail.StudentPaymentDetailID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res.BillingAddressEntity = billingAddressEntity
	res.NeedBillingAddressUpsert = true

	return res, nil
}

func (s *OpenAPIModifierService) removeBillingAndPaymentMethodEntity(
	studentPaymentDetail *entities.StudentPaymentDetail,
	billingAddress *entities.BillingAddress,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (*UpdateBillingProcessResult, error) {
	studentPaymentDetailEntity, billingAddressEntity, actionLogInfo, err := genRemovedBillingAddressAndPaymentMethod(studentPaymentDetail, billingAddress, studentPaymentActionLogBillingInfo)
	if err != nil {
		return nil, err
	}

	return &UpdateBillingProcessResult{
		StudentPaymentDetailEntity:         studentPaymentDetailEntity,
		BillingAddressEntity:               billingAddressEntity,
		StudentPaymentActionLogBillingInfo: actionLogInfo,
		NeedBillingAddressUpsert:           true,
		NeedStudentPaymentDetailUpsert:     true,
	}, nil
}

func (s *OpenAPIModifierService) updateBillingAndPayerNameEntity(
	ctx context.Context,
	db database.QueryExecer,
	studentPaymentDetail *entities.StudentPaymentDetail,
	billingAddress *entities.BillingAddress,
	info *UpdateBillingAddressEventInfo,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (*UpdateBillingProcessResult, error) {
	res := &UpdateBillingProcessResult{}

	if studentPaymentDetail.PayerName.String != info.PayerName {
		studentPaymentDetailEntity, actionLogInfo, err := genUpdatedStudentPayerName(info.PayerName, studentPaymentDetail, studentPaymentActionLogBillingInfo)
		if err != nil {
			return nil, err
		}

		res.NeedStudentPaymentDetailUpsert = true
		res.StudentPaymentDetailEntity = studentPaymentDetailEntity
		res.StudentPaymentActionLogBillingInfo = actionLogInfo
	}

	var hasFieldChanges bool
	hasFieldChanges, actionLogInfo, err := s.setBillingAddressFieldToUpdate(ctx, db, billingAddress, info, studentPaymentActionLogBillingInfo)
	if err != nil {
		return nil, err
	}

	res.StudentPaymentActionLogBillingInfo = actionLogInfo
	if hasFieldChanges {
		res.NeedBillingAddressUpsert = true
		res.BillingAddressEntity = billingAddress
	}

	return res, nil
}

func (s *OpenAPIModifierService) resetThePaymentMethodIfEmpty(
	ctx context.Context, db database.QueryExecer, studentPaymentDetail *entities.StudentPaymentDetail, res *UpdateBillingProcessResult,
) error {
	if strings.TrimSpace(studentPaymentDetail.PaymentMethod.String) != "" {
		return nil
	}

	bankAccount, err := s.BankAccountRepo.FindByStudentID(ctx, db, studentPaymentDetail.StudentID.String)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	paymentMethod := invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
	if bankAccount != nil && bankAccount.IsVerified.Bool {
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
	}

	// Set the payment method action log in UpdateBillingProcessResult
	res.StudentPaymentActionLogBillingInfo.Previous.PaymentMethod = ""
	res.StudentPaymentActionLogBillingInfo.New.PaymentMethod = paymentMethod

	// Check first if the student payment detail in UpdateBillingProcessResult is not nil
	// It is possible the StudentPaymentDetailEntity become nil if there are no changes in student's payer name
	if res.StudentPaymentDetailEntity == nil {
		res.StudentPaymentDetailEntity = studentPaymentDetail
	}

	// Set the payment method of student payment detail in UpdateBillingProcessResult
	err = res.StudentPaymentDetailEntity.PaymentMethod.Set(paymentMethod)
	if err != nil {
		return err
	}

	return nil
}
