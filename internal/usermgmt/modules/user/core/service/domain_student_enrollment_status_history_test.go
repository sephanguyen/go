package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	ppb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDomainStudent_validateEnrollmentStatusCreateRequestFromOrder(t *testing.T) {
	type args struct {
		enrollmentStatus string
		order            int
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL",
			args: args{
				enrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
				order:            1,
			},
			wantErr: nil,
		},
		{
			name: "happy case: StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL",
			args: args{
				enrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
				order:            1,
			},
			wantErr: nil,
		},
		{
			name: "happy case: StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY",
			args: args{
				enrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
				order:            1,
			},
			wantErr: nil,
		},
		{
			name: "bad case: StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED",
			args: args{
				enrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
				order:            1,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnrollmentStatusCreateRequestFromOrder(tt.args.enrollmentStatus, tt.args.order)
			if tt.wantErr != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)

				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
		})
	}
}

func TestDomainStudent_validateEntityEnrollmentStatusHistory(t *testing.T) {
	t.Parallel()
	type args struct {
		currentEnrollmentStatus entity.DomainEnrollmentStatusHistory
		reqEnrollmentStatus     entity.DomainEnrollmentStatusHistory
		latestEnrollmentStatus  entity.DomainEnrollmentStatusHistory
		idx                     int
	}
	now := time.Now()
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				currentEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				latestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				idx: 1,
			},
			wantErr: nil,
		},
		{
			name: "bad case: Cannot change Non-Potential to any status",
			args: args{
				currentEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-39*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				latestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				idx: 1,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", 1),
			},
		},
		{
			name: "bad case: Cannot update when req start date is smaller DB start date",
			args: args{
				currentEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-100*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				latestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				idx: 1,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.start_date", 1),
			},
		},
		{
			name: "happy case: Can change without compare millisecond in start_date",
			args: args{
				currentEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(2*time.Millisecond),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				latestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(2*time.Millisecond),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				idx: 1,
			},
			wantErr: nil,
		},
		{
			name: "happy case: can update with start date req is zero, change enrollment and existed record in DB",
			args: args{
				currentEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
				},
				latestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				idx: 1,
			},
			wantErr: nil,
		},
		{
			name: "bad case: start_date is diff but enrollment status is the same from client",
			args: args{
				currentEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(time.Hour)),
				},
				latestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				idx: 1,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.start_date", 1),
			},
		},
		{
			name: "bad case: enrollment status is diff but start_date is the same from client",
			args: args{
				currentEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				latestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				idx: 1,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.start_date", 1),
			},
		},
		{
			name: "happy case: Can change any status to temporary",
			args: args{
				currentEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
					now.Add(-39*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				latestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				idx: 1,
			},
			wantErr: nil,
		},
		{
			name: "happy case: can update when start_date is zero in req, enrollmentStatus is the same",
			args: args{
				currentEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
				},
				latestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				idx: 1,
			},
			wantErr: nil,
		},
		{
			name: "happy case: can update temporary status to another",
			args: args{
				currentEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
					startDate:        field.NewTime(now),
				},
				latestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				idx: 1,
			},
			wantErr: nil,
		},
		{
			name: "happy case: can update end date of temporary status",
			args: args{
				currentEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
					endDate:          field.NewTime(now),
				},
				latestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
				idx: 1,
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: can update end date of temporary status when there's a status in future",
			args: args{
				currentEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
					endDate:          field.NewTime(now),
				},
				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
					endDate:          field.NewTime(now.Add(48 * time.Hour)),
				},
				latestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
					startDate:        field.NewTime(now.Add(24 * time.Hour)),
				},
				idx: 1,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.end_date", 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEntityEnrollmentStatusHistory(tt.args.currentEnrollmentStatus, tt.args.reqEnrollmentStatus, tt.args.latestEnrollmentStatus, tt.args.idx)
			if tt.wantErr != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)

				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
		})
	}
}

func TestDomainStudent_hasActivatedEnrollmentStatusHistory(t *testing.T) {
	serviceMock, student := DomainStudentServiceMock()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()

	type args struct {
		ctx                          context.Context
		db                           libdatabase.QueryExecer
		reqEnrollmentStatusHistories entity.DomainEnrollmentStatusHistories
		studentID                    string
	}
	tests := []struct {
		name         string
		args         args
		setup        func(serviceMock *prepareDomainStudentMock)
		err          error
		expectedResp bool
	}{
		{
			name: "has activated enrollment in the req client",
			args: args{
				ctx: ctx,
				reqEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
						time.Now().Add(-100*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1),
				},
				studentID: "studentID-1",
			},
			setup: func(sm *prepareDomainStudentMock) {

				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil,
				)
			},
			err:          nil,
			expectedResp: true,
		},
		{
			name: "has activated enrollment in database",
			args: args{
				ctx:                          ctx,
				reqEnrollmentStatusHistories: nil,
				studentID:                    "studentID-1",
			},
			setup: func(sm *prepareDomainStudentMock) {

				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					},
					nil,
				)
			},
			err:          nil,
			expectedResp: true,
		},
		{
			name: "has activated enrollment in database and req client",
			args: args{
				ctx: ctx,
				reqEnrollmentStatusHistories: []entity.DomainEnrollmentStatusHistory{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
						time.Now().Add(-100*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1),
				},
				studentID: "studentID-1",
			},
			setup: func(sm *prepareDomainStudentMock) {

				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					},
					nil,
				)
			},
			err:          nil,
			expectedResp: true,
		},
		{
			name: "don't has activated enrollment in the req client and database",
			args: args{
				ctx: ctx,
				reqEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-100*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1),
				},
				studentID: "studentID-1",
			},
			setup: func(sm *prepareDomainStudentMock) {

				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					},
					nil,
				)
			},
			err:          nil,
			expectedResp: false,
		},
		{
			name: "don't has activated enrollment in the req client",
			args: args{
				ctx: ctx,
				reqEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-100*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1),
				},
				studentID: "studentID-1",
			},
			setup: func(sm *prepareDomainStudentMock) {

				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					},
					nil,
				)
			},
			err:          nil,
			expectedResp: true,
		},
		{
			name: "don't has activated enrollment in the database",
			args: args{
				ctx: ctx,
				reqEnrollmentStatusHistories: []entity.DomainEnrollmentStatusHistory{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
						time.Now().Add(-100*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1),
				},
				studentID: "studentID-1",
			},
			setup: func(sm *prepareDomainStudentMock) {

				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					},
					nil,
				)
			},
			err:          nil,
			expectedResp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(&serviceMock)
			hasActivatedEnrollmentStatus, err := student.hasActivatedEnrollmentStatusHistory(tt.args.ctx, serviceMock.db, tt.args.reqEnrollmentStatusHistories, tt.args.studentID)
			if tt.err != nil {
				assert.Equal(t, err, tt.err)
			} else {
				assert.Equal(t, hasActivatedEnrollmentStatus, tt.expectedResp)
			}
		})
	}
}

