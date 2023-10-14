package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInfoNotificationRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	notiID := "noti-id-1"
	testCases := []struct {
		Name  string
		Ent   *entities.InfoNotification
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &entities.InfoNotification{NotificationID: database.Text(notiID)},
			SetUp: func(ctx context.Context) {
				e := &entities.InfoNotification{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, ctx)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &InfoNotificationRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			id, err := repo.Upsert(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
				assert.Equal(t, notiID, id)
			}
		})
	}
}

func TestInfoNotificationRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	entInDB1 := &entities.InfoNotification{}
	entInDB2 := &entities.InfoNotification{}
	database.AllRandomEntity(entInDB1)
	database.AllRandomEntity(entInDB2)

	type TestCase struct {
		Name    string
		NotiIDs []string
		Filter  *FindNotificationFilter
		Err     error
		SetUp   func(ctx context.Context, this *TestCase)
	}

	testCases := []TestCase{
		{
			Name:    "happy case without paging",
			NotiIDs: []string{entInDB1.NotificationID.String, entInDB2.NotificationID.String},
			Filter:  NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotiIDs = database.TextArray([]string{entInDB1.NotificationID.String, entInDB2.NotificationID.String})
				this.Filter.Status = database.TextArray([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()})
				this.Filter.ResourcePath = database.Text("tenant_id")

				fields, vals1 := entInDB1.FieldMap()
				_, vals2 := entInDB2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, this.Filter.NotiIDs, this.Filter.ResourcePath, this.Filter.Status, this.Filter.Type)
			},
		},
		{
			Name:    "happy case with paging",
			NotiIDs: []string{entInDB1.NotificationID.String, entInDB2.NotificationID.String},
			Filter:  NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotiIDs = database.TextArray([]string{entInDB1.NotificationID.String, entInDB2.NotificationID.String})
				this.Filter.Status = database.TextArray([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()})
				this.Filter.ResourcePath = database.Text("tenant_id")
				this.Filter.Limit = database.Int8(2)
				this.Filter.Offset = database.Int8(0)

				fields, vals1 := entInDB1.FieldMap()
				_, vals2 := entInDB2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, this.Filter.NotiIDs, this.Filter.ResourcePath, this.Filter.Status, this.Filter.Type, this.Filter.Limit, this.Filter.Offset)
			},
		},
		{
			Name:    "more filters",
			NotiIDs: []string{entInDB1.NotificationID.String, entInDB2.NotificationID.String},
			Filter:  NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotificationMsgIDs = database.TextArray([]string{"noti-msg-id"})
				this.Filter.NotiIDs = database.TextArray([]string{entInDB1.NotificationID.String, entInDB2.NotificationID.String})
				this.Filter.Status = database.TextArray([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()})
				this.Filter.ResourcePath = database.Text("tenant_id")
				this.Filter.FromScheduled.Set(time.Now())
				this.Filter.ToScheduled.Set(time.Now())
				this.Filter.FromSent.Set(time.Now())
				this.Filter.ToSent.Set(time.Now())
				this.Filter.EditorIDs.Set([]string{"editor-id"})

				fields, vals1 := entInDB1.FieldMap()
				_, vals2 := entInDB2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					this.Filter.NotiIDs, this.Filter.NotificationMsgIDs,
					this.Filter.FromScheduled, this.Filter.ToScheduled,
					this.Filter.FromSent, this.Filter.ToSent,
					this.Filter.ResourcePath, this.Filter.EditorIDs, this.Filter.Status, this.Filter.Type)
			},
		},
		{
			Name:    "filters with From",
			NotiIDs: []string{entInDB1.NotificationID.String, entInDB2.NotificationID.String},
			Filter:  NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotificationMsgIDs = database.TextArray([]string{"noti-msg-id"})
				this.Filter.NotiIDs = database.TextArray([]string{entInDB1.NotificationID.String, entInDB2.NotificationID.String})
				this.Filter.Status = database.TextArray([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()})
				this.Filter.ResourcePath = database.Text("tenant_id")
				this.Filter.FromScheduled.Set(time.Now())
				this.Filter.FromSent.Set(time.Now())

				fields, vals1 := entInDB1.FieldMap()
				_, vals2 := entInDB2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					this.Filter.NotiIDs, this.Filter.NotificationMsgIDs,
					this.Filter.FromScheduled,
					this.Filter.FromSent,
					this.Filter.ResourcePath, this.Filter.Status, this.Filter.Type)
			},
		},
		{
			Name:    "filters with To",
			NotiIDs: []string{entInDB1.NotificationID.String, entInDB2.NotificationID.String},
			Filter:  NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotificationMsgIDs = database.TextArray([]string{"noti-msg-id"})
				this.Filter.NotiIDs = database.TextArray([]string{entInDB1.NotificationID.String, entInDB2.NotificationID.String})
				this.Filter.Status = database.TextArray([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()})
				this.Filter.ResourcePath = database.Text("tenant_id")
				this.Filter.ToScheduled.Set(time.Now())
				this.Filter.ToSent.Set(time.Now())

				fields, vals1 := entInDB1.FieldMap()
				_, vals2 := entInDB2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					this.Filter.NotiIDs, this.Filter.NotificationMsgIDs,
					this.Filter.ToScheduled,
					this.Filter.ToSent,
					this.Filter.ResourcePath, this.Filter.Status, this.Filter.Type)
			},
		},
	}

	repo := &InfoNotificationRepo{}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx, &testCase)
			res, err := repo.Find(ctx, db, testCase.Filter)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
			assert.Nil(t, err)
			assert.Equal(t, entInDB1, res[0])
			assert.Equal(t, entInDB2, res[1])
		})
	}
}

