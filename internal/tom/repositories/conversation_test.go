package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationRepo_CountUnreadConversationsByAccessPaths(t *testing.T) {
	t.Parallel()
	repo := &ConversationRepo{}
	db := testutil.NewMockDB()
	userID := database.Text("user")
	msgType := database.TextArray([]string{"CONVERSATION_STUDENT"})
	ap := database.TextArray([]string{"loc1/loca", "loc1/locb"})
	total := int64(1)
	db.MockQueryRowArgs(t, mock.Anything, mock.MatchedBy(func(sql string) bool {
		// at least check its syntax
		testutil.ParseSQL(t, sql)
		return true
	}), userID, msgType, database.TextArray([]string{"loc1/loca%", "loc1/locb%"}))
	db.MockRowScanFields(nil, []string{""}, []interface{}{&total})
	ret, err := repo.CountUnreadConversationsByAccessPaths(context.Background(), db.DB, userID, msgType, ap)
	assert.NoError(t, err)
	assert.Equal(t, total, ret)
}

func TestConversationRepo_CountUnreadConversationsByAccessPathsV2(t *testing.T) {
	t.Parallel()
	repo := &ConversationRepo{}
	db := testutil.NewMockDB()
	userID := database.Text("user")
	studentAps := database.TextArray([]string{"loc1/loca", "loc1/locb"})
	parentAps := database.TextArray([]string{"loc1/loca", "loc1/locb"})
	total := int64(1)
	db.MockQueryRowArgs(t, mock.Anything, mock.MatchedBy(func(sql string) bool {
		// at least check its syntax
		testutil.ParseSQL(t, sql)
		return true
	}), userID, database.TextArray([]string{"loc1/loca%", "loc1/locb%"}), database.TextArray([]string{"loc1/loca%", "loc1/locb%"}))
	db.MockRowScanFields(nil, []string{""}, []interface{}{&total})
	ret, err := repo.CountUnreadConversationsByAccessPathsV2(context.Background(), db.DB, userID, studentAps, parentAps)
	assert.NoError(t, err)
	assert.Equal(t, total, ret)
}

func TestConversationRepo_BulkUpdateResourcePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationRepo{}

	offsetID := pgtype.Text{}
	offsetID.Set(nil)
	convIDs := []string{"conversation-1", "conversation-2"}
	resourcePath := "manabie"

	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "conversations" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.Text(resourcePath), database.TextArray(convIDs))

		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "conversation_members" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.TextArray(convIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "messages" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.TextArray(convIDs))

		err := r.BulkUpdateResourcePath(ctx, db, convIDs, resourcePath)
		assert.NoError(t, err)
	})
	t.Run("err select", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "conversations" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.Text(resourcePath), database.TextArray(convIDs))

		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "conversation_members" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.TextArray(convIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, puddle.ErrClosedPool, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "messages" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.TextArray(convIDs))

		err := r.BulkUpdateResourcePath(ctx, db, convIDs, resourcePath)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
}

func TestConversationRepo_ListAll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationRepo{}

	offsetID := pgtype.Text{}
	offsetID.Set(nil)
	schoolID := database.Text("manabie")
	limit := uint32(10)
	conversationTypesAccepted := database.TextArray([]string{"CONVERSATION_STUDENT", "CONVERSATION_PARENT"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, offsetID, limit, conversationTypesAccepted, schoolID)

		conversation, err := r.ListAll(ctx, db, offsetID, limit, conversationTypesAccepted, schoolID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success - happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, offsetID, limit, conversationTypesAccepted, schoolID)
		e := &entities.Conversation{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.ListAll(ctx, db, offsetID, limit, conversationTypesAccepted, schoolID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, "conversation_id", "name", "status", "conversation_type", "created_at", "updated_at", "last_message_id", "owner")

		mockDB.RawStmt.AssertSelectedTable(t, "conversations", "")
	})
}