func TestDomainStudent_upsertEnrollmentStatusHistory(t *testing.T) {
	t.Parallel()
	serviceMock, student := DomainStudentServiceMock()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()

	type args struct {
		ctx             context.Context
		db              libdatabase.QueryExecer
		enrolmentStatus entity.DomainEnrollmentStatusHistory
		order           int
	}
	tests := []struct {
		name  string
		args  args
		setup func(serviceMock *prepareDomainStudentMock)
		err   error
	}{
		{
			name: "happy case: can update when current enrollment and last enrollment is diff",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					time.Now().Add(201*time.Hour),
					time.Time{},
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				//Mock all enrollment status fo student
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Once().Return(
					entity.DomainEnrollmentStatusHistories{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
						},
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
						},
					}, nil,
				)
				//Mock current enrollment status
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, true).Once().Return(
					entity.DomainEnrollmentStatusHistories{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
						},
					}, nil,
				)
				// Mock last enrollment status
				serviceMock.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
						},
					}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				serviceMock.enrollmentStatusHistoryRepo.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			},
			err: nil,
		},
		{
			name: "happy case: can update when current enrollment is empty and last enrollment have values",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					time.Now().Add(201*time.Hour),
					time.Time{},
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				//Mock all enrollment status fo student
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Once().Return(
					entity.DomainEnrollmentStatusHistories{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
						},
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
						},
					}, nil,
				)
				//Mock current enrollment status
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, true).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				// Mock last enrollment status
				serviceMock.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
						},
					}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				serviceMock.enrollmentStatusHistoryRepo.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			},
			err: nil,
		},
		{
			name: "happy case: can update end_date with STUDENT_ENROLLMENT_STATUS_TEMPORARY",
			args: args{
				ctx: ctx,
				enrolmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("location-id"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
					startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
					endDate:          field.NewTime(time.Now().Add(200 * time.Hour)),
				},
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				//Mock all enrollment status fo student
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Once().Return(
					entity.DomainEnrollmentStatusHistories{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewNullTime(),
						},
					}, nil,
				)
				//Mock current enrollment status
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, true).Once().Return(
					entity.DomainEnrollmentStatusHistories{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewNullTime(),
						},
					}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("location-id"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()),
							startDate:        field.NewTime(time.Now().Add(-100 * time.Hour)),
							endDate:          field.NewNullTime(),
						},
					}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			},
			err: nil,
		},
		{
			name: "happy case: create enrollment status history on LMS",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Now().Add(-100*time.Hour),
					time.Now().Add(200*time.Hour),
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			err: nil,
		},
		{
			name: "happy case: create student enrollment history on ERP",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Now().Add(-100*time.Hour),
					time.Now().Add(200*time.Hour),
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			err: nil,
		},
		{
			name: "happy case: update enrolment status history LMS",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					time.Now().Add(time.Hour),
					time.Now().Add(200*time.Hour),
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, true).Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			err: nil,
		},

		{
			name: "bad case: throw error when create enrollment with only StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
					time.Now().Add(-100*time.Hour),
					time.Now().Add(200*time.Hour),
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			err: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", 1),
			},
		},
		{
			name: "bad case: can not create student enrollment history",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Now().Add(-100*time.Hour),
					time.Now().Add(200*time.Hour),
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("create error"))
			},
			err: errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(errors.New("create error"), "service.EnrollmentStatusHistoryRepo.Create"),
			},
		},
		{
			name: "bad case: can not create student enrollment history on ERP with StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED",
			args: args{
				ctx: ctx,
				enrolmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					time.Now().Add(-100*time.Hour),
					time.Now().Add(200*time.Hour),
					"order-id",
					1),
				order: 1,
			},
			setup: func(sm *prepareDomainStudentMock) {
				serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			err: errors.Wrap(errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", 1),
			}, "service.validateEnrollmentStatusCreateRequestFromOrder"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			tt.setup(&serviceMock)
			err := student.upsertEnrollmentStatusHistory(tt.args.ctx, serviceMock.db, tt.args.enrolmentStatus)
			if err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
			}
		})
	}
}

func TestDomainStudent_getConfiguration(t *testing.T) {
	mockDomainStudent, service := DomainStudentServiceMock()
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *mpb.Configuration
		wantErr error
	}{
		{
			name: "Can not get config",
			args: args{ctx: ctx},
			setup: func() {
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", mock.Anything, mock.Anything).
					Once().
					Return(nil, fmt.Errorf("error"))
			},
			want: nil,
			wantErr: errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(fmt.Errorf("error"), "service.ConfigurationClient.GetConfigurationByKey"),
			},
		},
		{
			name: "Config not found",
			args: args{ctx: ctx},
			setup: func() {
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", mock.Anything, mock.Anything).
					Once().
					Return(&mpb.GetConfigurationByKeyResponse{}, nil)
			},
			want: nil,
			wantErr: errcode.Error{
				Code: errcode.InternalError,
				Err:  fmt.Errorf("not found config"),
			},
		},
		{
			name: "Get config but wrong token",
			args: args{ctx: ctx},
			setup: func() {
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", mock.Anything, mock.Anything).
					Once().
					Return(nil, fmt.Errorf("wrong token"))
			},
			want: &mpb.Configuration{
				ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
				ConfigValue: constant.ConfigValueOff,
			},
		},
		{
			name: "Get config success",
			args: args{ctx: ctx},
			setup: func() {
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", mock.Anything, mock.Anything).
					Once().
					Return(&mpb.GetConfigurationByKeyResponse{Configuration: &mpb.Configuration{}}, nil)
			},
			want: &mpb.Configuration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := service.getConfiguration(tt.args.ctx)
			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, wantErr.Code, e.Code)
			}
		})
	}
}

