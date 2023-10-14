package utils

import (
	"fmt"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
)

func EncodeByteToShiftJIS(byteContent []byte) ([]byte, error) {
	jpEncoder := japanese.ShiftJIS.NewEncoder()
	encodedByte, err := jpEncoder.Bytes(byteContent)
	if err != nil {
		return nil, fmt.Errorf("cannot encrypt byte content err: %v", err)
	}
	return encodedByte, nil
}

func DecodeByteToShiftJIS(byteContent []byte) ([]byte, error) {
	jpDecoded := japanese.ShiftJIS.NewDecoder()
	decodedByte, err := jpDecoded.Bytes(byteContent)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt byte content err: %v", err)
	}

	return decodedByte, nil
}

func GetShiftJISDecoder() *encoding.Decoder {
	return japanese.ShiftJIS.NewDecoder()
}
