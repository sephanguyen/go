package services

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/entryexitmgmt/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParentNotif(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDb := new(mock_database.Ext)
	mockStudentQRRepo := new(mock_repositories.MockStudentQRRepo)
	mockStudentEntryExitRecordsRepo := new(mock_repositories.MockStudentEntryExitRecordsRepo)
	mockStudentRepo := new(mock_repositories.MockStudentRepo)
	mockStudentParentRepo := new(mock_repositories.MockStudentParentRepo)
	mockJsm := new(mock_nats.JetStreamManagement)

	s := &EntryExitModifierService{
		DB:                          mockDb,
		StudentQRRepo:               mockStudentQRRepo,
		StudentEntryExitRecordsRepo: mockStudentEntryExitRecordsRepo,
		StudentRepo:                 mockStudentRepo,
		StudentParentRepo:           mockStudentParentRepo,
		JSM:                         mockJsm,
	}
	user1 := &entities.User{
		FullName: database.Text("Test User"),
		Country:  database.Text("COUNTRY_VN"),
	}
	student1 := &entities.Student{
		ID:       database.Text("student-id-1"),
		SchoolID: database.Int4(1),
		User:     *user1,
	}
	const (
		errorJSM     = "s.JSM.PublishContext error: publish error"
		publishError = "publish error"
	)
	testcases := []TestCase{
		{
			name: "happy case entry record",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    student1,
				TouchEvent: eepb.TouchEvent_TOUCH_ENTRY,
				Title:      "Entry & Exit Activity",
				Message:    "Student 1 entered the center at 2022-01-02 08:00",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "happy case exit record",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    student1,
				TouchEvent: eepb.TouchEvent_TOUCH_EXIT,
				Title:      "Entry & Exit Activity",
				Message:    "Student 1 exited the center at 2022-01-02 15:00",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "failed to publish parents notification entry time internal error",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    student1,
				TouchEvent: eepb.TouchEvent_TOUCH_ENTRY,
				Title:      "Entry & Exit Activity",
				Message:    "Student 1 exited the center at 2022-01-02 16:00",
			},
			expectedErr: status.Error(codes.Internal, errorJSM),
			setup: func(ctx context.Context) {
				// attempt four times
				for i := 0; i < 4; i++ {
					mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, errors.New(publishError))
				}
			},
		},
		{
			name: "failed to publish parents notification exit time internal error",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    student1,
				TouchEvent: eepb.TouchEvent_TOUCH_EXIT,
				Title:      "Entry & Exit Activity",
				Message:    "Student 1 exited the center at 2022-01-02 17:00",
			},
			expectedErr: status.Error(codes.Internal, errorJSM),
			setup: func(ctx context.Context) {
				// attempt four times
				for i := 0; i < 4; i++ {
					mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, errors.New(publishError))
				}
			},
		},
		{
			name: "failed to publish parents notification entry time no student id",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    &entities.Student{},
				TouchEvent: eepb.TouchEvent_TOUCH_ENTRY,
			},
			expectedErr: status.Error(codes.Internal, errorJSM),
			setup: func(ctx context.Context) {
				// attempt four times
				for i := 0; i < 4; i++ {
					mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, errors.New(publishError))
				}
			},
		},
		{
			name: "failed to publish parents notification exit time no student id",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    &entities.Student{},
				TouchEvent: eepb.TouchEvent_TOUCH_EXIT,
			},
			expectedErr: status.Error(codes.Internal, errorJSM),
			setup: func(ctx context.Context) {
				// attempt four times
				for i := 0; i < 4; i++ {
					mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, errors.New(publishError))
				}
			},
		},
		{
			name: "failed to publish parents notification exit time empty student id",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    &entities.Student{ID: pgtype.Text{}},
				TouchEvent: eepb.TouchEvent_TOUCH_EXIT,
			},
			expectedErr: status.Error(codes.Internal, errorJSM),
			setup: func(ctx context.Context) {
				// attempt four times
				for i := 0; i < 4; i++ {
					mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, errors.New(publishError))
				}
			},
		},
		{
			name: "failed to publish parents notification entry time empty student id",
			ctx:  ctx,
			req: &EntryExitNotifyDetails{
				Student:    &entities.Student{ID: pgtype.Text{}},
				TouchEvent: eepb.TouchEvent_TOUCH_ENTRY,
			},
			expectedErr: status.Error(codes.Internal, errorJSM),
			setup: func(ctx context.Context) {
				// attempt four times
				for i := 0; i < 4; i++ {
					mockJsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, errors.New(publishError))
				}
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.Notify(testCase.ctx, testCase.req.(*EntryExitNotifyDetails))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			}
			mock.AssertExpectationsForObjects(t, mockDb, mockStudentParentRepo, mockJsm)
		})
	}
}
func Test_generateEntryExitNotifyDetailsMessages(t *testing.T) {
	user1 := &entities.User{
		FullName: database.Text("Test User"),
		Country:  database.Text("COUNTRY_VN"),
	}
	student1 := &entities.Student{
		ID:       database.Text("student-id-1"),
		SchoolID: database.Int4(1),
		User:     *user1,
	}
	user2 := &entities.User{
		FullName: database.Text("Test User2"),
		Country:  database.Text("COUNTRY_JP"),
	}
	student2 := &entities.Student{
		ID:       database.Text("student-id-2"),
		SchoolID: database.Int4(1),
		User:     *user2,
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	touchEntryTime := time.Date(2021, 2, 24, 1, 10, 0, 0, loc)
	touchExitTime := time.Date(2021, 2, 24, 8, 10, 0, 0, loc)

	expectedTouchEntryTime := fmt.Sprintf("%d/%02d/%02d %02d:%02d", 2021, 2, 24, 1, 10)
	expectedTouchExitTime := fmt.Sprintf("%d/%02d/%02d %02d:%02d", 2021, 2, 24, 8, 10)

	type args struct {
		notifyDetails *EntryExitNotifyDetails
	}
	tests := []struct {
		name                 string
		args                 args
		expectedNotifMessage string
		expectedNotifTitle   string
	}{
		{
			name: "Generate Notification Message for Student Exit When Scan",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student1,
					TouchEvent: eepb.TouchEvent_TOUCH_EXIT,
					RecordType: eepb.RecordType_QR_CODE_SCAN,
					TouchTime:  touchExitTime,
				},
			},
			expectedNotifMessage: "Test User exited the center at " + expectedTouchExitTime,
			expectedNotifTitle:   "Entry & Exit Activity",
		},
		{
			name: "Generate Notification Message for Student Entry When Scan",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student1,
					TouchEvent: eepb.TouchEvent_TOUCH_ENTRY,
					RecordType: eepb.RecordType_QR_CODE_SCAN,
					TouchTime:  touchEntryTime,
				},
			},
			expectedNotifMessage: "Test User entered the center at " + expectedTouchEntryTime,
			expectedNotifTitle:   "Entry & Exit Activity",
		},
		{
			name: "Generate Notification Message for JP Student Exit When Scan",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student2,
					TouchEvent: eepb.TouchEvent_TOUCH_EXIT,
					RecordType: eepb.RecordType_QR_CODE_SCAN,
					TouchTime:  touchExitTime,
				},
			},
			expectedNotifMessage: expectedTouchExitTime + "にTest User2が教室から退室しました",
			expectedNotifTitle:   "入退室記録",
		},
		{
			name: "Generate Notification Message for JP Student Entry When Scan",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student2,
					TouchEvent: eepb.TouchEvent_TOUCH_ENTRY,
					RecordType: eepb.RecordType_QR_CODE_SCAN,
					TouchTime:  touchEntryTime,
				},
			},
			expectedNotifMessage: expectedTouchEntryTime + "にTest User2が教室に入室しました",
			expectedNotifTitle:   "入退室記録",
		},
		{
			name: "Generate Notification Message for Student Entry When Create",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student1,
					RecordType: eepb.RecordType_CREATE_MANUAL,
				},
			},
			expectedNotifMessage: "There are new records for entry & exit of Test User. Please review them on your kid history",
			expectedNotifTitle:   "Entry & Exit Activity",
		},
		{
			name: "Generate Notification Message for JP Student Entry When Create",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student2,
					RecordType: eepb.RecordType_CREATE_MANUAL,
				},
			},
			expectedNotifMessage: "Test User2の新しい入退室記録が追加されました。入退室記録から確認してください",
			expectedNotifTitle:   "入退室記録",
		},
		{
			name: "Generate Notification Message for Student Entry When Create",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student1,
					RecordType: eepb.RecordType_UPDATE_MANUAL,
				},
			},
			expectedNotifMessage: "There are updated records for entry & exit of Test User. Please review them on your kid history",
			expectedNotifTitle:   "Entry & Exit Activity",
		},
		{
			name: "Generate Notification Message for JP Student Entry When Create",
			args: args{
				notifyDetails: &EntryExitNotifyDetails{
					Student:    student2,
					RecordType: eepb.RecordType_UPDATE_MANUAL,
				},
			},
			expectedNotifMessage: "Test User2の新しい入退室記録が更新されました。入退室記録から確認してください",
			expectedNotifTitle:   "入退室記録",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualNotifyDetails := generateEntryExitNotifyTitleMessage(tt.args.notifyDetails)
			if !reflect.DeepEqual(actualNotifyDetails.Message, tt.expectedNotifMessage) {
				t.Errorf("generateEntryExitNotifyDetailsMessage() = %v, want %v", actualNotifyDetails.Message, tt.expectedNotifMessage)
			}
			if !reflect.DeepEqual(actualNotifyDetails.Title, tt.expectedNotifTitle) {
				t.Errorf("generateEntryExitNotifyDetailsMessage() = %v, want %v", actualNotifyDetails.Title, tt.expectedNotifTitle)
			}
		})
	}
}