func TestDomainStudent_validateEnrollmentStatusHistories(t *testing.T) {
	mockDomainStudent, service := DomainStudentServiceMock()
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	domainStudent := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			UserID: field.NewString(idutil.ULIDNow()),
		},
	}
	internalConfig := repository.NewInternalConfiguration(entity.NullDomainConfiguration{})
	internalConfig.InternalConfigurationAttribute.ConfigValue = field.NewString(constant.ConfigValueOff)

	type args struct {
		ctx      context.Context
		students []aggregate.DomainStudent
	}
	tests := []struct {
		ctx     context.Context
		name    string
		args    args
		setup   func(ctx context.Context)
		wantErr error
	}{
		{
			name: "validate enrollment status histories for creating student failed",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory("student-id", "location-id",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
								time.Now().Add(-100*time.Hour),
								time.Now().Add(200*time.Hour),
								"order-id",
								1),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(entity.DomainEnrollmentStatusHistories{}, nil)
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", 0),
				Index:     0,
			},
		},
		{
			name: "validate enrollment status histories for creating student success",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory("student-id", "location-id",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
								time.Now().Add(-100*time.Hour),
								time.Now().Add(200*time.Hour),
								"order-id",
								1),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(entity.DomainEnrollmentStatusHistories{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "validate enrollment status histories for updating student failed",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory("student-id", "location-id",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
								time.Now().Add(-101*time.Hour),
								time.Now().Add(200*time.Hour),
								"order-id",
								1),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Twice().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, false).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(200*time.Hour),
							time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(2*time.Millisecond),
							"order-id",
							1,
						),
					}, nil,
				)
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.start_date", 0),
			},
		},
		{
			name: "validate enrollment status histories with skipping if nothing change",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory("student-id", "location-id",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
								time.Now().Add(-100*time.Hour),
								time.Now().Add(200*time.Hour),
								"order-id",
								1),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Twice().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, false).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(200*time.Hour),
							time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(2*time.Millisecond),
							"order-id",
							1,
						),
					}, nil,
				)
			},
			wantErr: nil,
		},
		{
			name: "validate enrollment status histories with status potential can change to any status",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory("student-id", "location-id",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
								time.Now().Add(-90*time.Hour),
								time.Now().Add(200*time.Hour),
								"order-id",
								1),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Twice().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, false).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(200*time.Hour),
							time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(2*time.Millisecond),
							"order-id",
							1,
						),
					}, nil,
				)
			},
			wantErr: nil,
		},
		{
			name: "validate enrollment status histories for updating student success",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory(
								"student-id",
								"Manabie",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
								time.Time{}, // Zero date
								time.Time{},
								"",
								0,
							),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)

				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("Manabie"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
						},
						createMockDomainEnrollmentStatusHistory(
							"student-id",
							"Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC),
							time.Time{},
							"",
							0,
						),
					}, nil,
				)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil,
				)
			},
			wantErr: nil,
		},
		{
			name: "validate enrollment status histories with other status can change to temporary",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory("student-id", "location-id",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
								time.Now().Add(-90*time.Hour),
								time.Now().Add(200*time.Hour),
								"order-id",
								1),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Twice().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, false).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(200*time.Hour),
							time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(2*time.Millisecond),
							"order-id",
							1,
						),
					}, nil,
				)
			},
			wantErr: nil,
		},
		{
			name: "happpy case: should skip validation if nothing change in req client from flow ERP",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory(
								"student-id",
								"Manabie",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
								time.Time{}, // Zero date
								time.Time{},
								"",
								0,
							),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)

				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("Manabie"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
						},
						createMockDomainEnrollmentStatusHistory(
							"student-id",
							"Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC),
							time.Time{},
							"",
							0,
						),
					}, nil,
				)

			},
			wantErr: nil,
		},
		{
			name: "happpy case: should skip validation if nothing change in req client from flow ERP when enable toggle IsFeatureEnabledOnOrganization",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory(
								"student-id",
								"Manabie",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
								time.Time{}, // Zero date
								time.Time{},
								"",
								0,
							),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(true, nil)
				mockDomainStudent.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)

				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						&MockDomainEnrollmentStatusHistory{
							userID:           field.NewString("student-id"),
							locationID:       field.NewString("Manabie"),
							enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
						},
						createMockDomainEnrollmentStatusHistory(
							"student-id",
							"Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC),
							time.Time{},
							"",
							0,
						),
					}, nil,
				)

			},
			wantErr: nil,
		},
		{
			name: "unhapppy case: non-ERP status can not be changed to others at flow ERP",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: domainStudent,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							createMockDomainEnrollmentStatusHistory(
								"student-id",
								"Manabie",
								upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
								time.Time{}, // Zero date
								time.Time{},
								"",
								0,
							),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDomainStudent.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				mockDomainStudent.configurationClient.
					On("GetConfigurationByKey", ctx, mock.Anything).
					Once().Return(
					&mpb.GetConfigurationByKeyResponse{
						Configuration: &mpb.Configuration{
							ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
							ConfigValue: constant.ConfigValueOff,
						},
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", ctx, mock.Anything, mock.Anything, mock.Anything).Twice().Return([]entity.DomainEnrollmentStatusHistory{}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
							time.Now().Add(-100*time.Hour),
							time.Time{},
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
							time.Now().Add(-100*time.Hour),
							time.Time{},
							"order-id",
							1),
					}, nil)

				// mock current enrollmentStatusHistory
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything, true).
					Once().Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
							time.Now().Add(-100*time.Hour),
							time.Time{},
							"order-id",
							1),
					}, nil)
				mockDomainStudent.enrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
							time.Now().Add(-100*time.Hour),
							time.Time{},
							"order-id",
							1),
					}, nil,
				)
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories.enrollment_status", 0),
				Index:     0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			tt.ctx = interceptors.ContextWithUserID(ctx, idutil.ULIDNow())
			tt.ctx = interceptors.ContextWithJWTClaims(tt.ctx, claim)
			tt.setup(tt.ctx)

			err := service.validateEnrollmentStatusHistories(tt.ctx, tt.args.students...)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestDomainStudent_upsertEnrollmentStatusHistories(t *testing.T) {
	serviceMock, service := DomainStudentServiceMock()
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	t.Run("happy case: should skip update when updating without enrollment status histories", func(t *testing.T) {
		t.Parallel()
		domainStudent := aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					UserID: field.NewString("user-id"),
				},
			},
			EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
		}

		err := service.upsertEnrollmentStatusHistories(ctx, serviceMock.db, domainStudent)
		assert.Nil(t, err)
	})
	t.Run("happy case: should run create logic", func(t *testing.T) {
		t.Parallel()
		domainStudent := aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					UserID: field.NewString("user-id"),
				},
			},
			EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
				mock_usermgmt.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
						UserID:           field.NewString("user-id"),
						LocationID:       field.NewString("location-id"),
						EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						StartDate:        field.NewTime(time.Now()),
					},
				},
			},
		}

		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
			entity.DomainEnrollmentStatusHistories{}, nil,
		)

		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", mock.Anything, mock.Anything, mock.Anything).Once().Return(
			[]entity.DomainEnrollmentStatusHistory{}, nil,
		)

		serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

		err := service.upsertEnrollmentStatusHistories(ctx, serviceMock.db, domainStudent)
		assert.Nil(t, err)
	})

}

