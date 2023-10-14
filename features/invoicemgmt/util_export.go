package invoicemgmt

import "fmt"

func checkCSVHeaderForExport(expected []string, actual []string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("Expected header length to be %d got %d", len(expected), len(actual))
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			return fmt.Errorf("Expected header name to be %s got %s", expected[i], actual[i])
		}
	}

	return nil
}
