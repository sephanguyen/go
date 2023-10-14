package repositories

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQuestionnaireRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *entities.Questionnaire
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &entities.Questionnaire{},
			SetUp: func(ctx context.Context) {
				e := &entities.Questionnaire{}
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

	repo := &QuestionnaireRepo{}
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

func TestQuestionnaireRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	type Req struct {
		QuestionnaireIds []string
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
				QuestionnaireIds: []string{"1"},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				db.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &QuestionnaireRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.SoftDelete(ctx, db, testCase.Req.(*Req).QuestionnaireIds)
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestQuestionnaireRepo_FindQuestionsByQnID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &QuestionnaireRepo{}

	ent1 := &entities.QuestionnaireQuestion{}
	ent2 := &entities.QuestionnaireQuestion{}
	database.AllRandomEntity(ent1)
	database.AllRandomEntity(ent2)
	qnID := "qn_1"
	t.Run("success", func(t *testing.T) {
		fields, vals1 := ent1.FieldMap()
		_, vals2 := ent2.FieldMap()
		// scan twice using values from 2 entities
		mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(qnID))
		questions, err := r.FindQuestionsByQnID(ctx, db, qnID)
		assert.Nil(t, err)
		assert.Equal(t, ent1, questions[0])
		assert.Equal(t, ent2, questions[1])
	})
	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text(qnID))
		questions, err := r.FindQuestionsByQnID(ctx, db, qnID)
		assert.Nil(t, questions)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
	t.Run("error scan", func(t *testing.T) {
		fields, vals := ent1.FieldMap()
		mockDB.MockScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(qnID))
		questions, err := r.FindQuestionsByQnID(ctx, db, qnID)
		assert.Nil(t, questions)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestQuestionnaireRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &QuestionnaireRepo{}

	entInDB := &entities.Questionnaire{}
	database.AllRandomEntity(entInDB)
	qnID := "qn_1"
	t.Run("success", func(t *testing.T) {

		fields, vals := entInDB.FieldMap()
		mockDB.MockRowScanFields(nil, fields, vals)

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(qnID))
		retQn, err := r.FindByID(ctx, db, qnID)
		assert.Nil(t, err)
		assert.Equal(t, entInDB, retQn)
	})
	t.Run("err query row", func(t *testing.T) {
		fields, vals := entInDB.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(qnID))
		retQn, err := r.FindByID(ctx, db, qnID)
		assert.Nil(t, retQn)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestQuestionnaireRepo_FindUserAnswers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &QuestionnaireRepo{}

	entInDB1 := &entities.QuestionnaireUserAnswer{}
	entInDB2 := &entities.QuestionnaireUserAnswer{}
	database.AllRandomEntity(entInDB1)
	database.AllRandomEntity(entInDB2)

	filter := NewFindUserAnswersFilter()

	t.Run("success", func(t *testing.T) {
		filter.QuestionnaireQuestionIDs = database.TextArray([]string{mock.Anything, mock.Anything})
		filter.TargetIDs = database.TextArray([]string{mock.Anything, mock.Anything})
		filter.UserIDs = database.TextArray([]string{mock.Anything, mock.Anything})
		filter.UserNotificationIDs = database.TextArray([]string{mock.Anything, mock.Anything})

		fields, vals1 := entInDB1.FieldMap()
		_, vals2 := entInDB2.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.QuestionnaireQuestionIDs, filter.UserNotificationIDs, filter.UserIDs, filter.TargetIDs)

		answers, err := r.FindUserAnswers(ctx, db, &filter)
		assert.Nil(t, err)
		assert.Equal(t, entInDB1, answers[0])
		assert.Equal(t, entInDB2, answers[1])
	})
	t.Run("error scan", func(t *testing.T) {
		fields, vals1 := entInDB1.FieldMap()
		_, vals2 := entInDB2.FieldMap()
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{vals1, vals2})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.QuestionnaireQuestionIDs, filter.UserNotificationIDs, filter.UserIDs, filter.TargetIDs)

		answers, err := r.FindUserAnswers(ctx, db, &filter)
		assert.Nil(t, answers)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, filter.QuestionnaireQuestionIDs, filter.UserNotificationIDs, filter.UserIDs, filter.TargetIDs)

		answers, err := r.FindUserAnswers(ctx, db, &filter)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, answers)
	})
}