func TestDomainStudent_validateEnrollmentStatusHistoriesBeforeCreating(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []aggregate.DomainStudent
		wantErr error
	}{
		{
			name: "should not throw error when enrollment status history is empty and user id is not empty",
			args: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID: field.NewString("user-id"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
				},
			},
			wantErr: nil,
		},
		{
			name: "should throw error when enrollment status history has an invalid status",
			args: []aggregate.DomainStudent{
				// valid student
				{
					DomainStudent: &mock_usermgmt.Student{RandomStudent: mock_usermgmt.RandomStudent{UserID: field.NewString(idutil.ULIDNow())}},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
								StartDate:        field.NewTime(time.Now().AddDate(0, 0, 1)),
							},
						},
					},
				},
				// invalid student
				{
					// valid student profile
					DomainStudent: &mock_usermgmt.Student{RandomStudent: mock_usermgmt.RandomStudent{UserID: field.NewString(idutil.ULIDNow())}},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						// valid enrollment status history
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
								StartDate:        field.NewTime(time.Now().AddDate(0, 0, 1)),
							},
						},
						// invalid enrollment status history, need status
						mock_usermgmt.EnrollmentStatusHistory{},
					},
				},
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].enrollment_status", 0, 1),
				Index:     1,
			},
		},
		{
			name: "should not throw error when enrollment status history is empty and user access path is not empty",
			args: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
					UserAccessPaths: entity.DomainUserAccessPaths{
						mock_usermgmt.UserAccessPath{
							RandomUserAccessPath: mock_usermgmt.RandomUserAccessPath{
								LocationID: field.NewString("location-id"),
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should throw error when enrollment status histories and location are empty",
			args: []aggregate.DomainStudent{
				{
					DomainStudent:             &mock_usermgmt.Student{},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
					UserAccessPaths:           entity.DomainUserAccessPaths{},
				},
			},
			wantErr: errcode.Error{
				Code:      errcode.MissingMandatory,
				FieldName: fmt.Sprintf("students[%d].locations", 0),
				Index:     0,
			},
		},
		{
			name: "should throw error when enrollment status is Potential/Temporary/Non-Potential and start date is after current date with correct error-index",
			args: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
								StartDate:        field.NewTime(time.Now().AddDate(0, 0, 1)),
							},
						},
					},
				},
				{
					DomainStudent: &mock_usermgmt.Student{},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
								StartDate:        field.NewTime(time.Now().AddDate(0, 0, 1)),
							},
						},
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
								StartDate:        field.NewTime(time.Now().AddDate(0, 0, 1)),
							},
						},
					},
				},
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprint("students[0].enrollment_status_histories[1].start_date"),
				Index:     1,
			},
		},
		{
			name: "should not throw error when enrollment status is not Potential/Temporary/Non-Potential and start date is after current date",
			args: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
								StartDate:        field.NewTime(time.Now().AddDate(0, 0, 1)),
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should throw error when enrollment status start date is after or equal end date with correct error-index",
			args: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
								StartDate:        field.NewTime(time.Now()),
								EndDate:          field.NewTime(time.Now().AddDate(0, 0, -1)),
							},
						},
					},
				},
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprint("students[0].enrollment_status_histories[0].end_date"),
				Index:     0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnrollmentStatusHistoriesBeforeCreating(context.Background(), tt.args...)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestDeactivateAndReactivateStudentManager_DeactivateAndReactivateStudents(t *testing.T) {
	type DeactivateAndReactivateStudentManagerMock struct {
		enrollmentStatusHistoryRepo *mock_repositories.MockDomainEnrollmentStatusHistoryRepo
		userRepo                    *mock_repositories.MockDomainUserRepo
	}
	type Arg struct {
		studentIDs []string
	}
	tests := []struct {
		name          string
		arg           Arg
		expectedErr   error
		setupWithMock func(*DeactivateAndReactivateStudentManagerMock)
	}{
		{
			name: "should run correctly",
			arg: Arg{
				studentIDs: []string{"student-id-1", "student-id-2"},
			},
			expectedErr: nil,
			setupWithMock: func(managerMock *DeactivateAndReactivateStudentManagerMock) {
				managerMock.enrollmentStatusHistoryRepo.On("GetInactiveAndActiveStudents", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2"}, []string{"deactivated-enrollment-status"}).
					Once().Return([]entity.DomainEnrollmentStatusHistory{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:    field.NewString("student-id-1"),
							StartDate: field.NewTime(time.Now()),
						},
					},
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:    field.NewString("student-id-2"),
							StartDate: field.NewNullTime(),
						},
					},
				}, nil)
				managerMock.userRepo.On("UpdateActivation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "should return error when GetInactiveAndActiveStudents query failed",
			arg: Arg{
				studentIDs: []string{"student-id-1", "student-id-2"},
			},
			expectedErr: errcode.Error{
				Code: errcode.InternalError,
			},
			setupWithMock: func(managerMock *DeactivateAndReactivateStudentManagerMock) {
				managerMock.enrollmentStatusHistoryRepo.On("GetInactiveAndActiveStudents", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2"}, []string{"deactivated-enrollment-status"}).
					Once().Return([]entity.DomainEnrollmentStatusHistory{}, errcode.Error{
					Code: errcode.InternalError,
				})
			},
		},
		{
			name: "should return error when UpdateActivation call failed",
			arg: Arg{
				studentIDs: []string{"student-id-1"},
			},
			expectedErr: errcode.Error{
				Code: errcode.InternalError,
			},
			setupWithMock: func(managerMock *DeactivateAndReactivateStudentManagerMock) {
				managerMock.enrollmentStatusHistoryRepo.On("GetInactiveAndActiveStudents", mock.Anything, mock.Anything, []string{"student-id-1"}, []string{"deactivated-enrollment-status"}).
					Once().Return([]entity.DomainEnrollmentStatusHistory{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:    field.NewString("student-id-1"),
							StartDate: field.NewTime(time.Now()),
						},
					},
				}, nil)
				managerMock.userRepo.On("UpdateActivation", mock.Anything, mock.Anything, mock.Anything).Return(errcode.Error{
					Code: errcode.InternalError,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := DeactivateAndReactivateStudentManagerMock{
				enrollmentStatusHistoryRepo: &mock_repositories.MockDomainEnrollmentStatusHistoryRepo{},
				userRepo:                    &mock_repositories.MockDomainUserRepo{},
			}
			manager := StudentActivationStatusManager{
				EnrollmentStatusHistoryRepo: mockManager.enrollmentStatusHistoryRepo,
				UserRepo:                    mockManager.userRepo,
			}
			tt.setupWithMock(&mockManager)
			err := manager.DeactivateAndReactivateStudents(context.Background(), &mock_database.Ext{}, tt.arg.studentIDs, []string{"deactivated-enrollment-status"})
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestHandelOrderFlowEnrollmentStatus_HandleOrderFlowForTheExistedLocations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domainEnrollmentStatusHistoryRepo := new(mock_repositories.MockDomainEnrollmentStatusHistoryRepo)
	domainUserAccessPathRepo := new(mock_repositories.MockDomainUserAccessPathRepo)
	db := new(mock_database.Ext)
	now := time.Now()

	type fields struct {
		SyncEnrollmentStatusHistory     func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog, enrollmentStatus string) error
		DeactivateAndReactivateStudents func(ctx context.Context, db libdatabase.Ext, studentIDs []string) error
	}
	type args struct {
		req *OrderEventLog
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    bool
		wantErr error
	}{
		{
			name: "happy case with checking end date of the current enrollment status",
			fields: fields{
				SyncEnrollmentStatusHistory: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog, enrollmentStatus string) error {
					// if the EndDate of the current enrollment status is earlier than req.EndDate
					// the req.EndDate is adjusted to match the EndDate of the current enrollment status.
					if !req.EndDate.Equal(now.Add(200 * time.Hour)) {
						return errors.New("assert time is not correct")
					}

					if enrollmentStatus != constant.StudentEnrollmentStatusPotential {
						return errors.New("assert enrollment status is not correct")
					}
					return nil
				},
			},
			args: args{
				req: &OrderEventLog{
					OrderStatus:      ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:        ppb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
					StudentID:        "student-id",
					LocationID:       "Manabie",
					EnrollmentStatus: constant.StudentEnrollmentStatusEnrolled,
					StartDate:        now,
					EndDate:          now.Add(300 * time.Hour),
				},
			},
			want: false,
			setup: func() {
				domainEnrollmentStatusHistoryRepo.
					On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							constant.StudentEnrollmentStatusTemporary,
							now.Add(-40*time.Hour),
							now.Add(200*time.Hour),
							"order-id",
							1,
						),
					}, nil)

				domainEnrollmentStatusHistoryRepo.
					//                                                                                       End date = start date - 1sec
					On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, now.Add(-1*time.Second)).
					Once().
					Return(nil)
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HandelOrderFlowEnrollmentStatus{
				DomainEnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
				DomainUserAccessPathRepo:          domainUserAccessPathRepo,
				SyncEnrollmentStatusHistory:       tt.fields.SyncEnrollmentStatusHistory,
				DeactivateAndReactivateStudents:   tt.fields.DeactivateAndReactivateStudents,
			}
			tt.setup()
			got, err := s.HandleExistedLocations(ctx, db, tt.args.req)
			assert.Equalf(t, tt.want, got, "HandleExistedLocations(%v, %v, %v)", ctx, db, tt.args.req)
			assert.Equalf(t, tt.wantErr, err, "HandleExistedLocations(%v, %v, %v)", ctx, db, tt.args.req)
		})
	}
}