func TestInfoNotificationRepo_CountTotalNotificationForStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	res1 := &TotalNotificationForStatus{
		Status: database.Text(idutil.ULIDNow()),
		Total:  database.Int8(10),
	}
	res2 := &TotalNotificationForStatus{
		Status: database.Text(idutil.ULIDNow()),
		Total:  database.Int8(20),
	}

	type TestCase struct {
		Name   string
		Filter *FindNotificationFilter
		Err    error
		SetUp  func(ctx context.Context, this *TestCase)
	}

	testCases := []TestCase{
		{
			Name:   "happy case",
			Filter: NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				mockDB.MockScanArray(nil, []string{"status", "total"}, [][]interface{}{{&res1.Status, &res1.Total}, {&res2.Status, &res2.Total}})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, this.Filter.NotiIDs, this.Filter.Status, this.Filter.Type)
			},
		},
		{
			Name:   "more filters",
			Filter: NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotificationMsgIDs = database.TextArray([]string{"noti-msg-id"})
				this.Filter.FromSent.Set(time.Now())
				this.Filter.ToSent.Set(time.Now())

				mockDB.MockScanArray(nil, []string{"status", "total"}, [][]interface{}{{&res1.Status, &res1.Total}, {&res2.Status, &res2.Total}})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					this.Filter.NotiIDs,
					this.Filter.NotificationMsgIDs,
					this.Filter.FromSent,
					this.Filter.ToSent,
					this.Filter.Status,
					this.Filter.Type)
			},
		},
		{
			Name:   "more filters 2",
			Filter: NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotificationMsgIDs = database.TextArray([]string{"noti-msg-id"})
				this.Filter.FromSent.Set(time.Now())

				mockDB.MockScanArray(nil, []string{"status", "total"}, [][]interface{}{{&res1.Status, &res1.Total}, {&res2.Status, &res2.Total}})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					this.Filter.NotiIDs,
					this.Filter.NotificationMsgIDs,
					this.Filter.FromSent,
					this.Filter.Status,
					this.Filter.Type)
			},
		},
		{
			Name:   "more filters 3",
			Filter: NewFindNotificationFilter(),
			SetUp: func(ctx context.Context, this *TestCase) {
				this.Filter.NotificationMsgIDs = database.TextArray([]string{"noti-msg-id"})
				this.Filter.ToSent.Set(time.Now())

				mockDB.MockScanArray(nil, []string{"status", "total"}, [][]interface{}{{&res1.Status, &res1.Total}, {&res2.Status, &res2.Total}})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					this.Filter.NotiIDs,
					this.Filter.NotificationMsgIDs,
					this.Filter.ToSent,
					this.Filter.Status,
					this.Filter.Type)
			},
		},
	}

	repo := &InfoNotificationRepo{}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx, &testCase)
			res, err := repo.CountTotalNotificationForStatus(ctx, db, testCase.Filter)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
			assert.Nil(t, err)
			assert.Equal(t, uint32(res1.Total.Int), res[res1.Status.String])
			assert.Equal(t, uint32(res2.Total.Int), res[res2.Status.String])
			assert.Equal(t, uint32(res1.Total.Int+res2.Total.Int), res[cpb.NotificationStatus_NOTIFICATION_STATUS_NONE.String()])
		})
	}
}

func TestInfoNotificationMsgRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *entities.InfoNotificationMsg
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &entities.InfoNotificationMsg{},
			SetUp: func(ctx context.Context) {
				e := &entities.InfoNotification{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, ctx)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &InfoNotificationMsgRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.Upsert(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestInfoNotificationMsgRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	testCases := []struct {
		Name       string
		NotiMsgIDs []string
		Err        error
		SetUp      func(ctx context.Context)
	}{
		{
			Name:       "happy case",
			NotiMsgIDs: []string{"noti-id-1", "noti-id-2"},
			SetUp: func(ctx context.Context) {
				e := &entities.InfoNotificationMsg{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				fieldsRows := make([]pgproto3.FieldDescription, len(fields))
				for i := range fields {
					fieldsRows[i] = pgproto3.FieldDescription{Name: []byte(fields[i])}
				}
				rows.On("FieldDescriptions").Return(fieldsRows)
				db.On("Query", ctx, mock.AnythingOfType("string"), mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", values...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
			},
		},
	}

	repo := &InfoNotificationMsgRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			_, err := repo.GetByIDs(ctx, db, database.TextArray(testCase.NotiMsgIDs))
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestInfoNotificationMsgRepo_GetIDsByTitle(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &InfoNotificationMsgRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	title := "title"
	notiMsgIDsTextArrayReq := []pgtype.Text{database.Text("notification_msg_id-1"), database.Text("notification_msg_id-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.Text(title))
		mockDB.MockScanArray(nil, []string{"notification_msg_id"}, [][]interface{}{{&notiMsgIDsTextArrayReq[0]}, {&notiMsgIDsTextArrayReq[1]}})

		_, err := r.GetIDsByTitle(ctx, db, database.Text(title))
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(title))
		mockDB.MockScanArray(nil, []string{"notification_msg_id"}, [][]interface{}{{&notiMsgIDsTextArrayReq[0]}, {&notiMsgIDsTextArrayReq[1]}})

		notiMsgIDs, err := r.GetIDsByTitle(ctx, db, database.Text(title))
		assert.NoError(t, err)
		for i, notiMsgID := range notiMsgIDs {
			assert.Equal(t, notiMsgID, notiMsgIDsTextArrayReq[i].String)
		}
	})
}

func TestUsersInfoNotificationRepo_CountByStatus(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	row := &mock_database.Row{}
	testCases := []struct {
		Name   string
		UserID string
		Status string
		Err    error
		SetUp  func(ctx context.Context)
	}{
		{
			Name:   "happy case",
			UserID: "user-id-1",
			Status: cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String(),
			SetUp: func(ctx context.Context) {
				db.On(
					"QueryRow",
					ctx,
					mock.AnythingOfType("string"),
					database.Text("user-id-1"),
					database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()),
				).Once().Return(row, nil)
				row.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			_, _, err := repo.CountByStatus(ctx, db, database.Text(testCase.UserID), database.Text(testCase.Status))
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestUsersInfoNotificationRepo_UpdateUnreadUser(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	cmd := pgconn.CommandTag([]byte(`1`))
	testCases := []struct {
		Name           string
		NotificationID string
		UserID         []string
		Err            error
		SetUp          func(ctx context.Context)
	}{
		{
			Name:           "happy case",
			NotificationID: "notification-id-1",
			UserID:         []string{"user-id-1"},
			SetUp: func(ctx context.Context) {
				db.On(
					"Exec",
					ctx,
					mock.AnythingOfType("string"),
					database.Text("notification-id-1"),
					database.TextArray([]string{"user-id-1"}),
					database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()),
				).Once().Return(cmd, nil)
			},
		},
	}

	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.UpdateUnreadUser(ctx, db, database.Text(testCase.NotificationID), database.TextArray(testCase.UserID))
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestInfoNotificationRepo_IsNotificationDeleted(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name           string
		NotificationID string
		UserID         []string
		Err            error
		SetUp          func(ctx context.Context)
	}{
		{
			Name:           "happy case",
			NotificationID: "notification-id-1",
			UserID:         []string{"user-id-1"},
			SetUp: func(ctx context.Context) {
				row := &mock_database.Row{}
				db.On(
					"QueryRow",
					ctx,
					mock.AnythingOfType("string"),
					database.Text("notification-id-1"),
				).Once().Return(row, nil)
				row.On("Scan", mock.Anything).Once().Return(nil)
			},
		},
	}

	repo := &InfoNotificationRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			_, err := repo.IsNotificationDeleted(ctx, db, database.Text(testCase.NotificationID))
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestInfoNotificationRepo_UpdateNotification(t *testing.T) {
	t.Parallel()
	db := testutil.NewMockDB()
	type args struct {
		ctx            context.Context
		db             *testutil.MockDB
		notificationId pgtype.Text
		attributes     map[string]interface{}
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(bd *testutil.MockDB)
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name: "no update case",
			args: args{
				ctx:            context.Background(),
				db:             db,
				notificationId: pgtype.Text{String: idutil.ULIDNow()},
				attributes:     make(map[string]interface{}),
			},
			wantErr:  false,
			mockFunc: func(mockDB *testutil.MockDB) {},
		},
		{
			name: "update case",
			args: args{
				ctx:            context.Background(),
				db:             db,
				notificationId: pgtype.Text{String: idutil.ULIDNow()},
				attributes: map[string]interface{}{
					"status":  "NOTIFICATION_STATUS_SENT",
					"sent_at": time.Now(),
				},
			},
			wantErr: false,
			mockFunc: func(mockDB *testutil.MockDB) {
				params := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, params...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InfoNotificationRepo{}
			tt.mockFunc(tt.args.db)
			if e := r.UpdateNotification(tt.args.ctx, tt.args.db.DB, tt.args.notificationId, tt.args.attributes); tt.wantErr && e != nil {
				t.Errorf("want error but have no")
			}
		})
	}
}

func TestFindNotificationFilter_Validate(t *testing.T) {
	type fields struct {
		NotiIDs pgtype.TextArray
		Status  pgtype.TextArray
		From    pgtype.Timestamptz
		To      pgtype.Timestamptz
		Owner   pgtype.Text
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{
			name: "Error case",
			fields: fields{
				NotiIDs: database.TextArray(nil),
				Status:  database.TextArray(nil),
				From:    pgtype.Timestamptz{},
				To:      pgtype.Timestamptz{},
				Owner:   pgtype.Text{},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if err == nil {
					return false
				}
				if err.Error() != "FindNotificationFilter all field is null" {
					return false
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FindNotificationFilter{
				NotiIDs:       tt.fields.NotiIDs,
				Status:        tt.fields.Status,
				FromScheduled: tt.fields.From,
				ToScheduled:   tt.fields.To,
				ResourcePath:  tt.fields.Owner,
			}
			tt.wantErr(t, f.Validate(), fmt.Sprintf("Validate()"))
		})
	}
}

func TestInfoNotificationRepo_SetStatus(t *testing.T) {
	t.Parallel()
	mockDatabase := testutil.NewMockDB()
	type args struct {
		ctx            context.Context
		db             *testutil.MockDB
		notificationID pgtype.Text
		status         pgtype.Text
	}
	tests := []struct {
		name     string
		args     args
		wantErr  assert.ErrorAssertionFunc
		mockFunc func(mockDB *testutil.MockDB, notificationID, status pgtype.Text)
	}{
		// TODO: Add test cases.
		{
			name: "no row effect",
			args: args{
				ctx:            context.Background(),
				db:             mockDatabase,
				notificationID: database.Text("notificationId"),
				status:         database.Text(cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			mockFunc: func(mockDB *testutil.MockDB, notificationID, status pgtype.Text) {
				params := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, status, notificationID)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, params...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InfoNotificationRepo{}
			tt.mockFunc(tt.args.db, tt.args.notificationID, tt.args.status)
			tt.wantErr(t, r.SetStatus(tt.args.ctx, tt.args.db.DB, tt.args.notificationID, tt.args.status), fmt.Sprintf("SetStatus(%v, %v, %v, %v)", tt.args.ctx, tt.args.db, tt.args.notificationID, tt.args.status))
		})
	}
}

func TestInfoNotificationRepo_SetSentAt(t *testing.T) {
	t.Parallel()
	mockDatabase := testutil.NewMockDB()
	type args struct {
		ctx            context.Context
		db             *testutil.MockDB
		notificationID pgtype.Text
	}
	tests := []struct {
		name     string
		args     args
		wantErr  assert.ErrorAssertionFunc
		mockFunc func(mockDB *testutil.MockDB, notificationID pgtype.Text)
	}{
		// TODO: Add test cases.
		{
			name: "no row effect",
			args: args{
				ctx:            context.Background(),
				db:             mockDatabase,
				notificationID: database.Text("notificationId"),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			mockFunc: func(mockDB *testutil.MockDB, notificationID pgtype.Text) {
				params := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, notificationID)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, params...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InfoNotificationRepo{}
			tt.mockFunc(tt.args.db, tt.args.notificationID)
			tt.wantErr(t, r.SetSentAt(tt.args.ctx, tt.args.db.DB, tt.args.notificationID), fmt.Sprintf("SetStatus(%v, %v, %v)", tt.args.ctx, tt.args.db, tt.args.notificationID))
		})
	}
}

func TestInfoNotificationRepo_DiscardNotification(t *testing.T) {
	t.Parallel()
	mockDatabase := testutil.NewMockDB()
	type args struct {
		ctx            context.Context
		db             *testutil.MockDB
		notificationID pgtype.Text
		statues        pgtype.TextArray
	}
	tests := []struct {
		name     string
		args     args
		wantErr  assert.ErrorAssertionFunc
		mockFunc func(mockDB *testutil.MockDB, notificationID pgtype.Text)
	}{
		// TODO: Add test cases.
		{
			name: "1 row effect",
			args: args{
				ctx:            context.Background(),
				db:             mockDatabase,
				notificationID: database.Text("notificationId"),
				statues:        database.TextArray([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()}),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			mockFunc: func(mockDB *testutil.MockDB, notificationID pgtype.Text) {
				params := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, notificationID, mock.Anything)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, params...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InfoNotificationRepo{}
			tt.mockFunc(tt.args.db, tt.args.notificationID)
			tt.wantErr(t, r.DiscardNotification(tt.args.ctx, tt.args.db.DB, tt.args.notificationID, tt.args.statues), fmt.Sprintf("DiscardNotification(%v, %v, %v, %v)", tt.args.ctx, tt.args.db, tt.args.notificationID, tt.args.statues))
		})
	}
}

func TestInfoNotificationMsgRepo_GetByNotificationIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	notiIDs := []string{"noti-1", "noti-2"}
	infoNotificationMsg := &entities.InfoNotificationMsg{}
	database.AllRandomEntity(infoNotificationMsg)
	testCases := []struct {
		Name    string
		NotiIDs pgtype.TextArray
		Err     error
		Setup   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			NotiIDs: database.TextArray(notiIDs),
			Err:     nil,
			Setup: func(ctx context.Context) {
				f := []string{"notiID"}
				fields, vals := infoNotificationMsg.FieldMap()
				for _, field := range fields {
					f = append(f, field)
				}
				notiID := &pgtype.Text{String: "noti-1"}
				v := []interface{}{notiID}
				for _, val := range vals {
					v = append(v, val)
				}
				mockDB.MockQueryArgs(t, nil, ctx, mock.AnythingOfType("string"), database.TextArray(notiIDs))
				mockDB.MockScanArray(nil, f, [][]interface{}{v})
			},
		},
	}

	repo := &InfoNotificationMsgRepo{}
	ctx := context.Background()

	for _, testcase := range testCases {
		t.Run(testcase.Name, func(t *testing.T) {
			testcase.Setup(ctx)
			_, err := repo.GetByNotificationIDs(ctx, mockDB.DB, testcase.NotiIDs)
			if testcase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testcase.Err.Error(), err.Error())
			}
		})
	}
}

func TestInfoNotificationMsgRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	type Req struct {
		NotificationMsgIDs []string
	}
	testCases := []struct {
		Name        string
		Req         interface{}
		ExpectedErr error
		SetUp       func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: &Req{
				NotificationMsgIDs: []string{"1"},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				db.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &InfoNotificationMsgRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.SoftDelete(ctx, db, testCase.Req.(*Req).NotificationMsgIDs)
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestUsersInfoNotificationRepo_Upsert(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	userInfoNotification1 := &entities.UserInfoNotification{}
	userInfoNotification2 := &entities.UserInfoNotification{}
	database.AllRandomEntity(userInfoNotification1)
	database.AllRandomEntity(userInfoNotification2)
	records := []*entities.UserInfoNotification{
		userInfoNotification1,
		userInfoNotification2,
	}
	testCases := []struct {
		Name  string
		Data  []*entities.UserInfoNotification
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Data: records,
			Err:  nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", ctx, mock.Anything).Once().Return(batchResults)
				for range records {
					cmdTag := pgconn.CommandTag([]byte(`1`))
					batchResults.On("Exec").Once().Return(cmdTag, nil)
				}
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()

	for _, testcase := range testCases {
		t.Run(testcase.Name, func(t *testing.T) {
			testcase.Setup(ctx)
			err := repo.Upsert(ctx, mockDB.DB, testcase.Data)
			if testcase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testcase.Err.Error(), err.Error())
			}
		})
	}
}

func TestUsersInfoNotificationRepo_Find(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	userInfoNotification1 := &entities.UserInfoNotification{}
	userInfoNotification2 := &entities.UserInfoNotification{}
	database.AllRandomEntity(userInfoNotification1)
	database.AllRandomEntity(userInfoNotification2)
	filter := NewFindUserNotificationFilter()
	testCases := []struct {
		Name    string
		Filter  FindUserNotificationFilter
		ExpcErr error
		Setup   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			Filter:  filter,
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				fields, values1 := userInfoNotification1.FieldMap()
				_, values2 := userInfoNotification2.FieldMap()
				mockDB.MockQueryArgs(t, nil, ctx, mock.Anything,
					filter.UserNotificationIDs,
					filter.UserIDs,
					filter.NotiIDs,
					filter.UserStatus,
					filter.OffsetTime,
					filter.OffsetText,
					filter.StudentID,
					filter.ParentID,
					filter.IsImportant,
					filter.Limit)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values1, values2})
			},
		},
		{
			Name:    "error scan",
			Filter:  filter,
			ExpcErr: fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows),
			Setup: func(ctx context.Context) {
				fields, values1 := userInfoNotification1.FieldMap()
				_, values2 := userInfoNotification2.FieldMap()
				mockDB.MockQueryArgs(t, nil, ctx, mock.Anything,
					filter.UserNotificationIDs,
					filter.UserIDs,
					filter.NotiIDs,
					filter.UserStatus,
					filter.OffsetTime,
					filter.OffsetText,
					filter.StudentID,
					filter.ParentID,
					filter.IsImportant,
					filter.Limit)
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values1, values2})
			},
		},
	}

	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			res, err := repo.Find(ctx, mockDB.DB, testCase.Filter)
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
				assert.Equal(t, userInfoNotification1, res[0])
				assert.Equal(t, userInfoNotification2, res[1])
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}

