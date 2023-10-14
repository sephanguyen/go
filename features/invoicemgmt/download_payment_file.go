package invoicemgmt

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"golang.org/x/text/transform"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) thesePaymentsAlreadyBelongToPaymentRequestFileWithPaymentMethod(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &invoice_pb.CreatePaymentRequestRequest{
		PaymentIds: stepState.PaymentIDs,
	}

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.adminChoosesAsPaymentMethod(ctx, paymentMethod)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch paymentMethod {
	case "CONVENIENCE STORE":
		ctx, err = s.adminAddsPaymentDueDateFromAndPaymentDueDateUntil(ctx, -4, -3)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "DIRECT DEBIT":
		ctx, err = s.adminAddsPaymentDueDate(ctx, -7)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	paymentRequest := stepState.Request.(*invoice_pb.CreatePaymentRequestRequest)
	// Create the payment request
	resp, err := invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).CreatePaymentRequest(contextWithToken(ctx), paymentRequest)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Error on creating payment request err: %v", err)
	}
	stepState.BulkPaymentRequestID = resp.BulkPaymentRequestId

	// Assign the payment file ID
	ctx, err = s.thereArePaymentFileAssociatedToAPaymentRequest(ctx, 1) // one file always for downloading
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.PaymentRequestFileIDs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("No payment file generated")
	}

	// re-assign to nil to prevent conflict
	stepState.Request = nil

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminIsAtCreatePaymentRequestTable(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &invoice_pb.DownloadPaymentFileRequest{}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminSelectAndDownloadsThePaymentRequestFile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*invoice_pb.DownloadPaymentFileRequest)
	req.PaymentRequestFileId = stepState.PaymentRequestFileIDs[0]
	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).DownloadPaymentFile(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theDataByteReturnedIsNotEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.DownloadPaymentFileResponse)
	if response.Data == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("The data byte returned is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getPaymentInvoiceBySequenceNumber(ctx context.Context, seqNumber int) (*entities.PaymentInvoiceMap, error) {
	stmt := `
		SELECT
			p.payment_id,
			p.payment_due_date,
			p.payment_expiry_date,
			p.payment_sequence_number,
			p.student_id,
			i.invoice_id,
			i.total
		FROM payment p
		INNER JOIN invoice i
			ON p.invoice_id = i.invoice_id
		WHERE p.payment_sequence_number = $1 AND p.resource_path = $2
	`
	resourcePath := golibs.ResourcePathFromCtx(ctx)

	e := &entities.PaymentInvoiceMap{}
	e.Payment = &entities.Payment{}
	e.Invoice = &entities.Invoice{}

	if err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, stmt, seqNumber, resourcePath).ScanFields(
		&e.Payment.PaymentID,
		&e.Payment.PaymentDueDate,
		&e.Payment.PaymentExpiryDate,
		&e.Payment.PaymentSequenceNumber,
		&e.Payment.StudentID,
		&e.Invoice.InvoiceID,
		&e.Invoice.Total,
	); err != nil {
		return nil, err
	}

	return e, nil
}