func TestQuestionnaireRepoV2_FindQuestionnaireResponders(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &QuestionnaireRepo{}

	responder1 := &QuestionnaireResponder{
		UserNotificationID: database.Text(idutil.ULIDNow()),
		UserID:             database.Text(idutil.ULIDNow()),
		IsParent:           database.Bool(true),
		TargetID:           database.Text(idutil.ULIDNow()),
		Name:               database.Text(idutil.ULIDNow()),
		TargetName:         database.Text(idutil.ULIDNow()),
		SubmittedAt:        database.Timestamptz(time.Now()),
		IsIndividual:       database.Bool(true),
	}
	responder2 := &QuestionnaireResponder{
		UserNotificationID: database.Text(idutil.ULIDNow()),
		UserID:             database.Text(idutil.ULIDNow()),
		IsParent:           database.Bool(false),
		TargetID:           database.Text(idutil.ULIDNow()),
		Name:               database.Text(idutil.ULIDNow()),
		TargetName:         database.Text(idutil.ULIDNow()),
		SubmittedAt:        database.Timestamptz(time.Now()),
		IsIndividual:       database.Bool(false),
	}

	var totalCount uint32 = 2

	filter := NewFindQuestionnaireRespondersFilter()

	t.Run("success", func(t *testing.T) {
		filter.QuestionnaireID = database.Text(mock.Anything)
		filter.UserName = database.Text("")
		filter.Limit = database.Int8(math.MaxUint32)
		filter.Offset = database.Int8(math.MaxUint32)

		mockDB.MockScanArray(nil, []string{"user_notification_id", "user_id", "is_parent", "name", "target_id", "submitted_at", "target_name", "is_individual"}, [][]interface{}{{&responder1.UserNotificationID, &responder1.UserID, &responder1.IsParent, &responder1.Name, &responder1.TargetID, &responder1.SubmittedAt, &responder1.TargetName, &responder1.IsIndividual}, {&responder2.UserNotificationID, &responder2.UserID, &responder2.IsParent, &responder2.Name, &responder2.TargetID, &responder2.SubmittedAt, &responder2.TargetName, &responder2.IsIndividual}})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.UserName, filter.QuestionnaireID, filter.Limit, filter.Offset)
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, filter.UserName, filter.QuestionnaireID)
		mockDB.MockRowScanFields(nil, []string{"total_count"}, []interface{}{&totalCount})

		totalCountRes, responders, err := r.FindQuestionnaireResponders(ctx, db, &filter)
		assert.Nil(t, err)
		assert.Equal(t, responder1, responders[0])
		assert.Equal(t, responder2, responders[1])
		assert.Equal(t, totalCount, totalCountRes)
	})
	t.Run("error scan", func(t *testing.T) {
		mockDB.MockScanArray(pgx.ErrNoRows, []string{"user_notification_id", "user_id", "is_parent", "name", "target_id", "submitted_at", "target_name", "is_individual"}, [][]interface{}{{&responder1.UserNotificationID, &responder1.UserID, &responder1.IsParent, &responder1.Name, &responder1.TargetID, &responder1.SubmittedAt, &responder1.TargetName, &responder1.IsIndividual}, {&responder2.UserNotificationID, &responder2.UserID, &responder2.IsParent, &responder2.Name, &responder2.TargetID, &responder2.SubmittedAt, &responder2.TargetName, &responder2.IsIndividual}})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.UserName, filter.QuestionnaireID, filter.Limit, filter.Offset)

		_, answers, err := r.FindQuestionnaireResponders(ctx, db, &filter)
		assert.Nil(t, answers)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, filter.UserName, filter.QuestionnaireID, filter.Limit, filter.Offset)

		_, answers, err := r.FindQuestionnaireResponders(ctx, db, &filter)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, answers)
	})
}

