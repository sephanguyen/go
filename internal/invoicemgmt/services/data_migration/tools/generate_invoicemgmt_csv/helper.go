package generator

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
)

const InvoiceRawData = "invoice_raw_data"

func sortSliceByIndex(s [][]string, index int) {
	sort.Slice(s, func(i, j int) bool {
		id1, _ := strconv.ParseInt(s[i][index], 10, 64)
		id2, _ := strconv.ParseInt(s[j][index], 10, 64)

		return id1 < id2
	})
}

func assignRowIDToLines(s [][]string, index int) {
	for i, line := range s {
		line[index] = strconv.Itoa(i + 1)
	}
}

func writeLines(csvWrite *csv.Writer, lines [][]string, name string) error {
	err := csvWrite.WriteAll(lines)
	if err != nil {
		return fmt.Errorf("error writing in %s CSV file err: %v", name, err)
	}

	csvWrite.Flush()
	if err := csvWrite.Error(); err != nil {
		return fmt.Errorf("error encountered flushing in %s CSV file err: %v", name, err)
	}

	return nil
}
