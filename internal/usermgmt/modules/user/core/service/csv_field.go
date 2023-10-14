package service

import "strings"

type CsvField struct {
	Text  string
	Exist bool
}

func (c *CsvField) UnmarshalCSV(text string) error {
	if c != nil {
		if strings.TrimSpace(text) == "" {
			c.Text = ""
			c.Exist = false
		} else {
			c.Text = strings.TrimSpace(text)
			c.Exist = true
		}
	}
	return nil
}

func (c *CsvField) String() string {
	if c == nil {
		return ""
	}
	return c.Text
}

func (c *CsvField) CheckExist() bool {
	return c != nil && c.Exist
}
