package paymentfileutils

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	fixedwidth "github.com/ianlopshire/go-fixedwidth"
)

type DirectDebitTextPaymentFileConverter struct{}

type DirectDebitFileHeaderRecord struct {
	DataCategory     int    `fixed:"1,1"`
	TypeCode         int    `fixed:"2,3,right,0"`
	CodeCategory     int    `fixed:"4,4"`
	ConsignorCode    int    `fixed:"5,14,right,0"`
	ConsignorName    string `fixed:"15,54"`
	DepositDate      int    `fixed:"55,58,right,0"`
	BankNumber       int    `fixed:"59,62,right,0"`
	BankName         string `fixed:"63,77"`
	BankBranchNumber int    `fixed:"78,80,right,0"`
	BankBranchName   string `fixed:"81,95"`
	DepositItems     int    `fixed:"96,96"`
	AccountNumber    string `fixed:"97,103"`
	DummyFiller      string `fixed:"104,120"`
}

type DirectDebitFileDataRecord struct {
	DataCategory            int    `fixed:"1,1"`
	DepositBankNumber       int    `fixed:"2,5,right,0"`
	DepositBankName         string `fixed:"6,20"`
	DepositBankBranchNumber int    `fixed:"21,23,right,0"`
	DepositBankBranchName   string `fixed:"24,38"`
	DummyFiller1            string `fixed:"39,42"`
	DepositItems            int    `fixed:"43,43"`
	AccountNumber           string `fixed:"44,50"`
	AccountOwnerName        string `fixed:"51,80"`
	DepositAmount           int    `fixed:"81,90,right,0"`
	NewCustomerCode         string `fixed:"91,91"`
	CustomerNumber          string `fixed:"92,111"`
	ResultCode              string `fixed:"112,112"`
	DummyFiller2            string `fixed:"113,120"`
}

type DirectDebitFileTrailerRecord struct {
	DataCategory      int    `fixed:"1,1"`
	TotalTransactions int    `fixed:"2,7,right,0"`
	TotalAmount       int    `fixed:"8,19,right,0"`
	TransferredNumber int    `fixed:"20,25,right,0"`
	TransferredAmount int    `fixed:"26,37,right,0"`
	FailedNumber      int    `fixed:"38,43,right,0"`
	FailedAmount      int    `fixed:"44,55,right,0"`
	DummyFiller       string `fixed:"56,120"`
}

type DirectDebitFileEndRecord struct {
	DataCategory int    `fixed:"1,1"`
	DummyFiller  string `fixed:"2,120"`
}

type DirectDebitFile struct {
	Header  *DirectDebitFileHeaderRecord
	Data    []*DirectDebitFileDataRecord
	Trailer *DirectDebitFileTrailerRecord
	End     *DirectDebitFileEndRecord
}

func (t *DirectDebitTextPaymentFileConverter) ConvertFromBytesToPaymentFile(ctx context.Context, file []byte) (*PaymentFile, error) {
	lines := bytes.Split(file, []byte("\n"))

	// at least four lines for record types: header, data, trailer, end
	if len(lines) < 4 {
		return nil, fmt.Errorf("invalid file line count")
	}

	headerRecord, err := retrieveTextHeaderRecord(lines[0])
	if err != nil {
		return nil, err
	}

	var dataRecords []*DirectDebitFileDataRecord
	for i := 1; i <= len(lines)-3; i++ {
		dataRecord, err := retrieveTextDataRecord(lines[i])
		if err != nil {
			return nil, err
		}

		dataRecords = append(dataRecords, dataRecord)
	}

	trailerRecord, err := retrieveTextTrailerRecord(lines[len(lines)-2])
	if err != nil {
		return nil, err
	}

	endRecord, err := retrieveTextEndRecord(lines[len(lines)-1])
	if err != nil {
		return nil, err
	}

	directDebitFile := &DirectDebitFile{
		Header:  headerRecord,
		Data:    dataRecords,
		Trailer: trailerRecord,
		End:     endRecord,
	}

	paymentFile := &PaymentFile{
		DirectDebitFile: directDebitFile,
	}

	return paymentFile, nil
}

func retrieveTextHeaderRecord(line []byte) (*DirectDebitFileHeaderRecord, error) {
	var header *DirectDebitFileHeaderRecord

	decoder := fixedwidth.NewDecoder(strings.NewReader(string(line)))
	decoder.SetUseCodepointIndices(true)

	err := decoder.Decode(&header)
	if err != nil {
		return nil, err
	}

	if header.DataCategory != DataTypeHeaderRecord {
		return nil, fmt.Errorf("invalid header record's code category: %v", header.DataCategory)
	}

	return header, nil
}

func retrieveTextDataRecord(line []byte) (*DirectDebitFileDataRecord, error) {
	var data *DirectDebitFileDataRecord
	decoder := fixedwidth.NewDecoder(strings.NewReader(string(line)))
	decoder.SetUseCodepointIndices(true)

	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	if data.DataCategory != DataTypeDataRecord {
		return nil, fmt.Errorf("invalid data record's data category: %v", data.DataCategory)
	}

	return data, nil
}

func retrieveTextTrailerRecord(line []byte) (*DirectDebitFileTrailerRecord, error) {
	var trailer *DirectDebitFileTrailerRecord
	decoder := fixedwidth.NewDecoder(strings.NewReader(string(line)))
	decoder.SetUseCodepointIndices(true)

	err := decoder.Decode(&trailer)
	if err != nil {
		return nil, err
	}

	if trailer.DataCategory != DataTypeTrailerRecord {
		return nil, fmt.Errorf("invalid trailer data's code category: %v", trailer.DataCategory)
	}

	return trailer, nil
}

func retrieveTextEndRecord(line []byte) (*DirectDebitFileEndRecord, error) {
	var end *DirectDebitFileEndRecord
	decoder := fixedwidth.NewDecoder(strings.NewReader(string(line)))
	decoder.SetUseCodepointIndices(true)

	err := decoder.Decode(&end)
	if err != nil {
		return nil, err
	}

	if end.DataCategory != DataTypeEndRecord {
		return nil, fmt.Errorf("invalid end record's code category: %v", end.DataCategory)
	}

	return end, nil
}
