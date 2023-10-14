package service

import (
	"encoding/csv"
	"github.com/gocarina/gocsv"
	"strings"
	"testing"
)

func Test_CSV_Field(t *testing.T) {
	t.Parallel()

	type row struct {
		Field  *CsvField `csv:"id"`
		Field2 *CsvField `csv:"name"`
	}

	exampleCSV := `id
	1
1
    
`
	var rows []row
	r := csv.NewReader(strings.NewReader(exampleCSV))
	err := gocsv.UnmarshalCSV(r, &rows)
	if err != nil {
		t.Fatal(err.Error())
	}
	if rows[1].Field.String() != "1" {
		t.Fatalf("Expected %q, but got %q", "foo", string(rows[0].Field.Text))
	}
	if !rows[1].Field.CheckExist() {
		t.Fatalf("Expected %t, but got %t", true, false)
	}
	if rows[1].Field2.String() != "" {
		t.Fatalf("Expected %s, but got %s", "", rows[1].Field2.String())
	}
}