func TestHandelOrderFlowEnrollmentStatus_HandleEnrollmentStatusUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domainEnrollmentStatusHistoryRepo := new(mock_repositories.MockDomainEnrollmentStatusHistoryRepo)
	domainUserAccessPathRepo := new(mock_repositories.MockDomainUserAccessPathRepo)
	db := new(mock_database.Ext)
	now := time.Now()

	type fields struct {
		SyncEnrollmentStatusHistory     func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog, enrollmentStatus string) error
		DeactivateAndReactivateStudents func(ctx context.Context, db libdatabase.Ext, studentIDs []string) error
	}
	type args struct {
		req *OrderEventLog
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    bool
		wantErr error
	}{
		{
			name: "happy case with end date = start date - 1sec",
			fields: fields{
				SyncEnrollmentStatusHistory: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog, enrollmentStatus string) error {
					if enrollmentStatus != constant.StudentEnrollmentStatusWithdrawn {
						return errors.New("assert enrollment status is not correct")
					}
					return nil
				},
			},
			args: args{
				req: &OrderEventLog{
					OrderStatus:      ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:        ppb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
					StudentID:        "student-id",
					LocationID:       "Manabie",
					EnrollmentStatus: constant.StudentEnrollmentStatusWithdrawn,
					StartDate:        now,
					EndDate:          now.Add(300 * time.Hour),
				},
			},
			want: false,
			setup: func() {
				domainEnrollmentStatusHistoryRepo.
					On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							constant.StudentEnrollmentStatusGraduated,
							now.Add(-40*time.Hour),
							now.Add(200*time.Hour),
							"order-id",
							1,
						),
					}, nil)

				domainEnrollmentStatusHistoryRepo.
					//                                                                                       End date = start date - 1sec
					On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, now.Add(-1*time.Second)).
					Once().
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "skip sync when submitted enrollment status is already existed",
			fields: fields{},
			args: args{
				req: &OrderEventLog{
					OrderStatus:      ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:        ppb.OrderType_ORDER_TYPE_GRADUATE.String(),
					StudentID:        "student-id",
					LocationID:       "Manabie",
					EnrollmentStatus: constant.StudentEnrollmentStatusWithdrawn,
					StartDate:        now,
					EndDate:          now.Add(300 * time.Hour),
				},
			},
			want: false,
			setup: func() {
				domainEnrollmentStatusHistoryRepo.
					On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							constant.StudentEnrollmentStatusGraduated,
							now.Add(-40*time.Hour),
							now.Add(200*time.Hour),
							"order-id",
							1,
						),
					}, nil)
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HandelOrderFlowEnrollmentStatus{
				DomainEnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
				DomainUserAccessPathRepo:          domainUserAccessPathRepo,
				SyncEnrollmentStatusHistory:       tt.fields.SyncEnrollmentStatusHistory,
				DeactivateAndReactivateStudents:   tt.fields.DeactivateAndReactivateStudents,
			}
			tt.setup()
			got, err := s.HandleEnrollmentStatusUpdate(ctx, db, tt.args.req)
			assert.Equalf(t, tt.want, got, "HandleExistedLocations(%v, %v, %v)", ctx, db, tt.args.req)
			assert.Equalf(t, tt.wantErr, err, "HandleExistedLocations(%v, %v, %v)", ctx, db, tt.args.req)
		})
	}
}

