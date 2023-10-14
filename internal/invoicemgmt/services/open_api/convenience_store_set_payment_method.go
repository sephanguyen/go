package openapisvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	utils "github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	invoice_common "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/common"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func genActionLogInfoWithBillingAddress() *StudentPaymentActionDetailLogType {
	return &StudentPaymentActionDetailLogType{
		Previous: &PreviousDataStudentActionDetailLog{
			BillingAddress: &BillingAddressHistoryInfo{},
		},
		New: &NewDataStudentActionDetailLog{
			BillingAddress: &BillingAddressHistoryInfo{},
		},
	}
}

func (s *OpenAPIModifierService) AutoSetConvenienceStore(ctx context.Context, billingAddressEventInfo *BillingAddressInfo) error {
	// validate billing address event info
	err := validateBillingAddressEventInfo(billingAddressEventInfo)
	if err != nil {
		s.logger.Warnf("validation error for student %s err: %v", billingAddressEventInfo.StudentID, err)
		return nil
	}

	// check prefecture if exists
	existingPrefecture, err := s.PrefectureRepo.FindByPrefectureID(ctx, s.DB, billingAddressEventInfo.PrefectureID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("prefectureRepo.FindByPrefectureID: %v", err))
	}

	studentPaymentActionLogBillingInfo := genActionLogInfoWithBillingAddress()

	studentPaymentDetail, studentPaymentActionLogBillingInfo, err := s.getStudentPaymentDetailToUpsert(ctx, billingAddressEventInfo, studentPaymentActionLogBillingInfo)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	billingAddress, studentPaymentActionLogBillingInfo, err := s.getBillingAddressToUpsert(ctx, billingAddressEventInfo, studentPaymentDetail.StudentPaymentDetailID.String, existingPrefecture.PrefectureCode.String, studentPaymentActionLogBillingInfo)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := try.Do(func(attempt int) (bool, error) {
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
			// Upsert student payment detail and billing address info
			if err := s.StudentPaymentDetailRepo.Upsert(ctx, tx, studentPaymentDetail); err != nil {
				ok, postgresErr := checkPostgresSQLErrorCode(err)
				if ok && postgresErr != nil {
					return postgresErr
				}

				return status.Error(codes.Internal, fmt.Sprintf("student payment detail repo upsert err: %v", err.Error()))
			}
			if err := s.BillingAddressRepo.Upsert(ctx, tx, billingAddress); err != nil {
				ok, postgresErr := checkPostgresSQLErrorCode(err)
				if ok && postgresErr != nil {
					return postgresErr
				}

				return status.Error(codes.Internal, fmt.Sprintf("billing address repo upsert err: %v", err.Error()))
			}

			// Check if there is no changes with studentPaymentActionLogBillingInfo
			if reflect.DeepEqual(studentPaymentActionLogBillingInfo, genActionLogInfoWithBillingAddress()) {
				return nil
			}

			// Create action log
			err = s.createPaymentDetailActionLog(ctx, tx, studentPaymentDetail.StudentPaymentDetailID.String, invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_DETAILS.String(), studentPaymentActionLogBillingInfo)
			if err != nil {
				return err
			}

			return nil
		})
		// If no error, return and not retry
		if err == nil {
			return false, nil
		}

		if err.Error() != constant.PgConnForeignKeyError && err.Error() != constant.TableRLSError {
			return false, err
		}

		// retry if the error is foreign key violation error or RLS error.
		time.Sleep(constant.NatsEventRetryQuerySleep)
		log.Printf("Retrying the saving student payment detail and billing address. Attempt: %d \n", attempt)
		return attempt < 10, fmt.Errorf("cannot create student payment detail and billing address, err %v", err)
	}); err != nil {
		return err
	}

	return nil
}

