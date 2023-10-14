package log

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	defer infoLogger.SetOutput(os.Stdout)
	infoLogger.SetFlags(log.Lmsgprefix)
	defer infoLogger.SetFlags(log.LstdFlags)

	testcases := []struct {
		inFormat string
		inArgs   []any
		expected string
	}{
		{
			inFormat: "a",
			expected: colorBlue + "a" + colorReset + "\n",
		},
		{
			inFormat: "with format: %d, %s",
			inArgs:   []any{1, "message"},
			expected: colorBlue + "with format: 1, message" + colorReset + "\n",
		},
	}

	for _, tc := range testcases {
		buf := bytes.Buffer{}
		infoLogger.SetOutput(&buf)
		Info(tc.inFormat, tc.inArgs...)
		assert.Equal(t, tc.expected, buf.String())
	}
}

func TestWarn(t *testing.T) {
	defer warnLogger.SetOutput(os.Stdout)
	warnLogger.SetFlags(log.Lmsgprefix)
	defer warnLogger.SetFlags(log.LstdFlags)

	testcases := []struct {
		inFormat string
		inArgs   []any
		expected string
	}{
		{
			inFormat: "a",
			expected: colorYellow + "a" + colorReset + "\n",
		},
		{
			inFormat: "with format: %d, %s",
			inArgs:   []any{1, "message"},
			expected: colorYellow + "with format: 1, message" + colorReset + "\n",
		},
	}

	for _, tc := range testcases {
		buf := bytes.Buffer{}
		warnLogger.SetOutput(&buf)
		Warn(tc.inFormat, tc.inArgs...)
		assert.Equal(t, tc.expected, buf.String())
	}
}