func TestUsersInfoNotificationRepo_SetStatus(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()
	userID := "user-id"
	userInfoNotificationIDs := []string{"user-ifn-id-1", "user-ifn-id-1"}
	status := "status"
	testCases := []struct {
		Name                    string
		ExpecErr                error
		UserID                  pgtype.Text
		UserInfoNotificationIDs pgtype.TextArray
		Status                  pgtype.Text
		Setup                   func(ctx context.Context)
	}{
		{
			Name:                    "happy case",
			ExpecErr:                nil,
			UserID:                  database.Text(userID),
			UserInfoNotificationIDs: database.TextArray(userInfoNotificationIDs),
			Status:                  database.Text(status),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.MockExecArgs(t, cmdTag, nil, ctx, mock.Anything, database.Text(status), database.TextArray(userInfoNotificationIDs), database.Text(userID))
			},
		},
		{
			Name:                    "no rows affect",
			ExpecErr:                fmt.Errorf("no rows affected"),
			UserID:                  database.Text(userID),
			UserInfoNotificationIDs: database.TextArray(userInfoNotificationIDs),
			Status:                  database.Text(status),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				mockDB.MockExecArgs(t, cmdTag, nil, ctx, mock.Anything, database.Text(status), database.TextArray(userInfoNotificationIDs), database.Text(userID))
			},
		},
		{
			Name:                    "err",
			ExpecErr:                puddle.ErrClosedPool,
			UserID:                  database.Text(userID),
			UserInfoNotificationIDs: database.TextArray(userInfoNotificationIDs),
			Status:                  database.Text(status),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.MockExecArgs(t, cmdTag, puddle.ErrClosedPool, ctx, mock.Anything, database.Text(status), database.TextArray(userInfoNotificationIDs), database.Text(userID))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx)
			err := repo.SetStatus(ctx, mockDB.DB, tc.UserID, tc.UserInfoNotificationIDs, tc.Status)
			if tc.ExpecErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.ExpecErr.Error(), err.Error())
			}
		})
	}
}