func TestHandelOrderFlowEnrollmentStatus_HandleOrderFlowForVoidEnrollmentStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domainEnrollmentStatusHistoryRepo := new(mock_repositories.MockDomainEnrollmentStatusHistoryRepo)
	domainUserAccessPathRepo := new(mock_repositories.MockDomainUserAccessPathRepo)
	db := new(mock_database.Ext)
	now := time.Now()

	type fields struct {
		SyncEnrollmentStatusHistory     func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog, enrollmentStatus string) error
		DeactivateAndReactivateStudents func(ctx context.Context, db libdatabase.Ext, studentIDs []string) error
	}
	type args struct {
		req *OrderEventLog
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    bool
		wantErr error
	}{
		{
			name: "happy case existed 1 latest enrollment statuses",
			fields: fields{
				DeactivateAndReactivateStudents: func(ctx context.Context, db libdatabase.Ext, studentIDs []string) error {
					return nil
				},
			},
			args: args{
				req: &OrderEventLog{
					OrderStatus:      ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:        ppb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
					OrderID:          "order-id",
					StudentID:        "student-id",
					LocationID:       "Manabie",
					EnrollmentStatus: constant.StudentEnrollmentStatusWithdrawn,
					StartDate:        now,
					EndDate:          now.Add(300 * time.Hour),
				},
			},
			want: false,
			setup: func() {
				domainEnrollmentStatusHistoryRepo.
					On("GetLatestEnrollmentStudentOfLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return([]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
							constant.StudentEnrollmentStatusEnrolled,
							now.Add(-40*time.Hour),
							now.Add(200*time.Hour),
							"order-id",
							1,
						),
					}, nil)

				domainEnrollmentStatusHistoryRepo.
					On("SoftDeleteEnrollments", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(nil)

				domainUserAccessPathRepo.
					On("SoftDeleteByUserIDAndLocationIDs", mock.Anything, mock.Anything, "student-id", mock.Anything, []string{"Manabie"}).
					Once().Return(nil)

			},
			wantErr: nil,
		},
		{
			name: "happy case existed 2 latest enrollment statuses",
			fields: fields{
				DeactivateAndReactivateStudents: func(ctx context.Context, db libdatabase.Ext, studentIDs []string) error {
					return nil
				},
			},
			args: args{
				req: &OrderEventLog{
					OrderStatus:      ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:        ppb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
					OrderID:          "order-id",
					StudentID:        "student-id",
					LocationID:       "Manabie",
					EnrollmentStatus: constant.StudentEnrollmentStatusWithdrawn,
					StartDate:        now,
					EndDate:          now.Add(300 * time.Hour),
				},
			},
			want: false,
			setup: func() {
				domainEnrollmentStatusHistoryRepo.
					On("GetLatestEnrollmentStudentOfLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return([]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie1",
							constant.StudentEnrollmentStatusEnrolled,
							now.Add(-40*time.Hour),
							now.Add(200*time.Hour),
							"order-id",
							1,
						),
						createMockDomainEnrollmentStatusHistory("student-id", "Manabie2",
							constant.StudentEnrollmentStatusEnrolled,
							now.Add(-40*time.Hour),
							now.Add(200*time.Hour),
							"order-id",
							1,
						),
					}, nil)

				domainEnrollmentStatusHistoryRepo.
					On("SoftDeleteEnrollments", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(nil)

				domainUserAccessPathRepo.
					On("SoftDeleteByUserIDAndLocationIDs", mock.Anything, mock.Anything, "student-id", mock.Anything, []string{"Manabie1"}).
					Once().Return(nil)

				domainEnrollmentStatusHistoryRepo.
					On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(nil)

			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HandelOrderFlowEnrollmentStatus{
				DomainEnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
				DomainUserAccessPathRepo:          domainUserAccessPathRepo,
				SyncEnrollmentStatusHistory:       tt.fields.SyncEnrollmentStatusHistory,
				DeactivateAndReactivateStudents:   tt.fields.DeactivateAndReactivateStudents,
			}
			tt.setup()
			got, err := s.HandleVoidEnrollmentStatus(ctx, db, tt.args.req)
			assert.Equalf(t, tt.want, got, "HandleExistedLocations(%v, %v, %v)", ctx, db, tt.args.req)
			assert.Equalf(t, tt.wantErr, err, "HandleExistedLocations(%v, %v, %v)", ctx, db, tt.args.req)
		})
	}
}