func (s *suite) thePaymentRequestFileHasACorrectCSVFormat(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.DownloadPaymentFileResponse)

	// read the CSV lines
	var reader io.Reader
	reader = bytes.NewReader(response.Data)
	if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableEncodePaymentRequestFiles) {
		reader = transform.NewReader(bytes.NewReader(response.Data), utils.GetShiftJISDecoder())
	}

	r := csv.NewReader(reader)
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}

	// number of lines should be (number_of_payment * 6) + 1
	expectedLength := (len(stepState.PaymentIDs) * 6) + 1
	if expectedLength != len(lines) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting number of lines in a file to be %d got %d", expectedLength, len(lines))
	}

	// Testing this convenience store may be flaky because we are fetching the convenience store not by ID.
	// If in the middle of the test there is a partner convenience store that was created, this test may fail.
	// So let's make sure that the partner convenience store is a fixed value
	partnerCSRepo := &repositories.PartnerConvenienceStoreRepo{}
	partnerCS, err := partnerCSRepo.FindOne(ctx, s.InvoiceMgmtPostgresDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("partnerCSRepo.FindOne err: %v", err)
	}

	prefectureRepo := &repositories.PrefectureRepo{}
	prefectures, err := prefectureRepo.FindAll(ctx, s.InvoiceMgmtPostgresDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("PrefectureRepo.FindAll err: %v", err)
	}
	prefectureMap := make(map[string]string)
	for _, e := range prefectures {
		_, ok := prefectureMap[e.PrefectureCode.String]
		if ok {
			continue
		}
		prefectureMap[e.PrefectureCode.String] = e.Name.String
	}

	paymentInvoices := []*entities.PaymentInvoiceMap{}

	studentPaymentDetailRepo := &repositories.StudentPaymentDetailRepo{}
	billItemRepo := &repositories.BillItemRepo{}
	invoiceAdjustmentRepo := &repositories.InvoiceAdjustmentRepo{}

	enablePaymentRequestFileFormat := isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableFormatPaymentRequestFileFields)

	// chunk the lines per invoice record
	// validate the row per line number
	// if the line is the first line of the chunk, it should be the header
	// if the line is the second line of the chunk, it should be the invoice control record etc.
	// the data will look like [ [[1,1], [1,3], [3,1], [3,7,1] [3,7,2"], [3,7,3]] [[1,1], [1,3], [3,1], [3,7,1] [3,7,2"], [3,7,3]] [[9]] ]
	chunkLines := chunkSliceofSliceOfString(lines, 6)
	for i, chunkLine := range chunkLines {
		// validate end record
		if i == len(chunkLines)-1 {
			err := validateCSVEndRecord(paymentInvoices, chunkLine[0])
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validateCSVEndRecord err: %v", err)
			}
			continue
		}

		var (
			paymentInvoice *entities.PaymentInvoiceMap
			dataMapList    *FilePaymentDataMap
		)

		for j, line := range chunkLine {
			switch j {
			case 0:
				err = validateCSVHeaderRecord(partnerCS, line)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("validateCSVHeaderRecord err: %v", err)
				}
			case 1:
				if enablePaymentRequestFileFormat {
					err = validateCSVInvoiceControlRecordV2(partnerCS, line)
				} else {
					err = validateCSVInvoiceControlRecord(partnerCS, line)
				}
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("validateCSVInvoiceControlRecord err: %v", err)
				}
			case 2:
				// get the payment invoice
				paymentSeqNumber, err := strconv.Atoi(line[2])
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("Cannot get the int type of payment seq number err: %v", err)
				}

				paymentInvoice, err = s.getPaymentInvoiceBySequenceNumber(ctx, paymentSeqNumber)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				studentBillingDetails, err := studentPaymentDetailRepo.FindStudentBillingByStudentIDs(ctx, s.InvoiceMgmtPostgresDBTrace, []string{paymentInvoice.Payment.StudentID.String})
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("student billing details on payment student id: %v has error: %v", paymentInvoice.Payment.StudentID.String, err)
				}

				if len(studentBillingDetails) != 1 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected billing detail records count: %d on student id: %v", len(studentBillingDetails), paymentInvoice.Payment.StudentID.String)
				}

				dataMapList = &FilePaymentDataMap{
					Payment:            paymentInvoice.Payment,
					Invoice:            paymentInvoice.Invoice,
					StudentBillingInfo: studentBillingDetails[0],
				}

				paymentInvoices = append(paymentInvoices, paymentInvoice)

				if enablePaymentRequestFileFormat {
					err = validateCSVInvoiceRecordV2(dataMapList, line, prefectureMap, s.UnleashSuite)
				} else {
					err = validateCSVInvoiceRecord(dataMapList, line, prefectureMap, s.UnleashSuite)
				}
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("validateCSVInvoiceRecord err: %v", err)
				}
			case 3, 4, 5:
				if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableBillingMessageInCSVMessages) {
					invoiceBillItemMap, err := billItemRepo.FindInvoiceBillItemMapByInvoiceIDs(ctx, s.InvoiceMgmtPostgresDBTrace, []string{paymentInvoice.Invoice.InvoiceID.String})
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}

					invoiceAdjustment, err := invoiceAdjustmentRepo.FindByInvoiceIDs(ctx, s.InvoiceMgmtPostgresDBTrace, []string{paymentInvoice.Invoice.InvoiceID.String})
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}

					dataMapList.BillItemDetails = invoiceBillItemMap
					dataMapList.InvoiceAdjustments = invoiceAdjustment

					if enablePaymentRequestFileFormat {
						err = validateCSVMessageRecordWithBillingMessageV2(partnerCS, dataMapList, line, j-2)
					} else {
						err = validateCSVMessageRecordWithBillingMessage(partnerCS, dataMapList, line, j-2)
					}
					if err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("validateCSVMessageRecordWithBillingMessage err: %v", err)
					}
				} else {
					err = validateCSVMessageRecord(partnerCS, line, j-2) // the j - 2 will be the message code of message. The message codes are (1, 2, 3)
					if err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("validateCSVMessageRecord err: %v", err)
					}
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) partnerHasExistingPartnerBank(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	partnerBankRepo := &repositories.PartnerBankRepo{}
	_, err := partnerBankRepo.FindOne(ctx, s.InvoiceMgmtPostgresDBTrace)
	if err == nil {
		return StepStateToContext(ctx, stepState), nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("partnerBankRepo.FindOne err: %v", err)
	}

	// Only create partner cbank if there is no existing for partner
	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreatePartnerBank(ctx, s.InvoiceMgmtPostgresDBTrace, true),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thePaymentRequestFileHasACorrectBankTXTFormat(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*invoice_pb.DownloadPaymentFileResponse)

	var err error
	byteContent := response.Data
	if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableEncodePaymentRequestFiles) {
		byteContent, err = utils.DecodeByteToShiftJIS(response.Data)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	buf := bytes.NewBuffer(byteContent)
	content := buf.String()

	eachLine := strings.Split(content, "\n")

	// The total number of line should be 3 + number_of_payments
	expectedLength := len(stepState.PaymentIDs) + 3
	if expectedLength != len(eachLine) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting total number of lines to be %d got %d", expectedLength, len(eachLine))
	}

	paymentInvoices := []*entities.PaymentInvoiceMap{}
	studentPaymentDetailRepo := &repositories.StudentPaymentDetailRepo{}
	bankBranchRepo := &repositories.BankBranchRepo{}

	// all records are belong to a partner bank in a file
	// save the header line to validate when partner bank data is available on data record line
	var headerLine string

	for i, line := range eachLine {
		if len([]rune(line)) != 120 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting length per line to be %d got %d", 120, len([]rune(line)))
		}

		switch i {
		case 0:
			headerLine = line
		case len(eachLine) - 2:
			err := validateBankTxtTrailerRecord(paymentInvoices, line)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validateBankTxtTrailerRecord err: %v", err)
			}
		case len(eachLine) - 1:
			err := validateBankTxtEndRecord(line)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validateBankTxtEndRecord err: %v", err)
			}
		default:
			lineRune := []rune(line)
			paymentSeqNumber, err := strconv.Atoi(strings.TrimSpace(string(lineRune[91:111])))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("Cannot get the int type of payment seq number err: %v", err)
			}

			paymentInvoice, err := s.getPaymentInvoiceBySequenceNumber(ctx, paymentSeqNumber)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.getPaymentInvoiceBySequenceNumber err: %v", err)
			}

			// get bank account details of student
			studentBankAccountDetails, err := studentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs(ctx, s.InvoiceMgmtPostgresDBTrace, []string{paymentInvoice.Payment.StudentID.String})
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("student bank details on payment student id: %v has error: %v", paymentInvoice.Payment.StudentID.String, err)
			}
			// setting error initially for one record to be check on future use if multiple records
			if len(studentBankAccountDetails) != 1 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected bank account detail records count: %d on student id: %v", len(studentBankAccountDetails), paymentInvoice.Payment.StudentID.String)
			}

			// get related bank of student
			relatedBankOfBankBranch, err := bankBranchRepo.FindRelatedBankOfBankBranches(ctx, s.InvoiceMgmtPostgresDBTrace, []string{studentBankAccountDetails[0].BankAccount.BankBranchID.String})
			if err != nil {
				return StepStateToContext(ctx, stepState), status.Error(codes.Internal, fmt.Sprintf("bank branch find related bank of student id: %v has error: %v", paymentInvoice.Payment.StudentID.String, err))
			}
			if len(relatedBankOfBankBranch) != 1 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected bank account detail records count: %d on student id: %v", len(studentBankAccountDetails), paymentInvoice.Payment.StudentID.String)
			}

			// check new customer code history if existing
			newCustomerCodeHistory := &entities.NewCustomerCodeHistory{}
			fields, _ := newCustomerCodeHistory.FieldMap()
			query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_account_number = $1 AND student_id = $2", strings.Join(fields, ","), newCustomerCodeHistory.TableName())
			err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, studentBankAccountDetails[0].BankAccount.BankAccountNumber.String, paymentInvoice.Payment.StudentID.String).ScanOne(newCustomerCodeHistory)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("Error on selecting new customer code history: %w", err)
			}

			dataMapList := &FilePaymentDataMap{
				Payment:                paymentInvoice.Payment,
				Invoice:                paymentInvoice.Invoice,
				StudentBankDetails:     studentBankAccountDetails[0],
				StudentRelatedBank:     relatedBankOfBankBranch[0],
				NewCustomerCodeHistory: newCustomerCodeHistory,
			}
			// used for checking and amending on trailer record
			paymentInvoices = append(paymentInvoices, paymentInvoice)

			err = validateBankTxtHeaderRecord(relatedBankOfBankBranch[0].PartnerBank, headerLine)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validateBankTxtHeaderRecord err: %v", err)
			}

			err = validateBankTxtDataRecord(dataMapList, line)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("validateBankTxtDataRecord err: %v", err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) sendADownloadFileRequestWithEmptyFileID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*invoice_pb.DownloadPaymentFileRequest)
	req.PaymentRequestFileId = ""
	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).DownloadPaymentFile(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAPaymentFileThatHasNoAssociatedPayments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBulkPaymentRequest(ctx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		s.EntitiesCreator.CreateBulkPaymentRequestFile(ctx, s.InvoiceMgmtPostgresDBTrace, "csv"),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.PaymentRequestFileIDs = append(stepState.PaymentRequestFileIDs, stepState.BulkPaymentRequestFileID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnExistingPaymentFileInCloudStorage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Get the file name by payment ID
	query := `
		SELECT file_name, bulk_payment_request_id
		FROM bulk_payment_request_file
		WHERE bulk_payment_request_file_id = $1
	`
	var (
		fileName             string
		bulkPaymentRequestID string
	)
	if err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, stepState.PaymentRequestFileIDs[0]).ScanFields(
		&fileName,
		&bulkPaymentRequestID,
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	objectName := fmt.Sprintf("%s/%s-%s", invoiceConst.InvoiceFileFolderUploadPath, bulkPaymentRequestID, fileName)
	exists, err := s.MinIOClient.IsObjectExists(ctx, objectName)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.MinIOClient.IsObjectExists err: %V", err)
	}

	if !exists {
		return StepStateToContext(ctx, stepState), fmt.Errorf("object %v does not exists in MinIO storage", objectName)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAPaymentRequestFileWithPaymentMethodThatIsNotInCloudStorage(ctx context.Context, paymentMethodStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		fileType      string
		paymentMethod string
	)

	switch paymentMethodStr {
	case "CONVENIENCE STORE":
		fileType = "csv"
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
	case "DIRECT DEBIT":
		fileType = "txt"
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
	}

	// Create payment request file without associating payments and the file used is not uploaded to cloud storage
	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBulkPaymentRequest(ctx, s.InvoiceMgmtPostgresDBTrace, paymentMethod),
		s.EntitiesCreator.CreateBulkPaymentRequestFile(ctx, s.InvoiceMgmtPostgresDBTrace, fileType),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.PaymentRequestFileIDs = append(stepState.PaymentRequestFileIDs, stepState.BulkPaymentRequestFileID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billingItemsOfStudentsAreAdjustmentBillingType(ctx context.Context, identification string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := "UPDATE bill_item SET adjustment_price = final_price, bill_type = $1 WHERE bill_item_sequence_number = $2 AND resource_path = $3"
	for _, orderID := range stepState.OrderIDs {
		billItemRepo := &repositories.BillItemRepo{}
		billItems, err := billItemRepo.FindByOrderID(ctx, s.InvoiceMgmtPostgresDBTrace, orderID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(billItems) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("no bill item found on order with ID %s", orderID)
		}

		if identification == "one" {
			_, err := s.FatimaDBTrace.Exec(ctx, stmt, payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String(), billItems[0].BillItemSequenceNumber.Int, s.ResourcePath)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bill item: %v", err)
			}

			continue
		}

		for _, billItem := range billItems {
			_, err := s.FatimaDBTrace.Exec(ctx, stmt, payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String(), billItem.BillItemSequenceNumber.Int, s.ResourcePath)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bill item: %v", err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
