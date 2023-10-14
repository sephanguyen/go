package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileNameSuffix(t *testing.T) {
	type testCase struct {
		name           string
		input          time.Time
		expectedOutput string
	}

	testCases := []testCase{
		{
			name:           "file suffix for a specific date",
			input:          time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
			expectedOutput: "20230101",
		},
		{
			name:           "file suffix for a specific date",
			input:          time.Date(2023, 12, 12, 0, 0, 0, 0, time.Local),
			expectedOutput: "20231212",
		},
	}

	t.Parallel()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, FileNameSuffix(testCase.input))
		})
	}
}

func TestFileNameToDownload(t *testing.T) {
	type testCase struct {
		name                string
		inputFileNamePrefix string
		inputFileUploadTime time.Time
		expectedOutput      string
	}

	testCases := []testCase{
		{
			name:                "managara base file name",
			inputFileNamePrefix: "base",
			inputFileUploadTime: time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
			expectedOutput:      "base" + FileNameInfix + "20230101" + FileNameExtension,
		},

		{
			name:                "managara high school file name",
			inputFileNamePrefix: "hs",
			inputFileUploadTime: time.Date(2023, 12, 12, 0, 0, 0, 0, time.Local),
			expectedOutput:      "hs" + FileNameInfix + "20231212" + FileNameExtension,
		},
	}

	t.Parallel()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, FileNameToDownload(testCase.inputFileNamePrefix, testCase.inputFileUploadTime))
		})
	}
}

func TestTestLoadJSTTime(t *testing.T) {
	triggeredTimeInUTC := time.Date(2023, 01, 01, 19, 0, 0, 0, time.UTC)

	tokyoLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "20230102", FileNameSuffix(triggeredTimeInUTC.In(tokyoLocation)))

	cliInputDateStr := "20220101"
	clitInputTime, err := time.ParseInLocation(TimeFormatYYYYMMDD, cliInputDateStr, tokyoLocation)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "20220101", FileNameSuffix(clitInputTime))
}
