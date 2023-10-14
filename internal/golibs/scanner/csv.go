package scanner

import (
	"encoding/csv"
	"io"
	"strings"
)

type CSVScanner struct {
	reader *csv.Reader
	Head   map[string]int
	row    []string
	curRow int
}

func NewCSVScanner(r io.Reader) CSVScanner {
	csvReader := csv.NewReader(r)
	head, err := csvReader.Read()
	if err != nil {
		return CSVScanner{}
	}
	h := make(map[string]int)
	for index, column := range head {
		col := strings.ToLower(column)
		h[col] = index
	}
	return CSVScanner{reader: csvReader, Head: h, curRow: 1, row: head}
}

func (cs *CSVScanner) Scan() bool {
	r, err := cs.reader.Read()
	if err != nil {
		return false
	}
	cs.row = r
	cs.curRow++
	return true
}

func (cs *CSVScanner) Text(col string) string {
	idx, ok := cs.Head[strings.ToLower(col)]
	if !ok {
		return ""
	}
	return strings.TrimSpace(cs.row[idx])
}

func (cs *CSVScanner) RawText(col string) string {
	idx, ok := cs.Head[strings.ToLower(col)]
	if !ok {
		return ""
	}
	return cs.row[idx]
}

func (cs *CSVScanner) GetCurRow() int {
	return cs.curRow
}

func (cs *CSVScanner) GetRow() []string {
	return cs.row
}