func TestUsersInfoNotificationRepo_SetStatusByNotificationIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()
	userID := "user-id"
	notificationIDs := []string{"noti-id-1", "noti-id-2"}
	status := "status"
	testCases := []struct {
		Name            string
		Setup           func(ctx context.Context)
		ExpecErr        error
		UserID          pgtype.Text
		NotificationIDs pgtype.TextArray
		Status          pgtype.Text
	}{
		{
			Name:            "happy case",
			ExpecErr:        nil,
			UserID:          database.Text(userID),
			NotificationIDs: database.TextArray(notificationIDs),
			Status:          database.Text(status),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.MockExecArgs(t, cmdTag, nil, ctx, mock.Anything, database.Text(status), database.TextArray(notificationIDs), database.Text(userID))
			},
		},
		{
			Name:            "no rows affect",
			ExpecErr:        fmt.Errorf("no rows affected"),
			UserID:          database.Text(userID),
			NotificationIDs: database.TextArray(notificationIDs),
			Status:          database.Text(status),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				mockDB.MockExecArgs(t, cmdTag, nil, ctx, mock.Anything, database.Text(status), database.TextArray(notificationIDs), database.Text(userID))
			},
		},
		{
			Name:            "err",
			ExpecErr:        puddle.ErrClosedPool,
			UserID:          database.Text(userID),
			NotificationIDs: database.TextArray(notificationIDs),
			Status:          database.Text(status),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.MockExecArgs(t, cmdTag, puddle.ErrClosedPool, ctx, mock.Anything, database.Text(status), database.TextArray(notificationIDs), database.Text(userID))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx)
			err := repo.SetStatusByNotificationIDs(ctx, mockDB.DB, tc.UserID, tc.NotificationIDs, tc.Status)
			if tc.ExpecErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.ExpecErr.Error(), err.Error())
			}
		})
	}
}