func TestDomainStudent_updateEnrollmentStatusHistories(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	t.Run("happy case: should skip update when updating without enrollment status histories", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{}, nil)
		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Nil(t, err)
	})
	t.Run("happy case: should skip update when there is empty student to update", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{}

		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Nil(t, err)
	})

	t.Run("happy case: skip if there is a the same record enrollment status histories in BD", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
							StartDate:        field.NewTime(time.Now()),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)
		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Nil(t, err)
	})
	t.Run("happy case: update end_date of status temporary successfully", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusTemporary),
							StartDate:        field.NewTime(time.Now()),
							EndDate:          field.NewTime(time.Now().Add(36 * time.Hour)),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusTemporary),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)
		serviceMock.enrollmentStatusHistoryRepo.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Nil(t, err)
	})
	t.Run("unhappy case: update end_date of status temporary failed", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusTemporary),
							StartDate:        field.NewTime(time.Now()),
							EndDate:          field.NewTime(time.Now().Add(36 * time.Hour)),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusTemporary),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)
		serviceMock.enrollmentStatusHistoryRepo.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("error"))
		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Equal(t, errors.New("error").Error(), err.Error())
	})

	t.Run("happy case: deactivate enrollment status and create new one successfully", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
							StartDate:        field.NewTime(time.Now().Add(36 * time.Hour)),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)

		serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Nil(t, err)
	})

	t.Run("unhappy case: deactivate enrollment status failed", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
							StartDate:        field.NewTime(time.Now().Add(36 * time.Hour)),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)

		serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("error"))
		serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Equal(t, errors.New("error").Error(), err.Error())
	})

	t.Run("unhappy case: deactivate enrollment status and create failed", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
							StartDate:        field.NewTime(time.Now().Add(36 * time.Hour)),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)

		serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		serviceMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("error"))

		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Equal(t, errors.New("error").Error(), err.Error())
	})

	t.Run("happy case: update last enrollment status histories successfully", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusWithdrawn),
							StartDate:        field.NewTime(time.Now().Add(36 * time.Hour)),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
					StartDate:        field.NewTime(time.Now().Add(24 * time.Hour)),
				},
			},
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)

		serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		serviceMock.enrollmentStatusHistoryRepo.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Nil(t, err)
	})

	t.Run("happy case: update last enrollment status histories failed", func(t *testing.T) {
		serviceMock, service := DomainStudentServiceMock()
		t.Parallel()
		domainStudents := aggregate.DomainStudents{
			{
				DomainStudent: &mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("user-id"),
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id"),
							LocationID:       field.NewString("location-id"),
							EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusWithdrawn),
							StartDate:        field.NewTime(time.Now().Add(36 * time.Hour)),
						},
					},
				},
			},
		}
		serviceMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
					StartDate:        field.NewTime(time.Now().Add(24 * time.Hour)),
				},
			},
			mock_usermgmt.EnrollmentStatusHistory{
				RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
					UserID:           field.NewString("user-id"),
					LocationID:       field.NewString("location-id"),
					EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
					StartDate:        field.NewTime(time.Now()),
				},
			},
		}, nil)

		serviceMock.enrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		serviceMock.enrollmentStatusHistoryRepo.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("error"))

		err := service.updateEnrollmentStatusHistories(ctx, serviceMock.db, domainStudents)
		assert.Equal(t, errors.New("error").Error(), err.Error())
	})
}

