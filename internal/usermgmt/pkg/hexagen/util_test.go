package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferFromImportPaths(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		input    []string
		expected []byte
	}

	testCases := []testCase{
		{
			name: "only std pkg",
			input: []string{
				`log`, `fmt`,
			},
			expected: []byte("import (\n\t\"log\"\n\t\"fmt\"\n)\n"),
		},
		{
			name: "only internal pkg",
			input: []string{
				`github.com/manabie-com/backend/internal/usermgmt/pkg/field`, `github.com/manabie-com/backend/internal/usermgmt/pkg/constant`,
			},
			expected: []byte("import (\n\t\"github.com/manabie-com/backend/internal/usermgmt/pkg/field\"\n\t\"github.com/manabie-com/backend/internal/usermgmt/pkg/constant\"\n)\n"),
		},
		{
			name: "only external pkg",
			input: []string{
				`github.com/pkg/errors`, `go.uber.org/zap`,
			},
			expected: []byte("import (\n\t\"github.com/pkg/errors\"\n\t\"go.uber.org/zap\"\n)\n"),
		},
		{
			name: "have mixed all types of import paths",
			input: []string{
				`log`, `github.com/pkg/errors`, `github.com/manabie-com/backend/internal/usermgmt/pkg/field`, `fmt`, `go.uber.org/zap`, `github.com/manabie-com/backend/internal/usermgmt/pkg/constant`,
			},
			expected: []byte("import (\n\t\"log\"\n\t\"fmt\"\n\n\t\"github.com/manabie-com/backend/internal/usermgmt/pkg/field\"\n\t\"github.com/manabie-com/backend/internal/usermgmt/pkg/constant\"\n\n\t\"github.com/pkg/errors\"\n\t\"go.uber.org/zap\"\n)\n"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualBuffer, err := BufferFromImportPaths(testCase.input)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, actualBuffer)
		})
	}
}
