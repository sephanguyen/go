package invoicemgmt

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"golang.org/x/exp/slices"
)

func (s *suite) thesePaymentFileAreSavedAndUploadedSuccessfully(ctx context.Context, count int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.PaymentRequestFileIDs) != count {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number payment file to be %d got %d", count, len(stepState.PaymentRequestFileIDs))
	}

	for _, fileID := range stepState.PaymentRequestFileIDs {
		// Fetch the payment file entity
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
		// check in minio storage if file exist using the object name formatted
		objectName := fmt.Sprintf("%s/%s-%s", invoiceConst.InvoiceFileFolderUploadPath, e.BulkPaymentRequestID.String, e.FileName.String)
		exists, err := s.MinIOClient.IsObjectExists(ctx, objectName)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.MinIOClient.IsObjectExists err: %V", err)
		}

		if !exists {
			return StepStateToContext(ctx, stepState), fmt.Errorf("object %v does not exists in MinIO storage", objectName)
		}

		// further checking of contents are tested in download payment feature
		// get the file content by file url to check the payments included
		resp, err := http.Get(e.FileURL.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		defer resp.Body.Close()

		tempDir, err := os.MkdirTemp("", invoiceConst.InvoiceFileFolderUploadPath)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot create temporary directory: %v", tempDir)
		}

		fileTemp, err := createTempFile(tempDir, e.FileName.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error %v cannot create temporary file: %v on directory: %v", err, e.FileName.String, tempDir)
		}

		_, err = io.Copy(fileTemp, resp.Body)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		defer fileTemp.Close()

		defer cleanup(tempDir)
		// Read the file and get the byte content
		bytesData, err := os.ReadFile(fmt.Sprintf("%s-%s", tempDir, e.FileName.String))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error %v cannot read file: %v on directory: %v", err, e.FileName.String, tempDir)
		}

		defer cleanup(tempDir)

		// check payment method on the request
		req := stepState.Request.(*invoice_pb.CreatePaymentRequestRequest)
		switch req.PaymentMethod {
		case invoice_pb.PaymentMethod_CONVENIENCE_STORE:
			// check the payments included in the file
			ctx, err = s.checkPaymentsIncludedInCSVFile(ctx, bytesData, count)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		case invoice_pb.PaymentMethod_DIRECT_DEBIT:
			// check the payments included in the file
			ctx, err = s.checkPaymentsIncludedInTxtFile(ctx, bytesData, count)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		default:
			return nil, fmt.Errorf("invalid payment method on create payment request %v", req.PaymentMethod.String())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkPaymentsIncludedInTxtFile(ctx context.Context, fileByteContent []byte, fileCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	byteContent := fileByteContent
	if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableEncodePaymentRequestFiles) {
		byteContent, err = utils.DecodeByteToShiftJIS(fileByteContent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	buf := bytes.NewBuffer(byteContent)
	content := buf.String()
	eachLine := strings.Split(content, "\n")

	// total number of line should be 3 + number_of_payments divided into file
	expectedLength := len(stepState.PaymentIDs)/fileCount + 3
	if expectedLength != len(eachLine) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting total number of lines to be %d got %d", expectedLength, len(eachLine))
	}

	paymentRepo := &repositories.PaymentRepo{}

	for i, line := range eachLine {
		if len([]rune(line)) != 120 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting length per line to be %d got %d", 120, len([]rune(line)))
		}

		// not a header record, trailing and end record
		if i != 0 && i != len(eachLine)-2 && i != len(eachLine)-1 {
			lineRune := []rune(line)

			paymentSeqNum, err := strconv.Atoi(strings.TrimSpace(string(lineRune[91:111])))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get the int type of payment seq number err: %v", err)
			}

			payment, err := paymentRepo.FindByPaymentSequenceNumber(ctx, s.InvoiceMgmtPostgresDBTrace, paymentSeqNum)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("paymentRepo.FindByPaymentSequenceNumber err: %v", err)
			}

			if !slices.Contains(stepState.PaymentIDs, payment.PaymentID.String) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error payment sequence number: %v mismatched", payment.PaymentID.String)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkPaymentsIncludedInCSVFile(ctx context.Context, fileByteContent []byte, fileCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// read the CSV lines
	r := csv.NewReader(bytes.NewReader(fileByteContent))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// number of lines should be (number_of_payment * 6) + 1
	// total number of line should be (number_of_payment * 6) divided into file + 1
	expectedLength := (len(stepState.PaymentIDs)*6)/fileCount + 1
	if expectedLength != len(lines) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number of lines in a file to be %d got %d", expectedLength, len(lines))
	}

	paymentRepo := &repositories.PaymentRepo{}

	// chunk the lines per invoice record
	chunkLines := chunkSliceofSliceOfString(lines, 6)
	for _, chunkLine := range chunkLines {
		for j, line := range chunkLine {
			// the second line of the chunk, should be the invoice control record etc.
			if j == 2 {
				paymentSeqNum, err := strconv.Atoi(strings.TrimSpace(line[2]))
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get the int type of payment seq number err: %v", err)
				}

				payment, err := paymentRepo.FindByPaymentSequenceNumber(ctx, s.InvoiceMgmtPostgresDBTrace, paymentSeqNum)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("paymentRepo.FindByPaymentSequenceNumber err: %v", err)
				}

				if !slices.Contains(stepState.PaymentIDs, payment.PaymentID.String) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("error payment sequence number: %v mismatched", payment.PaymentID.String)
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereArePaymentFileAssociatedToAPaymentRequest(ctx context.Context, fileCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// select the payment request file ids last inserted to check with payment ids on the next step
	query := `
		SELECT 
			bulk_payment_request_file_id
		FROM bulk_payment_request_file
		WHERE bulk_payment_request_id = $1 AND resource_path = $2
	`
	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, query, stepState.BulkPaymentRequestID, stepState.ResourcePath)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting payment files err: %v", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Err() err: %v", err)
	}

	fileIDs := make([]string, 0)
	for rows.Next() {
		var fileID string
		if err := rows.Scan(&fileID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("row.Scan() err: %v", err)
		}

		fileIDs = append(fileIDs, fileID)
	}
	// Check if the count from the query is equal to the number of payments
	if len(fileIDs) != fileCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number payment file associated to payment request %d got %d", fileCount, len(fileIDs))
	}

	// Can be use for other test step
	stepState.PaymentRequestFileIDs = fileIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thesePaymentsBelongsToABulkPayment(ctx context.Context, bulkPaymentCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if bulkPaymentCount > 0 {
		err := InsertEntities(
			stepState,
			s.EntitiesCreator.CreateBulkPayment(ctx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING.String(), stepState.PaymentMethod),
		)

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		var paymentIDsBelongToBulk []string
		for i := 0; i < bulkPaymentCount; i++ {
			paymentID := stepState.PaymentIDs[i]
			paymentIDsBelongToBulk = append(paymentIDsBelongToBulk, paymentID)
		}

		stmt := "UPDATE payment SET bulk_payment_id = $1 WHERE payment_id = ANY($2) AND resource_path = $3"

		_, err = s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, stepState.BulkPaymentID, paymentIDsBelongToBulk, s.ResourcePath)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error updating payments with bulk payment id: %v", stepState.BulkPaymentID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsBulkPaymentStatusUpdatedToExported(ctx context.Context, existingStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if existingStatus == "existing" {
		// check bulk payment status that should be exported
		var bulkPaymentStatus string

		stmt := `SELECT bulk_payment_status FROM bulk_payment WHERE bulk_payment_id = $1 AND resource_path = $2`
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.BulkPaymentID, stepState.ResourcePath)
		err := row.Scan(&bulkPaymentStatus)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error selecting bulk payments with bulk payment id: %v", stepState.BulkPaymentID)
		}

		if bulkPaymentStatus != invoice_pb.BulkPaymentStatus_BULK_PAYMENT_EXPORTED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error bulk payment status expected: %v got: %v on bulk payment id: %v", invoice_pb.BulkPaymentStatus_BULK_PAYMENT_EXPORTED.String(), bulkPaymentStatus, stepState.BulkPaymentID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvoicesHaveInvoiceAdjustmentWithAmount(ctx context.Context, amount float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, invoiceID := range stepState.InvoiceIDs {
		err := InsertEntities(
			stepState,
			s.EntitiesCreator.CreateInvoiceAdjustment(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID, stepState.StudentIds[0], amount),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