func (s *OpenAPIModifierService) getStudentPaymentDetailToUpsert(
	ctx context.Context,
	billingAddressEventInfo *BillingAddressInfo,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (*entities.StudentPaymentDetail, *StudentPaymentActionDetailLogType, error) {
	studentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, s.DB, billingAddressEventInfo.StudentID)
	now := time.Now()

	switch err {
	case nil:
		// Set action log details first before modifying the student payment detail entity
		studentPaymentActionLogBillingInfo = setActionLogFromPaymentDetailAndEvent(studentPaymentDetail, billingAddressEventInfo, studentPaymentActionLogBillingInfo, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())

		// update the existing student payment detail fields
		if err := multierr.Combine(
			studentPaymentDetail.PayerName.Set(billingAddressEventInfo.PayerName),
			studentPaymentDetail.PaymentMethod.Set(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			studentPaymentDetail.UpdatedAt.Set(now),
		); err != nil {
			return nil, nil, fmt.Errorf("multierr.Combine: %w", err)
		}
	case pgx.ErrNoRows:
		// create new student payment detail entity
		studentPaymentDetail = new(entities.StudentPaymentDetail)
		database.AllNullEntity(studentPaymentDetail)

		if err := multierr.Combine(
			studentPaymentDetail.StudentPaymentDetailID.Set(database.Text(idutil.ULIDNow())),
			studentPaymentDetail.StudentID.Set(billingAddressEventInfo.StudentID),
			studentPaymentDetail.PayerName.Set(billingAddressEventInfo.PayerName),
			studentPaymentDetail.PaymentMethod.Set(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			studentPaymentDetail.CreatedAt.Set(now),
			studentPaymentDetail.UpdatedAt.Set(now),
		); err != nil {
			return nil, nil, fmt.Errorf("multierr.Combine: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("studentPaymentDetailRepo.FindByStudentID: %v", err)
	}

	return studentPaymentDetail, studentPaymentActionLogBillingInfo, nil
}

func (s *OpenAPIModifierService) getBillingAddressToUpsert(
	ctx context.Context,
	billingAddressEventInfo *BillingAddressInfo,
	studentPaymentDetailID string,
	prefectureCode string,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) (*entities.BillingAddress, *StudentPaymentActionDetailLogType, error) {
	billingAddress, err := s.BillingAddressRepo.FindByUserID(ctx, s.DB, billingAddressEventInfo.StudentID)
	now := time.Now()

	switch err {
	case nil:
		if billingAddress.StudentPaymentDetailID.String != studentPaymentDetailID {
			return nil, nil, fmt.Errorf("existing billing address student payment detail got: %v expected: %v", billingAddress.StudentPaymentDetailID.String, studentPaymentDetailID)
		}

		// Set action log details first before modifying the billing address entity
		studentPaymentActionLogBillingInfo = setActionLogFromBillingAddressAndEvent(billingAddress, prefectureCode, billingAddressEventInfo, studentPaymentActionLogBillingInfo)

		// update the existing billing address fields
		if err := multierr.Combine(
			billingAddress.PostalCode.Set(billingAddressEventInfo.PostalCode),
			billingAddress.PrefectureCode.Set(prefectureCode),
			billingAddress.City.Set(billingAddressEventInfo.City),
			billingAddress.Street1.Set(billingAddressEventInfo.Street1),
			billingAddress.Street2.Set(billingAddressEventInfo.Street2),
			billingAddress.UpdatedAt.Set(now),
		); err != nil {
			return nil, nil, fmt.Errorf("multierr.Combine: %w", err)
		}
	case pgx.ErrNoRows:
		// create new billing address entity
		billingAddress = new(entities.BillingAddress)
		database.AllNullEntity(billingAddress)

		if err := multierr.Combine(
			billingAddress.BillingAddressID.Set(database.Text(idutil.ULIDNow())),
			billingAddress.StudentPaymentDetailID.Set(studentPaymentDetailID),
			billingAddress.UserID.Set(billingAddressEventInfo.StudentID),
			billingAddress.PostalCode.Set(billingAddressEventInfo.PostalCode),
			billingAddress.PrefectureCode.Set(prefectureCode),
			billingAddress.City.Set(billingAddressEventInfo.City),
			billingAddress.Street1.Set(billingAddressEventInfo.Street1),
			billingAddress.Street2.Set(billingAddressEventInfo.Street2),
			billingAddress.CreatedAt.Set(now),
			billingAddress.UpdatedAt.Set(now),
		); err != nil {
			return nil, nil, fmt.Errorf("multierr.Combine: %w", err)
		}

	default:
		return nil, nil, fmt.Errorf("billingAddressRepo.FindByUserID: %v", err)
	}

	return billingAddress, studentPaymentActionLogBillingInfo, nil
}

func (s *OpenAPIModifierService) CheckAutoSetConvenienceStoreIsEnabled(ctx context.Context) (bool, error) {
	configValue, err := s.getInternalConfigForAutoSetConvenienceStore(ctx)
	if err != nil {
		return false, err
	}

	// check for feature flag of auto set convenience store
	isAutoSetCSFeature, err := s.UnleashClient.IsFeatureEnabled(constant.EnableAutoSetCCFeatureFlag, s.Env)
	if err != nil {
		return false, fmt.Errorf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableAutoSetCCFeatureFlag, err)
	}

	if configValue == "on" && isAutoSetCSFeature {
		return true, nil
	}

	return false, nil
}

func (s *OpenAPIModifierService) getInternalConfigForAutoSetConvenienceStore(ctx context.Context) (string, error) {
	var configValue string
	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return configValue, fmt.Errorf("s.MasterConfigurationService.GetConfigurations err: %s", err)
	}
	getConfigurationsRequest := &mpb.GetConfigurationsRequest{
		Paging:         &cpb.Paging{},
		Keyword:        AutoSetConvenienceStoreConfigKey,
		OrganizationId: resourcePath,
	}

	res, err := s.MasterConfigurationService.GetConfigurations(utils.SignCtx(ctx), getConfigurationsRequest)

	if err != nil {
		return configValue, fmt.Errorf("s.MasterConfigurationService.GetConfigurations err: %s", err)
	}

	if len(res.GetItems()) == 0 {
		return configValue, fmt.Errorf("s.MasterConfigurationService.GetConfigurations err: no %v config key found", AutoSetConvenienceStoreConfigKey)
	}

	configurationResponse := res.GetItems()[0]
	configValue = configurationResponse.GetConfigValue()

	return configValue, nil
}

func validateBillingAddressEventInfo(billingAddressEventInfo *BillingAddressInfo) error {
	switch {
	case strings.TrimSpace(billingAddressEventInfo.StudentID) == "":
		return errors.New("student id cannot be empty")
	case strings.TrimSpace(billingAddressEventInfo.PayerName) == "":
		return errors.New("payer name cannot be empty")
	case strings.TrimSpace(billingAddressEventInfo.PostalCode) == "":
		return errors.New("postal code cannot be empty")
	case strings.TrimSpace(billingAddressEventInfo.PrefectureID) == "":
		return errors.New("prefecture id cannot be empty")
	case strings.TrimSpace(billingAddressEventInfo.City) == "":
		return errors.New("city cannot be empty")
	}

	return nil
}

func (s *OpenAPIModifierService) createPaymentDetailActionLog(
	ctx context.Context,
	db database.QueryExecer,
	studentPaymentDetailID string,
	action string,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
) error {
	studentPaymentDetailActionLog, err := generateStudentPaymentDetailActionLog(ctx, &studentPaymentDetailActionLogData{
		actionDetailInfo:       database.JSONB(studentPaymentActionLogBillingInfo),
		action:                 action,
		StudentPaymentDetailID: studentPaymentDetailID,
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := s.StudentPaymentDetailActionLogRepo.Create(ctx, db, studentPaymentDetailActionLog); err != nil {
		ok, postgresErr := checkPostgresSQLErrorCode(err)
		if ok && postgresErr != nil {
			return postgresErr
		}

		return status.Error(codes.Internal, fmt.Sprintf("billing address repo upsert err: %v", err.Error()))
	}

	return nil
}

func setActionLogFromPaymentDetailAndEvent(
	studentPaymentDetail *entities.StudentPaymentDetail,
	billingAddressEventInfo *BillingAddressInfo,
	studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType,
	paymentMethod string,
) *StudentPaymentActionDetailLogType {
	if studentPaymentDetail.PayerName.String != billingAddressEventInfo.PayerName {
		studentPaymentActionLogBillingInfo.Previous.BillingAddress.PayerName = studentPaymentDetail.PayerName.String
		studentPaymentActionLogBillingInfo.New.BillingAddress.PayerName = billingAddressEventInfo.PayerName
	}

	if studentPaymentDetail.PaymentMethod.String != paymentMethod {
		studentPaymentActionLogBillingInfo.Previous.PaymentMethod = studentPaymentDetail.PaymentMethod.String
		studentPaymentActionLogBillingInfo.New.PaymentMethod = paymentMethod
	}

	return studentPaymentActionLogBillingInfo
}

func setActionLogFromBillingAddressAndEvent(billingAddress *entities.BillingAddress, prefectureCode string, billingAddressEventInfo *BillingAddressInfo, studentPaymentActionLogBillingInfo *StudentPaymentActionDetailLogType) *StudentPaymentActionDetailLogType {
	if billingAddress.PostalCode.String != billingAddressEventInfo.PostalCode {
		studentPaymentActionLogBillingInfo.Previous.BillingAddress.PostalCode = billingAddress.PostalCode.String
		studentPaymentActionLogBillingInfo.New.BillingAddress.PostalCode = billingAddressEventInfo.PostalCode
	}

	if billingAddress.City.String != billingAddressEventInfo.City {
		studentPaymentActionLogBillingInfo.Previous.BillingAddress.City = billingAddress.City.String
		studentPaymentActionLogBillingInfo.New.BillingAddress.City = billingAddressEventInfo.City
	}

	if billingAddress.Street1.String != billingAddressEventInfo.Street1 {
		studentPaymentActionLogBillingInfo.Previous.BillingAddress.Street1 = billingAddress.Street1.String
		studentPaymentActionLogBillingInfo.New.BillingAddress.Street1 = billingAddressEventInfo.Street1
	}

	if billingAddress.Street2.String != billingAddressEventInfo.Street2 {
		studentPaymentActionLogBillingInfo.Previous.BillingAddress.Street2 = billingAddress.Street2.String
		studentPaymentActionLogBillingInfo.New.BillingAddress.Street2 = billingAddressEventInfo.Street2
	}

	if billingAddress.PrefectureCode.String != prefectureCode {
		studentPaymentActionLogBillingInfo.Previous.BillingAddress.PrefectureCode = billingAddress.PrefectureCode.String
		studentPaymentActionLogBillingInfo.New.BillingAddress.PrefectureCode = prefectureCode
	}

	return studentPaymentActionLogBillingInfo
}