func TestModifyModifyStartDateEnrollmentStatusHistory(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domainEnrollmentStatusHistoryRepo := new(mock_repositories.MockDomainEnrollmentStatusHistoryRepo)
	db := new(mock_database.Ext)
	now := time.Now().Add(1 * time.Hour)

	type args struct {
		enrollmentStatusHistory entity.DomainEnrollmentStatusHistory
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    entity.DomainEnrollmentStatusHistory
		wantErr error
	}{
		{
			name: "happy case: enrollment status history with start date in the past",
			args: args{
				enrollmentStatusHistory: &repository.EnrollmentStatusHistory{
					EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
						StartDate: field.NewTime(now.Add(-24 * time.Hour)),
						OrderID:   field.NewString("order-id"),
					},
				},
			},
			want: &repository.EnrollmentStatusHistory{
				EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
					StartDate: field.NewTime(now.Add(-24 * time.Hour)),
					OrderID:   field.NewString("order-id"),
				},
			},
			wantErr: nil,
		},
		{
			name: "bad case: enrollment status history same start date and same order id",
			args: args{
				enrollmentStatusHistory: &repository.EnrollmentStatusHistory{
					EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
						StartDate: field.NewTime(now),
						OrderID:   field.NewString("order-id"),
					},
				},
			},
			setup: func() {
				domainEnrollmentStatusHistoryRepo.On("GetSameStartDateEnrollmentStatusHistory", mock.Anything, mock.Anything, mock.Anything).
					Twice().
					Return(entity.DomainEnrollmentStatusHistories{
						&repository.EnrollmentStatusHistory{
							EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
								StartDate: field.NewTime(now),
								OrderID:   field.NewString("order-id"),
							},
						},
					}, nil)
			},
			want: nil,
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.FieldEnrollmentStatusHistoryOrderID),
				EntityName: entity.Entity(entity.EnrollmentStatusHistories),
				Index:      0,
			},
		},
		{
			name: "happy case: we gonna modify start date with 1 micro second to avoid pk constraint",
			args: args{
				enrollmentStatusHistory: &repository.EnrollmentStatusHistory{
					EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
						StartDate: field.NewTime(now),
						OrderID:   field.NewString("order-id"),
					},
				},
			},
			setup: func() {
				domainEnrollmentStatusHistoryRepo.On("GetSameStartDateEnrollmentStatusHistory", mock.Anything, mock.Anything, mock.Anything).
					Twice().
					Return(entity.DomainEnrollmentStatusHistories{
						&repository.EnrollmentStatusHistory{
							EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
								StartDate: field.NewTime(now),
								OrderID:   field.NewString("order-id-1"),
							},
						},
					}, nil)

			},
			want: &repository.EnrollmentStatusHistory{
				EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
					StartDate: field.NewTime(now.Add(1 * time.Microsecond)),
					OrderID:   field.NewString("order-id"),
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: we gonna modify start date with 1 micro second to avoid pk constraint by the max time",
			args: args{
				enrollmentStatusHistory: &repository.EnrollmentStatusHistory{
					EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
						StartDate: field.NewTime(now),
						OrderID:   field.NewString("order-id"),
					},
				},
			},
			setup: func() {
				domainEnrollmentStatusHistoryRepo.On("GetSameStartDateEnrollmentStatusHistory", mock.Anything, mock.Anything, mock.Anything).
					Twice().
					Return(entity.DomainEnrollmentStatusHistories{
						&repository.EnrollmentStatusHistory{
							EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
								StartDate: field.NewTime(now.Add(1 * time.Microsecond)),
								OrderID:   field.NewString("order-id-1"),
							},
						},
						&repository.EnrollmentStatusHistory{
							EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
								StartDate: field.NewTime(now.Add(2 * time.Microsecond)),
								OrderID:   field.NewString("order-id-2"),
							},
						},
					}, nil)

			},
			want: &repository.EnrollmentStatusHistory{
				EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
					StartDate: field.NewTime(now.Add(2*time.Microsecond + 1*time.Microsecond)),
					OrderID:   field.NewString("order-id"),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StudentRegistrationService{
				DB:                                db,
				DomainEnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
			}

			if tt.setup != nil {
				tt.setup()
			}

			enrollmentStatusHistory, err := EnrollmentStatusHistoryStartDateModifier(ctx, db, domainEnrollmentStatusHistoryRepo, tt.args.enrollmentStatusHistory)
			if utils.IsFutureDate(tt.args.enrollmentStatusHistory.StartDate()) {
				sameStartDateEnrollmentStatusHistory, _ := s.DomainEnrollmentStatusHistoryRepo.GetSameStartDateEnrollmentStatusHistory(ctx, s.DB, tt.args.enrollmentStatusHistory)
				assertPrimaryKeyEnrollmentStatusHistories(t, append(sameStartDateEnrollmentStatusHistory, enrollmentStatusHistory))
			}
			assert.Equal(t, tt.wantErr, err)
			if tt.want != nil || enrollmentStatusHistory != nil {
				assert.Equal(t, tt.want.CreatedAt(), enrollmentStatusHistory.CreatedAt())
				assert.Equal(t, tt.want.EndDate(), enrollmentStatusHistory.EndDate())
				assert.Equal(t, tt.want.EnrollmentStatus(), enrollmentStatusHistory.EnrollmentStatus())
				assert.Equal(t, tt.want.OrderID(), enrollmentStatusHistory.OrderID())
				assert.Equal(t, tt.want.OrderSequenceNumber(), enrollmentStatusHistory.OrderSequenceNumber())
				assert.Equal(t, tt.want.OrganizationID(), enrollmentStatusHistory.OrganizationID())
				assert.Equal(t, tt.want.StartDate(), enrollmentStatusHistory.StartDate())
				assert.Equal(t, tt.want.UserID(), enrollmentStatusHistory.UserID())
				assert.Equal(t, tt.want.LocationID(), enrollmentStatusHistory.LocationID())
			}
		})
	}
}

// assertPrimaryKeyEnrollmentStatusHistories assert "pk__student_enrollment_status_history" PRIMARY KEY, btree (student_id, location_id, enrollment_status, start_date)
// with given enrollmentStatusHistories
func assertPrimaryKeyEnrollmentStatusHistories(t *testing.T, enrollmentStatusHistories entity.DomainEnrollmentStatusHistories) {
	mapPk := make(map[string]struct{})
	for _, enrollmentStatusHistory := range enrollmentStatusHistories {
		if enrollmentStatusHistory == nil {
			continue
		}
		pk := fmt.Sprintf("%s-%s-%s-%s", enrollmentStatusHistory.UserID().String(), enrollmentStatusHistory.LocationID().String(), enrollmentStatusHistory.EnrollmentStatus().String(), enrollmentStatusHistory.StartDate().Time().Format(time.RFC3339Nano))
		if _, ok := mapPk[pk]; ok {
			t.Error("duplicate primary key!")
		}
		mapPk[pk] = struct{}{}
	}
}
