package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDomainTag struct {
	tagID field.String
	entity.EmptyDomainTag
}

func createMockDomainTag(tagID string) entity.DomainTag {
	return &MockDomainTag{tagID: field.NewString(tagID)}
}

func (m *MockDomainTag) TagID() field.String {
	return m.tagID
}

func TestDomainTagRepo_GetByIDs(t *testing.T) {
	ctx := context.Background()
	db := new(mock_database.Ext)
	userIDs := []string{idutil.ULIDNow()}

	var mockValues []interface{}
	tag := NewTag(entity.EmptyDomainTag{})
	fieldNames, _ := tag.FieldMap()
	for range fieldNames {
		mockValues = append(mockValues, mock.Anything)
	}

	tests := []struct {
		name    string
		wantErr error
		setup   func()
	}{
		{
			"happy case",
			nil,
			func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(userIDs)).Return(true)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Scan", mockValues...).Times(len(userIDs)).Return(nil)
			},
		},
		{
			"error: db.Query error",
			InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "db.Query")},
			func() {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			"error: rows.Close error",
			InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "rows.Err")},
			func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(fmt.Errorf("error"))
			},
		},
		{
			"error: rows.Scan error",
			InternalError{RawError: errors.Wrap(fmt.Errorf("error"), "rows.Scan")},
			func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mockValues...).Once().Return(fmt.Errorf("error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := &DomainTagRepo{}
			if tt.setup != nil {
				tt.setup()
			}
			_, err := ut.GetByIDs(ctx, db, userIDs)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.wantErr)
			}
		})
	}
}

func TestDomainTagRepo_GetByPartnerInternalIDs(t *testing.T) {
	ctx := context.Background()
	db := new(mock_database.Ext)
	userTagPartnerIDs := []string{idutil.ULIDNow()}

	var mockValues []interface{}
	tag := NewTag(entity.EmptyDomainTag{})
	fieldNames, _ := tag.FieldMap()
	for range fieldNames {
		mockValues = append(mockValues, mock.Anything)
	}

	tests := []struct {
		name    string
		wantErr error
		setup   func()
	}{
		{
			"happy case",
			nil,
			func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(userTagPartnerIDs)).Return(true)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Scan", mockValues...).Times(len(userTagPartnerIDs)).Return(nil)
			},
		},
		{
			"error: db.Query error",
			InternalError{
				RawError: fmt.Errorf("db.Query: %v", fmt.Errorf("error")),
			},
			func() {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			"error: rows.Close error",
			InternalError{
				RawError: errors.Wrap(fmt.Errorf("error"), "rows.Err"),
			},
			func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(fmt.Errorf("error"))
			},
		},
		{
			"error: rows.Scan error",
			InternalError{
				RawError: errors.Wrap(fmt.Errorf("error"), "rows.Scan"),
			},
			func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mockValues...).Once().Return(fmt.Errorf("error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := &DomainTagRepo{}
			if tt.setup != nil {
				tt.setup()
			}
			_, err := ut.GetByPartnerInternalIDs(ctx, db, userTagPartnerIDs)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.wantErr)
			}
		})
	}
}
