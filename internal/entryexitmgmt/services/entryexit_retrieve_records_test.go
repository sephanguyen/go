package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	mock_repositories "github.com/manabie-com/backend/mock/entryexitmgmt/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEntryExitModifierService_RetrieveEntryExitRecords(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockStudentEntryExitRecordsRepo := new(mock_repositories.MockStudentEntryExitRecordsRepo)

	s := &EntryExitModifierService{
		DB:                          mockDB,
		StudentEntryExitRecordsRepo: mockStudentEntryExitRecordsRepo,
	}
	studentID := pgtype.Text{}
	_ = studentID.Set("test-id")

	fakeStudentID := pgtype.Text{}
	_ = studentID.Set("fake-id")

	entryExitRecordsFilterAll := []*entities.StudentEntryExitRecords{}
	entryExitRecordsFilterLastMonth := []*entities.StudentEntryExitRecords{}
	entryExitRecordsFilterThisMonth := []*entities.StudentEntryExitRecords{}
	entryExitRecordsFilterThisYear := []*entities.StudentEntryExitRecords{}
	emptyEntryExitRecord := []*entities.StudentEntryExitRecords{}

	// create data for this year
	for i := 0; i <= 19; i++ {
		entryExitRecordsFilterAll = append(entryExitRecordsFilterAll, &entities.StudentEntryExitRecords{
			StudentID: studentID,
			ID:        pgtype.Int4{Int: int32(i + 1)},
			EntryAt:   pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
			ExitAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Hour).AddDate(-1, 0, 0)},
		})
	}
	responseFilterAll := make([]*eepb.EntryExitRecord, 0, len(entryExitRecordsFilterLastMonth))
	for _, ee := range entryExitRecordsFilterAll {
		responseFilterAll = append(responseFilterAll, &eepb.EntryExitRecord{
			EntryexitId: ee.ID.Int,
			EntryAt:     timestamppb.New(ee.EntryAt.Time),
			ExitAt:      timestamppb.New(ee.ExitAt.Time),
		})
	}

	// create data for last month
	for i := 0; i <= 5; i++ {
		entryExitRecordsFilterLastMonth = append(entryExitRecordsFilterLastMonth, &entities.StudentEntryExitRecords{
			StudentID: studentID,
			ID:        pgtype.Int4{Int: int32(i + 1)},
			EntryAt:   pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0)},
			ExitAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Hour).AddDate(0, -1, 0)},
		})
	}
	responseFilterLastMonth := make([]*eepb.EntryExitRecord, 0, len(entryExitRecordsFilterLastMonth))
	for _, ee := range entryExitRecordsFilterLastMonth {
		responseFilterLastMonth = append(responseFilterLastMonth, &eepb.EntryExitRecord{
			EntryexitId: ee.ID.Int,
			EntryAt:     timestamppb.New(ee.EntryAt.Time),
			ExitAt:      timestamppb.New(ee.ExitAt.Time),
		})
	}
	// create data for this month
	for i := 0; i <= 9; i++ {
		entryExitRecordsFilterThisMonth = append(entryExitRecordsFilterThisMonth, &entities.StudentEntryExitRecords{
			StudentID: studentID,
			ID:        pgtype.Int4{Int: int32(i + 1)},
			EntryAt:   pgtype.Timestamptz{Time: time.Now()},
			ExitAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Hour)},
		})
	}
	responseFilterThisMonth := make([]*eepb.EntryExitRecord, 0, len(entryExitRecordsFilterThisMonth))
	for _, ee := range entryExitRecordsFilterThisMonth {
		responseFilterThisMonth = append(responseFilterThisMonth, &eepb.EntryExitRecord{
			EntryexitId: ee.ID.Int,
			EntryAt:     timestamppb.New(ee.EntryAt.Time),
			ExitAt:      timestamppb.New(ee.ExitAt.Time),
		})
	}
	// create data for this year
	entryExitRecordsFilterThisYear = append(entryExitRecordsFilterLastMonth, entryExitRecordsFilterThisMonth...)
	responseFilterThisYear := make([]*eepb.EntryExitRecord, 0, len(entryExitRecordsFilterThisMonth))
	for _, ee := range entryExitRecordsFilterThisYear {
		responseFilterThisYear = append(responseFilterThisYear, &eepb.EntryExitRecord{
			EntryexitId: ee.ID.Int,
			EntryAt:     timestamppb.New(ee.EntryAt.Time),
			ExitAt:      timestamppb.New(ee.ExitAt.Time),
		})
	}
	testcases := []TestCase{
		{
			name:        "no record for other student id",
			ctx:         ctx,
			expectedErr: nil,
			req: &eepb.RetrieveEntryExitRecordsRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				RecordFilter: eepb.RecordFilter_THIS_MONTH,
				StudentId:    fakeStudentID.String,
			},
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockStudentEntryExitRecordsRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything).Once().Return(emptyEntryExitRecord, nil)
			},
		},
		{
			name:        "happy case for retrieving entry and exit record for this month",
			ctx:         ctx,
			expectedErr: nil,
			req: &eepb.RetrieveEntryExitRecordsRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				RecordFilter: eepb.RecordFilter_THIS_MONTH,
				StudentId:    studentID.String,
			},
			expectedResp: &eepb.RetrieveEntryExitRecordsResponse{
				EntryExitRecords: responseFilterThisMonth,
				NextPage: &cpb.Paging{
					Limit: uint32(10),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(10) + int64(0),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentEntryExitRecordsRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything).Once().Return(entryExitRecordsFilterThisMonth, nil)
			},
		},
		{
			name:        "happy case for retrieving entry and exit record for last month",
			ctx:         ctx,
			expectedErr: nil,
			req: &eepb.RetrieveEntryExitRecordsRequest{
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				RecordFilter: eepb.RecordFilter_LAST_MONTH,
				StudentId:    studentID.String,
			},
			expectedResp: &eepb.RetrieveEntryExitRecordsResponse{
				EntryExitRecords: responseFilterLastMonth,
				NextPage: &cpb.Paging{
					Limit: uint32(10),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(10) + int64(0),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentEntryExitRecordsRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything).Once().Return(entryExitRecordsFilterLastMonth, nil)
			},
		},
		{
			name:        "happy case for retrieving entry and exit record for this year",
			ctx:         ctx,
			expectedErr: nil,
			req: &eepb.RetrieveEntryExitRecordsRequest{
				Paging: &cpb.Paging{
					Limit: 20,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				RecordFilter: eepb.RecordFilter_THIS_YEAR,
				StudentId:    studentID.String,
			},
			expectedResp: &eepb.RetrieveEntryExitRecordsResponse{
				EntryExitRecords: responseFilterThisYear,
				NextPage: &cpb.Paging{
					Limit: uint32(20),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(20) + int64(0),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentEntryExitRecordsRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything).Once().Return(entryExitRecordsFilterThisYear, nil)
			},
		},
		{
			name:        "happy case for retrieving all entry and exit records",
			ctx:         ctx,
			expectedErr: nil,
			req: &eepb.RetrieveEntryExitRecordsRequest{
				Paging: &cpb.Paging{
					Limit: 20,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				RecordFilter: eepb.RecordFilter_ALL,
				StudentId:    studentID.String,
			},
			expectedResp: &eepb.RetrieveEntryExitRecordsResponse{
				EntryExitRecords: responseFilterAll,
				NextPage: &cpb.Paging{
					Limit: uint32(20),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(20) + int64(0),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentEntryExitRecordsRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything).Once().Return(entryExitRecordsFilterAll, nil)
			},
		},
		{
			name: "failed to retrieve entry exit record",
			ctx:  ctx,
			req: &eepb.RetrieveEntryExitRecordsRequest{
				Paging: &cpb.Paging{
					Limit: 20,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				RecordFilter: eepb.RecordFilter_ALL,
				StudentId:    studentID.String,
			},
			expectedErr: status.Error(codes.Internal, "closed pool"),
			setup: func(ctx context.Context) {
				mockStudentEntryExitRecordsRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
		},
		{
			name: "failed to get parent ids err no rows",
			ctx:  ctx,
			req: &eepb.RetrieveEntryExitRecordsRequest{
				Paging: &cpb.Paging{
					Limit: 20,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				RecordFilter: eepb.RecordFilter_ALL,
				StudentId:    studentID.String,
			},
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockStudentEntryExitRecordsRepo.On("RetrieveRecordsByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.RetrieveEntryExitRecords(testCase.ctx, testCase.req.(*eepb.RetrieveEntryExitRecordsRequest))
			if err != nil {
				fmt.Println(err)
			}
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			}
			if testCase.expectedResp != nil {
				assert.Equal(t, len(testCase.expectedResp.(*eepb.RetrieveEntryExitRecordsResponse).EntryExitRecords), len(resp.EntryExitRecords))
				assert.Equal(t, testCase.expectedResp.(*eepb.RetrieveEntryExitRecordsResponse).NextPage.Limit, resp.NextPage.Limit)
				assert.Equal(t, testCase.expectedResp.(*eepb.RetrieveEntryExitRecordsResponse).NextPage.Offset, resp.NextPage.Offset)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockStudentEntryExitRecordsRepo)
		})
	}
}