func TestUsersInfoNotificationRepo_FindUserIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()
	filter := NewFindUserNotificationFilter()
	userInfoNotificationsMap := make(map[string]entities.UserInfoNotifications)
	userInfoNotification1 := &entities.UserInfoNotification{}
	userInfoNotification2 := &entities.UserInfoNotification{}
	database.AllRandomEntity(userInfoNotification1)
	database.AllRandomEntity(userInfoNotification2)
	userInfoNotificationsMap[userInfoNotification1.NotificationID.String] = append(userInfoNotificationsMap[userInfoNotification1.NotificationID.String], userInfoNotification1)
	userInfoNotificationsMap[userInfoNotification2.NotificationID.String] = append(userInfoNotificationsMap[userInfoNotification2.NotificationID.String], userInfoNotification2)
	testCases := []struct {
		Name     string
		Setup    func(ctx context.Context)
		ExpecErr error
		Filter   FindUserNotificationFilter
	}{
		{
			Name:     "happy case",
			Filter:   filter,
			ExpecErr: nil,
			Setup: func(ctx context.Context) {
				fields, values1 := userInfoNotification1.FieldMap()
				_, values2 := userInfoNotification2.FieldMap()
				mockDB.MockQueryArgs(t, nil, ctx, mock.Anything,
					filter.NotiIDs,
					filter.UserStatus,
					filter.OffsetText,
					filter.Limit)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values1, values2})
			},
		},
		{
			Name:     "err scan",
			Filter:   filter,
			ExpecErr: fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows),
			Setup: func(ctx context.Context) {
				fields, values1 := userInfoNotification1.FieldMap()
				_, values2 := userInfoNotification2.FieldMap()
				mockDB.MockQueryArgs(t, nil, ctx, mock.Anything,
					filter.NotiIDs,
					filter.UserStatus,
					filter.OffsetText,
					filter.Limit)
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values1, values2})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx)
			res, err := repo.FindUserIDs(ctx, mockDB.DB, tc.Filter)
			if tc.ExpecErr == nil {
				assert.Nil(t, err)
				assert.Equal(t, userInfoNotificationsMap[userInfoNotification1.NotificationID.String], res[userInfoNotification1.NotificationID.String])
				assert.Equal(t, userInfoNotificationsMap[userInfoNotification2.NotificationID.String], res[userInfoNotification2.NotificationID.String])
			} else {
				assert.Equal(t, tc.ExpecErr.Error(), err.Error())
			}
		})
	}
}

