package invoicemgmt

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoicemgmt_entities "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	import_service "github.com/manabie-com/backend/internal/invoicemgmt/services/import_service"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) aRequestPayloadFileWithPartnerBankRecord(ctx context.Context, existNotExist, recordType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	req := &invoice_pb.ImportPartnerBankRequest{}
	req.Payload, err = s.generateImportPartnerBankPayload(ctx, recordType, false)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.StepState.Request = req

	if existNotExist == "existing" {
		ctx, err := s.signedAsAccount(ctx, "school admin")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ctx, err = s.createPartnerBankRecordsFromRequest(ctx, recordType)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importsPartnerBankRecords(ctx context.Context, recordType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.createPartnerBankRecordsFromRequest(ctx, recordType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importsInvalidPartnerBankRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.StepState.Request.(*invoice_pb.ImportPartnerBankRequest)
	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewImportMasterDataServiceClient(s.InvoiceMgmtConn).ImportPartnerBank(contextWithToken(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createPartnerBankRecordsFromRequest(ctx context.Context, recordType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.StepState.Request.(*invoice_pb.ImportPartnerBankRequest)
	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewImportMasterDataServiceClient(s.InvoiceMgmtConn).ImportPartnerBank(contextWithToken(ctx), req)

	ctx, err := s.retrievePartnerBankRecordsImported(ctx, recordType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// Creates the payload in bytes, containing the CSV contents
//
//nolint:gosec
func (s *suite) generateImportPartnerBankPayload(ctx context.Context, recordType string, isArchived bool) ([]byte, error) {
	stepState := StepStateFromContext(ctx)
	var payload []byte
	var csv string
	var err error
	// generate header csv
	headerTitles := import_service.GenerateHeaderTitles(invoice_pb.ImportMasterAction_PARTNER_BANK.String())
	headerStr := strings.Join(headerTitles, ",")
	singleRemarks := "single-valid-remarks"
	singleLimitRemarks := "single-valid-limit-remarks"
	multipleRemarks := "multiple-with-default-valid-remarks"
	if isArchived {
		singleRemarks += "-archived"
		multipleRemarks += "-archived"
	}
	switch recordType {
	case "single-valid":
		csv = `%v
		,%v,ｶ)ｹ-ｲ-ｼ-,%v,ﾅﾝﾄ,%v,ｲｺﾏ%v,1,%v,,%v,,`
		payload = []byte(fmt.Sprintf(csv, headerStr, fmt.Sprint(rand.Intn(9999999999)), fmt.Sprint(rand.Intn(9999)), fmt.Sprint(rand.Intn(999)), fmt.Sprint(rand.Intn(99)), fmt.Sprintf("%07d", rand.Intn(9999999)), singleRemarks))
		if (isArchived) && len(stepState.PartnerBankIDs) > 0 {

			partnerBank := &entities.PartnerBank{}
			fields, _ := partnerBank.FieldMap()

			query := fmt.Sprintf("SELECT %s FROM %s WHERE partner_bank_id = $1", strings.Join(fields, ","), partnerBank.TableName())
			err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, stepState.PartnerBankIDs[0]).ScanOne(partnerBank)
			if err != nil {
				return nil, fmt.Errorf("error on retrieving partner bank record: %v", err)
			}

			csv = `%v
			%v,%v,%v,%v,%v,%v,%v,1,%v,1,%v,,`

			payload = []byte(fmt.Sprintf(csv, headerStr, stepState.PartnerBankIDs[0], partnerBank.ConsignorCode.String, partnerBank.ConsignorName.String, partnerBank.BankNumber.String, partnerBank.BankName.String, partnerBank.BankBranchNumber.String, partnerBank.BankBranchName.String, partnerBank.AccountNumber.String, singleRemarks))
		}
	case "single-valid-limit":
		csv = `%v
		,%v,ｶ)ｹ-ｲ-ｼ-,%v,ﾅﾝﾄ,%v,ｲｺﾏ%v,1,%v,,%v,,10`
		payload = []byte(fmt.Sprintf(csv, headerStr, fmt.Sprint(rand.Intn(9999999999)), fmt.Sprint(rand.Intn(9999)), fmt.Sprint(rand.Intn(999)), fmt.Sprint(rand.Intn(99)), fmt.Sprintf("%07d", rand.Intn(9999999)), singleLimitRemarks))
		if (isArchived) && len(stepState.PartnerBankIDs) > 0 {
			partnerBank := &entities.PartnerBank{}
			fields, _ := partnerBank.FieldMap()

			query := fmt.Sprintf("SELECT %s FROM %s WHERE partner_bank_id = $1", strings.Join(fields, ","), partnerBank.TableName())
			err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, stepState.PartnerBankIDs[0]).ScanOne(partnerBank)
			if err != nil {
				return nil, fmt.Errorf("error on retrieving partner bank record: %v", err)
			}

			csv = `%v
			%v,%v,%v,%v,%v,%v,%v,1,%v,1,%v,,10`

			payload = []byte(fmt.Sprintf(csv, headerStr, stepState.PartnerBankIDs[0], partnerBank.ConsignorCode.String, partnerBank.ConsignorName.String, partnerBank.BankNumber.String, partnerBank.BankName.String, partnerBank.BankBranchNumber.String, partnerBank.BankBranchName.String, partnerBank.AccountNumber.String, singleLimitRemarks))
		}
	case "multiple-with-default-valid":
		csv = `%v
		,%v,ｶ)ｹ-ｲ-ｼ-,%v,ﾅﾝﾄ,%v,ｲｺﾏ%v,1,%v,,%v,1,2
		,%v,ｶ)ｹ-ｲ-ｼ-,%v,ﾅﾝﾄ,%v,ｲｺﾏ%v,1,%v,,%v,,2`

		payload = []byte(fmt.Sprintf(csv, headerStr, fmt.Sprint(rand.Intn(9999999999)), fmt.Sprint(rand.Intn(9999)), fmt.Sprint(rand.Intn(999)), fmt.Sprint(rand.Intn(99)), fmt.Sprintf("%07d", rand.Intn(9999999)), multipleRemarks, fmt.Sprint(rand.Intn(9999999999)), fmt.Sprint(rand.Intn(9999)), fmt.Sprint(rand.Intn(999)), fmt.Sprint(rand.Intn(99)), fmt.Sprintf("%07d", rand.Intn(9999999)), multipleRemarks))

		if (isArchived) && len(stepState.PartnerBankIDs) > 0 {
			var partnerBanks PartnerBanks
			var partnerBankIDs pgtype.TextArray
			_ = partnerBankIDs.Set(stepState.PartnerBankIDs)
			partnerBank := &entities.PartnerBank{}
			fields, _ := partnerBank.FieldMap()

			query := fmt.Sprintf("SELECT %s FROM %s WHERE partner_bank_id = ANY($1)", strings.Join(fields, ","), partnerBank.TableName())
			err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, partnerBankIDs).ScanAll(&partnerBanks)
			if err != nil {
				return nil, fmt.Errorf("error on retrieving partner bank records: %v", err)
			}
			// amend partner bank id and is archive
			csv = `%v
			%v,%v,%v,%v,%v,%v,%v,1,%v,1,%v,1,2
			%v,%v,%v,%v,%v,%v,%v,1,%v,1,%v,,2`

			payload = []byte(fmt.Sprintf(csv, headerStr, partnerBanks[0].PartnerBankID.String, partnerBanks[0].ConsignorCode.String, partnerBanks[0].ConsignorName.String, partnerBanks[0].BankNumber.String, partnerBanks[0].BankName.String, partnerBanks[0].BankBranchNumber.String, partnerBanks[0].BankBranchName.String, partnerBanks[0].AccountNumber.String, multipleRemarks, partnerBanks[1].PartnerBankID.String, partnerBanks[1].ConsignorCode.String, partnerBanks[1].ConsignorName.String, partnerBanks[0].BankNumber.String, partnerBanks[1].BankName.String, partnerBanks[1].BankBranchNumber.String, partnerBanks[1].BankBranchName.String, partnerBanks[1].AccountNumber.String, multipleRemarks))
		}

	case "multiple-with-default-invalid-required":
		csv = `%v
		,000000481923,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,test-branch,1,1234567,,,1,
		,0000004819,ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,test-branch,1,1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,"032674",ﾅﾝﾄ,150,test-branch,1,1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄ,150,test-branch,1,1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,"1520",test-branch,1,1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏ,1,1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,"12",1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,"12345678",,,,`

		payload = []byte(fmt.Sprintf(csv, headerStr))

	case "multiple-with-default-invalid-format":
		csv = `%v
		,aswasad,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,test-branch,1,1234567,,,1,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,wasss,ﾅﾝﾄ,150,test-branch,1,1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,testInvalid,test-branch,1,1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,150,ｲｺﾏ,"asd",1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,150,ｲｺﾏ,"8",1234567,,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,150,ｲｺﾏ,1,1234567,asd,,,
		,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,1234567,,,asd,`

		payload = []byte(fmt.Sprintf(csv, headerStr))
	}

	return payload, err
}

func (s *suite) partnerBankCsvIsImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if s.StepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to import partner bank csv err: %v", s.StepState.ResponseErr)
	}

	errors := s.StepState.Response.(*invoice_pb.ImportPartnerBankResponse).Errors
	if len(errors) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("expected empty errors arr from response but received %v: %v", len(errors), errors))
	}

	ctx, err := s.checkRecordsImported(ctx, false)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

type PartnerBanks []*invoicemgmt_entities.PartnerBank

func (u *PartnerBanks) Add() database.Entity {
	e := &invoicemgmt_entities.PartnerBank{}
	*u = append(*u, e)

	return e
}

func (s *suite) partnerBankCsvIsArchivedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.checkRecordsImported(ctx, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkRecordsImported(ctx context.Context, isArchived bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count int32
	var partnerBanks pgtype.TextArray
	_ = partnerBanks.Set(stepState.PartnerBankIDs)
	query := `SELECT COUNT(*) FROM partner_bank WHERE partner_bank_id = ANY($1)`
	if isArchived {
		query = `SELECT COUNT(*) FROM partner_bank WHERE partner_bank_id = ANY($1) AND is_archived = true`
	}

	err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, partnerBanks).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error partner bank records imported: %v", err)
	}

	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error no partner bank records imported")
	}

	if count != int32(len(stepState.PartnerBankIDs)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of imported records not match")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) archivesPartnerBankRecords(ctx context.Context, recordType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	req := &invoice_pb.ImportPartnerBankRequest{}

	req.Payload, err = s.generateImportPartnerBankPayload(ctx, recordType, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// import record and check successfully
	s.StepState.Request = req

	ctx, err = s.createPartnerBankRecordsFromRequest(ctx, recordType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) partnerBankCsvIsImportedUnsuccessfully(ctx context.Context, recordType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	errors := s.StepState.Response.(*invoice_pb.ImportPartnerBankResponse).Errors
	if len(errors) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf(fmt.Sprintf("expected empty errors arr from response but received null"))
	}

	err := checkPartnerBankCsvRowsErrors(recordType, errors)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func checkPartnerBankCsvRowsErrors(recordType string, errors []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError) error {
	var expectedErr []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError

	switch recordType {
	case "multiple-with-default-invalid-required":
		expectedErr = []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
			{
				RowNumber: 2,
				Error:     "unable to parse partner bank detail: invalid consignor code digit limit",
			},
			{
				RowNumber: 3,
				Error:     "unable to parse partner bank detail: invalid consignor name character limit",
			},
			{
				RowNumber: 4,
				Error:     "unable to parse partner bank detail: invalid bank number digit limit",
			},
			{
				RowNumber: 5,
				Error:     "unable to parse partner bank detail: invalid bank name character limit",
			},
			{
				RowNumber: 6,
				Error:     "unable to parse partner bank detail: invalid bank branch number digit limit",
			},
			{
				RowNumber: 7,
				Error:     "unable to parse partner bank detail: invalid bank branch name character limit",
			},
			{
				RowNumber: 8,
				Error:     "unable to parse partner bank detail: invalid deposit items digit limit",
			},
			{
				RowNumber: 9,
				Error:     "unable to parse partner bank detail: the account number can only accept 7 digit numbers",
			},
		}

	case "multiple-with-default-invalid-format":
		expectedErr = []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
			{
				RowNumber: 2,
				Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "consignor_code field has invalid half width number").Error()),
			},
			{
				RowNumber: 3,
				Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "bank_number field has invalid half width number").Error()),
			},
			{
				RowNumber: 4,
				Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "bank_branch_number field has invalid half width number").Error()),
			},
			{
				RowNumber: 5,
				Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "deposit_items field has invalid half width number").Error()),
			},
			{
				RowNumber: 6,
				Error:     "unable to parse partner bank detail: invalid deposit items account",
			},
			{
				RowNumber: 7,
				Error:     "unable to parse partner bank detail: invalid archive value",
			},
			{
				RowNumber: 8,
				Error:     "unable to parse partner bank detail: invalid default value",
			},
		}
	}
	if len(errors) != len(expectedErr) {
		return fmt.Errorf(fmt.Sprintf("expected %v errors length from response but got %v", len(expectedErr), len(errors)))
	}

	for i, csvError := range errors {
		if csvError.RowNumber != expectedErr[i].RowNumber {
			return fmt.Errorf(fmt.Sprintf("expected RowNumber %v from response error list but got %v", expectedErr[i].RowNumber, csvError.RowNumber))
		}

		if csvError.Error != expectedErr[i].Error {
			return fmt.Errorf(fmt.Sprintf("expected Error %v from response error list but got %v", expectedErr[i].Error, csvError.Error))
		}
	}
	return nil
}

