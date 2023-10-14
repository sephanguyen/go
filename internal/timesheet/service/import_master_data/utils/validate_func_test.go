package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCsvHeader(t *testing.T) {
	t.Run("Happy case", func(t *testing.T) {
		err := ValidateCsvHeader(
			2,
			[]string{
				"id",
				"name",
			},
			[]string{
				"id",
				"name",
			},
		)
		require.Nil(t, err)
	})

	t.Run("Failed by inlvaid number columns", func(t *testing.T) {
		expectedNumberColumns := 2
		columnNames := []string{
			"id",
		}
		expectedColumnNames := []string{
			"id",
			"name",
		}
		err := ValidateCsvHeader(
			expectedNumberColumns,
			columnNames,
			expectedColumnNames,
		)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("csv file invalid format - number of column should be %d", expectedNumberColumns))
	})

	t.Run("Failed by unmatch colmn names and expected column names", func(t *testing.T) {
		expectedNumberColumns := 2
		columnNames := []string{
			"Number",
			"name",
		}
		expectedColumnNames := []string{
			"id",
			"name",
		}
		expectedWrongIdx := 0
		err := ValidateCsvHeader(
			expectedNumberColumns,
			columnNames,
			expectedColumnNames,
		)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf(
			"csv file invalid format - %s column (toLowerCase) should be '%s'",
			NumberNames[expectedWrongIdx],
			expectedColumnNames[expectedWrongIdx],
		))
	})
}