func TestQuestionnaireRepo_FindQuestionnaireCSVResponders(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &QuestionnaireRepo{}

	responder1 := &QuestionnaireCSVResponder{
		UserNotificationID: database.Text(idutil.ULIDNow()),
		UserID:             database.Text(idutil.ULIDNow()),
		IsParent:           database.Bool(true),
		TargetID:           database.Text(idutil.ULIDNow()),
		StudentID:          database.Text(""),
		Name:               database.Text(idutil.ULIDNow()),
		TargetName:         database.Text(idutil.ULIDNow()),
		SubmittedAt:        database.Timestamptz(time.Now()),
		IsIndividual:       database.Bool(true),
		LocationNames:      database.TextArray([]string{idutil.ULIDNow()}),
		StudentExternalID:  database.Text(""),
		SubmissionStatus:   database.Text("USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED"),
	}
	responder2 := &QuestionnaireCSVResponder{
		UserNotificationID: database.Text(idutil.ULIDNow()),
		UserID:             database.Text(idutil.ULIDNow()),
		IsParent:           database.Bool(false),
		TargetID:           database.Text(idutil.ULIDNow()),
		StudentID:          database.Text(idutil.ULIDNow()),
		Name:               database.Text(idutil.ULIDNow()),
		TargetName:         database.Text(idutil.ULIDNow()),
		SubmittedAt:        database.Timestamptz(time.Now()),
		IsIndividual:       database.Bool(false),
		LocationNames:      database.TextArray([]string{idutil.ULIDNow()}),
		StudentExternalID:  database.Text(idutil.ULIDNow()),
		SubmissionStatus:   database.Text("USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED"),
	}

	questionnaireID := database.Text("questionnaire-id")

	t.Run("success", func(t *testing.T) {
		mockDB.MockScanArray(nil, []string{
			"user_notification_id",
			"user_id",
			"is_parent",
			"name",
			"target_id",
			"student_id",
			"submitted_at",
			"target_name",
			"is_individual",
			"student_external_id",
			"location_names",
			"submission_status",
		}, [][]interface{}{
			{
				&responder1.UserNotificationID,
				&responder1.UserID,
				&responder1.IsParent,
				&responder1.Name,
				&responder1.TargetID,
				&responder1.StudentID,
				&responder1.SubmittedAt,
				&responder1.TargetName,
				&responder1.IsIndividual,
				&responder1.StudentExternalID,
				&responder1.LocationNames,
				&responder1.SubmissionStatus,
			}, {
				&responder2.UserNotificationID,
				&responder2.UserID,
				&responder2.IsParent,
				&responder2.Name,
				&responder2.TargetID,
				&responder2.StudentID,
				&responder2.SubmittedAt,
				&responder2.TargetName,
				&responder2.IsIndividual,
				&responder2.StudentExternalID,
				&responder2.LocationNames,
				&responder2.SubmissionStatus,
			},
		})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, questionnaireID)

		responders, err := r.FindQuestionnaireCSVResponders(ctx, db, questionnaireID.String)
		assert.Nil(t, err)
		assert.Equal(t, responder1, responders[0])
		assert.Equal(t, responder2, responders[1])
	})
	t.Run("error scan", func(t *testing.T) {
		mockDB.MockScanArray(pgx.ErrNoRows, []string{
			"user_notification_id",
			"user_id",
			"is_parent",
			"name",
			"target_id",
			"student_id",
			"submitted_at",
			"target_name",
			"is_individual",
			"student_external_id",
			"location_names",
			"submission_status",
		}, [][]interface{}{
			{
				&responder1.UserNotificationID,
				&responder1.UserID,
				&responder1.IsParent,
				&responder1.Name,
				&responder1.TargetID,
				&responder1.StudentID,
				&responder1.SubmittedAt,
				&responder1.TargetName,
				&responder1.IsIndividual,
				&responder1.StudentExternalID,
				&responder1.LocationNames,
				&responder1.SubmissionStatus,
			}, {
				&responder2.UserNotificationID,
				&responder2.UserID,
				&responder2.IsParent,
				&responder2.Name,
				&responder2.TargetID,
				&responder2.StudentID,
				&responder2.SubmittedAt,
				&responder2.TargetName,
				&responder2.IsIndividual,
				&responder2.StudentExternalID,
				&responder2.LocationNames,
				&responder2.SubmissionStatus,
			},
		})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, questionnaireID)

		responders, err := r.FindQuestionnaireCSVResponders(ctx, db, questionnaireID.String)
		assert.Nil(t, responders)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, questionnaireID)

		responders, err := r.FindQuestionnaireCSVResponders(ctx, db, questionnaireID.String)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, responders)
	})
}

func TestQuestionnaireQuestionRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	type Req struct {
		QuestionnaireIds []string
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
				QuestionnaireIds: []string{"1"},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				db.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &QuestionnaireQuestionRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.SoftDelete(ctx, db, testCase.Req.(*Req).QuestionnaireIds)
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestQuestionnaireQuestionRepo_BulkForceUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	testCases := []struct {
		Name        string
		Req         interface{}
		ExpectedErr error
		SetUp       func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: []*entities.QuestionnaireQuestion{
				{
					QuestionnaireQuestionID: database.Text("1"),
				},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	repo := &QuestionnaireQuestionRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkForceUpsert(ctx, db, testCase.Req.([]*entities.QuestionnaireQuestion))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestQuestionnaireUserAnswerRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	type Req struct {
		AnswerIDs []string
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
				AnswerIDs: []string{"1"},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				db.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &QuestionnaireUserAnswerRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.SoftDelete(ctx, db, testCase.Req.(*Req).AnswerIDs)
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestQuestionnaireUserAnswerRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	testCases := []struct {
		Name        string
		Req         interface{}
		ExpectedErr error
		SetUp       func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: entities.QuestionnaireUserAnswers{
				{
					AnswerID:                database.Text("answer-id-1"),
					UserNotificationID:      database.Text("user-noti-id-1"),
					QuestionnaireQuestionID: database.Text("questionnaire-question-id-1"),
					UserID:                  database.Text("user-id-1"),
					TargetID:                database.Text("target-id-1"),
					Answer:                  database.Text("answer"),
					SubmittedAt:             database.Timestamptz(time.Now()),
				},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	repo := &QuestionnaireUserAnswerRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.(entities.QuestionnaireUserAnswers))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestQuestionnaireUserAnswerRepo_SoftDeleteByQuestionnaireID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &QuestionnaireUserAnswerRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.Anything, database.TextArray([]string{"questionnaire-id"}))

		err := r.SoftDeleteByQuestionnaireID(ctx, db, []string{"questionnaire-id"})
		assert.NoError(t, err)
	})
}
