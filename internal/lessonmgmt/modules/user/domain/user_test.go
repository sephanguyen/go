package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeachers_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name    string
		input   Teachers
		isValid bool
	}{
		{
			name: "full field",
			input: Teachers{
				{
					ID:        "id-1",
					Name:      "name-1",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "id-2",
					Name:      "name-2",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			isValid: true,
		},
		{
			name: "missing name",
			input: Teachers{
				{
					ID:        "id-1",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "id-2",
					Name:      "name-2",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.IsValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestStudent_TestNewStudent(t *testing.T) {
	id := "id"
	name := "name"
	email := "email"

	student := NewStudent(id, name, email)
	assert.Equal(t, student.ID, id)
	assert.Equal(t, student.Name, name)
	assert.Equal(t, student.Email, email)
}
