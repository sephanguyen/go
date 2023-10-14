package invoicemgmt

import (
	"context"
	"fmt"

	pfutils "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_utils"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	fixedwidth "github.com/ianlopshire/go-fixedwidth"
)

func (s *suite) generateDirectDebitFile(ctx context.Context) (string, error) {
	stepState := StepStateFromContext(ctx)

	// Create the header record part
	headerRecord := pfutils.DirectDebitFileHeaderRecord{
		DataCategory: pfutils.DataTypeHeaderRecord,
	}

	headerRecordBytes, err := fixedwidth.Marshal(headerRecord)
	if err != nil {
		return "", fmt.Errorf("error marshalling header record: %v", err)
	}

	// Create the data record part
	dataRecordsBytes := make([][]byte, 0)

	for _, paymentToValidate := range stepState.PaymentListToValidate {
		exactAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(paymentToValidate.Amount, "2")
		if err != nil {
			return "", err
		}

		dataRecord := &pfutils.DirectDebitFileDataRecord{
			DataCategory:   pfutils.DataTypeDataRecord,
			CustomerNumber: fmt.Sprintf("%v", paymentToValidate.PaymentSequenceNumber.Int),
			DepositAmount:  int(exactAmount),
			ResultCode:     paymentToValidate.ResultCode.String,
		}

		dataRecordBytesData, err := fixedwidth.Marshal(dataRecord)
		if err != nil {
			return "", fmt.Errorf("error marshalling data record: %v", err)
		}

		dataRecordsBytes = append(dataRecordsBytes, dataRecordBytesData)
	}

	// Create the trailer record part
	trailerRecord := pfutils.DirectDebitFileTrailerRecord{
		DataCategory: pfutils.DataTypeTrailerRecord,
	}

	trailerRecordBytes, err := fixedwidth.Marshal(trailerRecord)
	if err != nil {
		return "", fmt.Errorf("error marshalling trailer record: %v", err)
	}

	// Create the end record part
	endRecord := pfutils.DirectDebitFileEndRecord{
		DataCategory: pfutils.DataTypeEndRecord,
	}

	endRecordBytes, err := fixedwidth.Marshal(endRecord)
	if err != nil {
		return "", fmt.Errorf("error marshalling end record: %v", err)
	}

	// Combine all file parts into a single string before converting into bytes
	fileContentsStr := fmt.Sprintf("%v\n", string(headerRecordBytes))
	for _, dataBytes := range dataRecordsBytes {
		fileContentsStr += fmt.Sprintf("%v\n", string(dataBytes))
	}
	fileContentsStr += fmt.Sprintf("%v\n%v", string(trailerRecordBytes), string(endRecordBytes))

	return fileContentsStr, nil
}
