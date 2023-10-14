package invoicemgmt

import (
	"context"
	"fmt"

	invoiceCmd "github.com/manabie-com/backend/cmd/server/invoicemgmt"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

func (s *suite) adminIsLoggedInBackOfficeOnOrganization(ctx context.Context, organizationID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.setResourcePathAndClaims(ctx, organizationID)

	// get the location of org and assign to stepState.LocationID
	locationID, err := s.getLocationWithAccessPath(ctx, userConstant.RoleSchoolAdmin, organizationID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.LocationID = locationID
	stepState.OrganizationID = organizationID

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Assign the user_id to Manabie claims
	claims := interceptors.JWTClaimsFromContext(ctx)
	if claims != nil {
		claims.Manabie.UserID = stepState.CurrentUserID
		ctx = interceptors.ContextWithJWTClaims(ctx, claims)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thesePaymentsBelongToOldPaymentRequestFiles(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Sort payments by payment method
	paymentRepo := &repositories.PaymentRepo{}
	invoiceRepo := &repositories.InvoiceRepo{}
	paymentInvoices, err := paymentRepo.FindPaymentInvoiceByIDs(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.PaymentIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	directDebitPayments := []*entities.Payment{}
	convenienceStorePayments := []*entities.Payment{}
	for _, e := range paymentInvoices {
		switch e.Payment.PaymentMethod.String {
		case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
			directDebitPayments = append(directDebitPayments, e.Payment)
		case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
			convenienceStorePayments = append(convenienceStorePayments, e.Payment)
		}
	}

	// For each payment method, create bulk payment request

	// Create DIRECT DEBIT payment request
	if len(directDebitPayments) != 0 {
		err = InsertEntities(
			stepState,
			s.EntitiesCreator.CreateBulkPaymentRequest(ctx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
			s.EntitiesCreator.CreateBulkPaymentRequestFile(ctx, s.InvoiceMgmtPostgresDBTrace, "txt"),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, payment := range directDebitPayments {
			err = InsertEntities(
				stepState,
				s.EntitiesCreator.CreateBulkPaymentRequestFilePayment(ctx, s.InvoiceMgmtPostgresDBTrace, payment.PaymentID.String),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	// Create CONVENIENCE STORE payment request
	if len(convenienceStorePayments) != 0 {
		err = InsertEntities(
			stepState,
			s.EntitiesCreator.CreateBulkPaymentRequest(ctx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			s.EntitiesCreator.CreateBulkPaymentRequestFile(ctx, s.InvoiceMgmtPostgresDBTrace, "csv"),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, payment := range convenienceStorePayments {
			err = InsertEntities(
				stepState,
				s.EntitiesCreator.CreateBulkPaymentRequestFilePayment(ctx, s.InvoiceMgmtPostgresDBTrace, payment.PaymentID.String),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	// Update payment is_exported field
	err = paymentRepo.UpdateIsExportedByPaymentIDs(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.PaymentIDs, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Update invoice is_exported field
	err = invoiceRepo.UpdateIsExportedByInvoiceIDs(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.InvoiceIDs, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anAdminRunsTheUploadPaymentRequestFileJobScript(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Set as panic so the we will not see the logs of the script while running tests
	zLogger := logger.NewZapLogger("panic", true)
	err := invoiceCmd.UploadExistingPaymentRequestFile(
		ctx,
		&s.Cfg.Storage,
		s.InvoiceMgmtPostgresDBTrace,
		zLogger.Sugar(),
		[]string{stepState.OrganizationID},
	)
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theFileURLofPaymentFileIsNotEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, fileID := range stepState.PaymentRequestFileIDs {
		query := `
			SELECT
				bulk_payment_request_id,
				file_name,
				file_url
			FROM bulk_payment_request_file
			WHERE bulk_payment_request_file_id = $1
			AND resource_path = $2
		`
		e := &entities.BulkPaymentRequestFile{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, fileID, stepState.ResourcePath).Scan(&e.BulkPaymentRequestID, &e.FileName, &e.FileURL)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting payment file err: %v", err)
		}

		if e.FileURL.String == "" || e.FileURL.Status != pgtype.Present {
			return StepStateToContext(ctx, stepState), fmt.Errorf("the payment file with ID %v has no file_url", fileID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thesePaymentFilesAreUploadedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.thesePaymentFileAreSavedAndUploadedSuccessfully(ctx, len(stepState.PaymentRequestFileIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thePaymentRequestFilesHaveACorrectFormat(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, fileID := range stepState.PaymentRequestFileIDs {
		req := &invoice_pb.DownloadPaymentFileRequest{
			PaymentRequestFileId: fileID,
		}

		resp, err := invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).DownloadPaymentFile(contextWithToken(ctx), req)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response = resp

		query := `
			SELECT
				pr.payment_method
			FROM bulk_payment_request_file prf
			INNER JOIN bulk_payment_request pr
				ON pr.bulk_payment_request_id = prf.bulk_payment_request_id
			WHERE prf.bulk_payment_request_file_id = $1
			AND resource_path = $2
		`
		var paymentMethod string
		err = s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, fileID, stepState.ResourcePath).Scan(&paymentMethod)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting the file payment method err: %v", err)
		}

		switch paymentMethod {
		case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
			ctx, err = s.thePaymentRequestFileHasACorrectCSVFormat(ctx)
		case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
			ctx, err = s.thePaymentRequestFileHasACorrectBankTXTFormat(ctx)
		}
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getLocationWithAccessPath(ctx context.Context, roleName, organizationID string) (string, error) {
	locationWithAccessPathQuery := `
		SELECT grap.location_id
		FROM  granted_role gr
		INNER JOIN role r
			ON gr.role_id = r.role_id
		INNER JOIN granted_role_access_path grap
			ON gr.granted_role_id = grap.granted_role_id
		WHERE r.role_name = $1 AND r.resource_path = $2
		LIMIT 1
	`

	defaultOrgLocationQuery := `
		SELECT location_id FROM locations 
		WHERE resource_path = $1
		ORDER BY created_at ASC
		LIMIT 1
	`

	// Get the location with access path set to a role name
	var locationID string
	err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, locationWithAccessPathQuery, roleName, organizationID).Scan(&locationID)
	if err == nil {
		return locationID, nil
	}

	// If there is no current location that has access path, use the default location of an org
	var defaultLocationID string
	err = s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, defaultOrgLocationQuery, organizationID).Scan(&defaultLocationID)
	if err != nil {
		return "", fmt.Errorf("error fetching default location err: %v", err)
	}

	return defaultLocationID, nil
}

func (s *suite) thereIsPaymentRequestFileThatHasNoAssociatedPayment(ctx context.Context, paymentMethodStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		paymentMethod string
		fileType      string
	)
	switch paymentMethodStr {
	case "DIRECT DEBIT":
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
		fileType = "txt"
	case "CONVENIENCE STORE":
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
		fileType = "csv"
	}

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBulkPaymentRequest(ctx, s.InvoiceMgmtPostgresDBTrace, paymentMethod),
		s.EntitiesCreator.CreateBulkPaymentRequestFile(ctx, s.InvoiceMgmtPostgresDBTrace, fileType),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theUploadPaymentRequestFileScriptHasNoError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error occurred on upload payment file script err: %v", stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theUploadPaymentRequestFileScriptReturnsError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), errors.New("expecting an error but received nil")
	}

	return StepStateToContext(ctx, stepState), nil
}