func TestUsersInfoNotificationRepo_SetQuestionnareStatusAndSubmittedAt(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	repo := &UsersInfoNotificationRepo{}
	ctx := context.Background()
	userNotiID := "user-noti-id"
	status := "status"
	submittedAt := time.Now()
	testCases := []struct {
		Name               string
		ExpcErr            error
		Setup              func(ctx context.Context)
		UserNotificationID string
		Status             string
		SubmittedAt        pgtype.Timestamptz
	}{
		{
			Name:               "happy case",
			ExpcErr:            nil,
			UserNotificationID: userNotiID,
			Status:             status,
			SubmittedAt:        database.Timestamptz(submittedAt),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.MockExecArgs(t, cmdTag, nil, ctx, mock.Anything, status, database.Timestamptz(submittedAt), userNotiID)
			},
		},
		{
			Name:               "no row affected",
			ExpcErr:            fmt.Errorf("no rows affected"),
			UserNotificationID: userNotiID,
			Status:             status,
			SubmittedAt:        database.Timestamptz(submittedAt),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				mockDB.MockExecArgs(t, cmdTag, nil, ctx, mock.Anything, status, database.Timestamptz(submittedAt), userNotiID)
			},
		},
		{
			Name:               "err",
			ExpcErr:            puddle.ErrClosedPool,
			UserNotificationID: userNotiID,
			Status:             status,
			SubmittedAt:        database.Timestamptz(submittedAt),
			Setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				mockDB.MockExecArgs(t, cmdTag, puddle.ErrClosedPool, ctx, mock.Anything, status, database.Timestamptz(submittedAt), userNotiID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(ctx)
			err := repo.SetQuestionnareStatusAndSubmittedAt(ctx, mockDB.DB, tc.UserNotificationID, tc.Status, tc.SubmittedAt)
			if tc.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.ExpcErr.Error(), err.Error())
			}
		})
	}
}

func TestUsersInfoNotificationRepo_GetNotificationIDWithFullyQnStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &UsersInfoNotificationRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	status := database.Text("answered_status")
	notiIDsReq := database.TextArray([]string{"notification_id-1", "notification_id-2"})
	notficationIDsRes := []pgtype.Text{database.Text("notification_id-1"), database.Text("notification_id-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, notiIDsReq, status)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		_, err := r.GetNotificationIDWithFullyQnStatus(ctx, db, notiIDsReq, status)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, notiIDsReq, status)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		notIDs, err := r.GetNotificationIDWithFullyQnStatus(ctx, db, notiIDsReq, status)
		assert.NoError(t, err)
		for i, notID := range notIDs {
			assert.Equal(t, notID, notficationIDsRes[i].String)
		}
	})
}

func TestUsersInfoNotificationRepo_SoftDeleteByNotificationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &UsersInfoNotificationRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.Anything, database.Text("noti-id"))

		err := r.SoftDeleteByNotificationID(ctx, db, "noti-id")
		assert.NoError(t, err)
	})
}