func (s *suite) retrievePartnerBankRecordsImported(ctx context.Context, recordType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.PartnerBankIDs = []string{}

	req := s.StepState.Request.(*invoice_pb.ImportPartnerBankRequest).Payload
	buf := bytes.NewBuffer(req)
	content := buf.String()
	eachLine := strings.Split(content, "\n")

	// ignore header data
	// maximum valid records imported in request is two
	for i := 1; i < len(eachLine); i++ {
		lineValuesInArray := strings.Split(eachLine[i], ",")
		var err error

		isArchivedValue := "false"
		partnerBankID := strings.TrimSpace(lineValuesInArray[0])
		consignorCode := strings.TrimSpace(lineValuesInArray[1])
		consignorName := strings.TrimSpace(lineValuesInArray[2])
		bankNumber := strings.TrimSpace(lineValuesInArray[3])
		bankName := strings.TrimSpace(lineValuesInArray[4])
		bankBranchNumber := strings.TrimSpace(lineValuesInArray[5])
		bankBranchName := strings.TrimSpace(lineValuesInArray[6])
		accountNumber := strings.TrimSpace(lineValuesInArray[8])
		isArchived := strings.TrimSpace(lineValuesInArray[9])
		remarks := strings.TrimSpace(lineValuesInArray[10])
		recordLimitStr := strings.TrimSpace(lineValuesInArray[12])

		partnerBank := &entities.PartnerBank{}
		fields, _ := partnerBank.FieldMap()

		query := fmt.Sprintf("SELECT %s FROM %s WHERE account_number = $1", strings.Join(fields, ","), partnerBank.TableName())

		switch isArchived {
		case "1":
			isArchivedValue = "true"
			query = fmt.Sprintf("SELECT %s FROM %s WHERE partner_bank_id = $1", strings.Join(fields, ","), partnerBank.TableName())
			err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, &partnerBankID).ScanOne(partnerBank)
		default:
			err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, &accountNumber).ScanOne(partnerBank)
		}
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting partner bank err: %v", err)
		}

		depositItemInt, err := strconv.Atoi(strings.TrimSpace(lineValuesInArray[7]))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on converting deposit item value to integer: %v", err)
		}

		var recordLimitInt int
		if recordLimitStr != "" {
			recordLimitInt, err = strconv.Atoi(recordLimitStr)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error on converting record_limit to integer: %v", err)
			}
		}

		// to compare partner bank archive to string on common function isEqual
		partnerBankArchived := "false"
		if partnerBank.IsArchived.Bool {
			partnerBankArchived = "true"
		}

		if err := multierr.Combine(
			isEqual(consignorCode, partnerBank.ConsignorCode.String, "consignor_code"),
			isEqual(consignorName, partnerBank.ConsignorName.String, "consignor_name"),
			isEqual(bankNumber, partnerBank.BankNumber.String, "bank_number"),
			isEqual(bankName, partnerBank.BankName.String, "bank_name"),
			isEqual(bankBranchNumber, partnerBank.BankBranchNumber.String, "bank_branch_number"),
			isEqual(bankBranchName, partnerBank.BankBranchName.String, "bank_branch_name"),
			isEqual(constant.PartnerBankDepositItems[depositItemInt], partnerBank.DepositItems.String, "deposit_items"),
			isEqual(isArchivedValue, partnerBankArchived, "is_archived"),
			isEqual(remarks, partnerBank.Remarks.String, "remarks"),
			isEqualInt(recordLimitInt, int(partnerBank.RecordLimit.Int), "record_limit"),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.PartnerBankIDs = append(stepState.PartnerBankIDs, partnerBank.PartnerBankID.String)
	}

	return StepStateToContext(ctx, stepState), nil
}