func TestConversationRepo_FindByIDsReturnMapByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	rows := mockDB.Rows
	r := &ConversationRepo{}
	e := &entities.Conversation{}

	conversationIDs := database.TextArray([]string{"conversation-1", "conversation-2", "conversation-3"})
	staffRoles := database.TextArray(constant.ConversationStaffRoles)

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &staffRoles, &conversationIDs)
		conversation, err := r.FindByIDsReturnMapByID(ctx, db, conversationIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success", func(t *testing.T) {
		expectedMap := make(map[pgtype.Text]core.ConversationFull)
		expectedMap[pgtype.Text{}] = core.ConversationFull{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &staffRoles, &conversationIDs)

		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		var (
			isReply   pgtype.Bool
			studentID pgtype.Text
		)
		scanFields = append(scanFields, &isReply, &studentID)

		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)
		rows.On("Close").Once().Return(nil)

		conversation, err := r.FindByIDsReturnMapByID(ctx, db, conversationIDs)
		assert.Equal(t, err, nil)
		assert.Equal(t, expectedMap, conversation)
		assert.Equal(t, len(expectedMap), len(conversation), "the length of map conversationfull not equal")
	})
}

func TestConversationRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	row := mockDB.Row

	r := &ConversationRepo{}
	e := &entities.Conversation{}
	_, values := e.FieldMap()
	conversationID := database.Text("conversation-1")
	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &conversationID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(puddle.ErrClosedPool)
		conversation, err := r.FindByID(ctx, db, conversationID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &conversationID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(nil)
		conversation, err := r.FindByID(ctx, db, conversationID)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, conversation)
		mockDB.RawStmt.AssertSelectedFields(t, "conversation_id", "name", "status", "conversation_type", "created_at", "updated_at", "last_message_id", "owner")
		mockDB.RawStmt.AssertSelectedTable(t, "conversations", "")
	})
}

func TestConversationRepo_ListConversationUnjoinedInLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationRepo{}

	filter := &core.ListConversationUnjoinedFilter{
		UserID:      database.Text("user-id"),
		OwnerIDs:    database.TextArray([]string{"owner-id-1"}),
		AccessPaths: database.TextArray([]string{"orgloc/loc1", "orgloc/loc2"}),
	}
	likeAccpaths := []string{"orgloc/loc1%", "orgloc/loc2%"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, filter.OwnerIDs, filter.UserID, database.TextArray(likeAccpaths))

		conversation, err := r.ListConversationUnjoinedInLocations(ctx, db, filter)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success - happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.OwnerIDs, filter.UserID, database.TextArray(likeAccpaths))
		e := &entities.Conversation{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.ListConversationUnjoinedInLocations(ctx, db, filter)
		assert.Nil(t, err)
	})
}

func TestConversationRepo_ListConversationUnjoined(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationRepo{}

	filter := &core.ListConversationUnjoinedFilter{
		UserID:   database.Text("user-id"),
		OwnerIDs: database.TextArray([]string{"owner-id-1"}),
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &filter.OwnerIDs, &filter.UserID)

		conversation, err := r.ListConversationUnjoined(ctx, db, filter)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success - happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		e := &entities.Conversation{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.ListConversationUnjoined(ctx, db, filter)
		assert.Nil(t, err)
	})
}

func TestConversationRepo_FindByStudentQuestionID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationRepo{}

	studentQuestionID := database.Text("student-question-id-1")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&studentQuestionID)
		e := &entities.Conversation{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(puddle.ErrClosedPool, fields, values)
		conversation, err := r.FindByStudentQuestionID(ctx, db, studentQuestionID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success - happy case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&studentQuestionID)
		e := &entities.Conversation{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)

		_, err := r.FindByStudentQuestionID(ctx, db, studentQuestionID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, "conversation_id", "name", "status", "conversation_type", "created_at", "updated_at", "last_message_id", "owner")

		mockDB.RawStmt.AssertSelectedTable(t, "conversations", "")
	})
}

func TestConversationRepo_FindByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	row := mockDB.Row

	r := &ConversationRepo{}
	templateConv := &entities.Conversation{
		ID:               randomText(),
		Name:             randomText(),
		ConversationType: randomText(),
		Status:           randomText(),
	}
	emptyConv := &entities.Conversation{}
	_, scannedValues := emptyConv.FieldMap()
	lessonID := database.Text("lesson-1")
	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", scannedValues...).Once().Return(puddle.ErrClosedPool)
		conversation, err := r.FindByLessonID(ctx, db, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		db.On("QueryRow").Once().Return(row, nil)
		fields, templateValues := templateConv.FieldMap()
		mockDB.MockRowScanFields(nil, fields, templateValues)
		row.On("Scan", scannedValues...).Once().Return(nil)
		conversation, err := r.FindByLessonID(ctx, db, lessonID)
		assert.True(t, errors.Is(err, nil))
		assert.Equal(t, conversation, templateConv)
	})
}
