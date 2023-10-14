package services

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo_lessonmgmt "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShareMaterialCommandHandler(t *testing.T) {
	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonRoomStateRepo := &mock_repo_lessonmgmt.MockLessonRoomStateRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)

	tcs := []struct {
		name     string
		command  *ShareMaterialCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute share material command successfully",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				State: &CurrentMaterial{
					MediaID: "media-2",
					VideoState: &VideoState{
						CurrentTime: Duration(12 * time.Second),
						PlayerState: PlayerStatePause,
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &CurrentMaterial{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, "media-2", state.MediaID)
						assert.Equal(t, 12*time.Second, state.VideoState.CurrentTime.Duration())
						assert.Equal(t, PlayerStatePause, state.VideoState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
		},
		{
			name: "execute share material command without video state",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				State: &CurrentMaterial{
					MediaID: "media-2",
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &CurrentMaterial{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, "media-2", state.MediaID)
						assert.Nil(t, state.VideoState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
		},
		{
			name: "execute share material command without material",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentMaterial)
					}).
					Return(nil).Once()
			},
		},
		{
			name: "execute share material command when current material is null",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				State: &CurrentMaterial{
					MediaID: "media-2",
					VideoState: &VideoState{
						CurrentTime: Duration(12 * time.Second),
						PlayerState: PlayerStatePause,
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState:     database.JSONB(nil),
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &CurrentMaterial{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, "media-2", state.MediaID)
						assert.Equal(t, 12*time.Second, state.VideoState.CurrentTime.Duration())
						assert.Equal(t, PlayerStatePause, state.VideoState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
		},
		{
			name: "execute share material command without media id",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				State: &CurrentMaterial{
					MediaID: "",
					VideoState: &VideoState{
						CurrentTime: Duration(12 * time.Second),
						PlayerState: PlayerStatePause,
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute share material command without PlayerState",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				State: &CurrentMaterial{
					MediaID: "media-2",
					VideoState: &VideoState{
						CurrentTime: Duration(12 * time.Second),
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute share material command with media id not belong to lesson",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				State: &CurrentMaterial{
					MediaID: "media-5",
					VideoState: &VideoState{
						CurrentTime: Duration(12 * time.Second),
						PlayerState: PlayerStatePause,
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID:      database.Text("lesson-1"),
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonGroupRepo.
					On("Get", ctx, tx, database.Text("lesson-group-1"), database.Text("course-1")).
					Return(&entities.LessonGroup{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)
			handler := &ShareMaterialCommandHandler{
				command:             tc.command,
				DB:                  db,
				LessonRepo:          lessonRepo,
				LessonMemberRepo:    lessonMemberRepo,
				LessonGroupRepo:     lessonGroupRepo,
				LessonRoomStateRepo: lessonRoomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonGroupRepo, lessonMemberRepo)
		})
	}
}

func TestStopSharingMaterialCommandHandler(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonRoomStateRepo := &mock_repo_lessonmgmt.MockLessonRoomStateRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)

	tcs := []struct {
		name     string
		command  *StopSharingMaterialCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute stop share material command successfully",
			command: &StopSharingMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentMaterial)
					}).
					Return(nil).Once()
			},
		},
		{
			name: "execute stop share material command with non-existing lesson id",
			command: &StopSharingMaterialCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(nil, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &StopSharingMaterialCommandHandler{
				command:             tc.command,
				DB:                  db,
				LessonRepo:          lessonRepo,
				LessonMemberRepo:    lessonMemberRepo,
				LessonRoomStateRepo: lessonRoomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonMemberRepo)
		})
	}
}

func TestFoldHandAllCommandHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	db := &mock_database.Ext{}

	tcs := []struct {
		name    string
		command *FoldHandAllCommand
		setup   func(ctx context.Context)
	}{
		{
			name: "execute fold hand all member command successfully",
			command: &FoldHandAllCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						db,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeHandsUp)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &FoldHandAllCommandHandler{
				command:          tc.command,
				DB:               db,
				LessonMemberRepo: lessonMemberRepo,
			}
			err := handler.Execute(ctx)
			require.NoError(t, err)

			mock.AssertExpectationsForObjects(t, db, lessonMemberRepo)
		})
	}
}

func TestUpdateHandsUpCommandHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	db := &mock_database.Ext{}

	tcs := []struct {
		name     string
		command  *UpdateHandsUpCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute hands up command successfully",
			command: &UpdateHandsUpCommand{
				CommanderID: "student-1",
				UserID:      "student-1",
				LessonID:    "lesson-1",
				State: &UserHandsUp{
					Value: true,
				},
			},
			setup: func(ctx context.Context) {
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "student-1", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, true, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
			},
		},
		{
			name: "execute hands up command unsuccessfully",
			command: &UpdateHandsUpCommand{
				CommanderID: "student-2",
				UserID:      "student-2",
				LessonID:    "lesson-1",
				State: &UserHandsUp{
					Value: true,
				},
			},
			setup: func(ctx context.Context) {
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "student-2", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, true, state.BoolValue.Bool)
					}).
					Return(fmt.Errorf("got a error")).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &UpdateHandsUpCommandHandler{
				command:          tc.command,
				DB:               db,
				LessonMemberRepo: lessonMemberRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, lessonMemberRepo)
		})
	}
}

func TestResetAllStatesCommand(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonRoomStateRepo := &mock_repo_lessonmgmt.MockLessonRoomStateRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)

	tcs := []struct {
		name     string
		command  *ResetAllStatesCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute reset all states command successfully",
			command: &ResetAllStatesCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`),
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentMaterial)
					}).
					Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeHandsUp)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeAnnotation)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "` + string(nowString) + `",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				},
				"current_polling": {
					"options": [
						{
							"answer": "A",
							"is_correct": true
						},
						{
							"answer": "B",
							"is_correct": false
						},
						{
							"answer": "C",
							"is_correct": false
						}
					],
					"status": "POLLING_STATE_STARTED",
					"created_at": "` + string(nowString) + `"
				}
			}`),
					}, nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypePollingAnswer)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentPolling)
					}).
					Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeChat)),
						&entities.StateValue{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				// reset recording
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
			{
				"recording": {
					"is_recording": true,
					"creator": "user-id-1"
				}
			}`),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.False(t, state.Recording.IsRecording)
						assert.Nil(t, state.Recording.Creator)
					}).
					Return(nil).Once()
				lessonRoomStateRepo.
					On("UnSpotlight", ctx, tx, database.Text("lesson-1")).
					Return(nil).Once()
				lessonRoomStateRepo.
					On("UpsertWhiteboardZoomState", ctx, tx, "lesson-1", new(domain.WhiteboardZoomState).SetDefault()).
					Return(nil).Once()
			},
		},
		{
			name: "execute reset all states command with non-existing lesson id",
			command: &ResetAllStatesCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(nil, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &ResetAllStatesCommandHandler{
				command:             tc.command,
				DB:                  db,
				LessonRepo:          lessonRepo,
				LessonMemberRepo:    lessonMemberRepo,
				LessonRoomStateRepo: lessonRoomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonMemberRepo)
		})
	}
}

func TestUpdateAnnotationCommandHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonRepo := &mock_repositories.MockLessonRepo{}

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	learners := []string{"learner-1"}
	tcs := []struct {
		name     string
		command  *UpdateAnnotationCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute enable annotation command successfully",
			command: &UpdateAnnotationCommand{
				CommanderID: "teacher-1",
				UserIDs:     learners,
				LessonID:    "lesson-1",
				State: &UserAnnotation{
					Value: true,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeAnnotation)),
						database.TextArray(learners),
						&entities.StateValue{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
			},
		},
		{
			name: "execute disable annotation command successfully",
			command: &UpdateAnnotationCommand{
				CommanderID: "teacher-1",
				UserIDs:     learners,
				LessonID:    "lesson-1",
				State: &UserAnnotation{
					Value: false,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeAnnotation)),
						database.TextArray(learners),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &UpdateAnnotationCommandHandler{
				command:          tc.command,
				DB:               db,
				LessonMemberRepo: lessonMemberRepo,
				LessonRepo:       lessonRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonMemberRepo, lessonRepo)
		})
	}
}

func TestStartPollingCommandHandler(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	tcs := []struct {
		name     string
		command  *StartPollingCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute start polling command successfully",
			command: &StartPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				Options: []*PollingOption{
					{
						Answer:    "A",
						IsCorrect: true,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, PollingStateStarted, state.CurrentPolling.Status)
						assert.False(t, state.CurrentPolling.CreatedAt.IsZero())
						options := []*PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						}
						assert.True(t, reflect.DeepEqual(state.CurrentPolling.Options, options))
					}).
					Return(nil).Once()
			},
		},
		{
			name: "execute start polling command with option empty",
			command: &StartPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				Options:     []*PollingOption{},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute start polling command with 11 option",
			command: &StartPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				Options: []*PollingOption{
					{
						Answer:    "A",
						IsCorrect: true,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute start polling command with none correct option",
			command: &StartPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				Options: []*PollingOption{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &StartPollingCommandHandler{
				command:    tc.command,
				DB:         db,
				LessonRepo: lessonRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo)
		})
	}
}

func TestStopPollingCommandHandler(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)
	tcs := []struct {
		name     string
		command  *StopPollingCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute stop polling command successfully",
			command: &StopPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STARTED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Equal(t, PollingStateStopped, state.CurrentPolling.Status)
						assert.False(t, state.CurrentPolling.CreatedAt.IsZero())
						assert.False(t, now.Equal(state.CurrentPolling.StoppedAt))
					}).
					Return(nil).Once()
			},
		},
		{
			name: "execute stop polling command when none polling",
			command: &StopPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute stop polling command when stopped polling",
			command: &StopPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STOPPED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &StopPollingCommandHandler{
				command:    tc.command,
				DB:         db,
				LessonRepo: lessonRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo)
		})
	}
}

func TestEndPollingCommandHandler(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonPollingRepo := &mock_repositories.MockLessonPollingRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)
	tcs := []struct {
		name     string
		command  *EndPollingCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute end polling command successfully",
			command: &EndPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STOPPED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
				lessonMemberRepo.
					On(
						"GetLessonMemberStatesWithParams",
						ctx,
						tx,
						mock.Anything,
					).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue:        database.Bool(false),
								StringArrayValue: database.TextArray([]string{"A"}),
								DeleteAt:         database.Timestamptz(now),
							},
						},
						nil,
					).
					Once()
				lessonPollingRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(&entities.LessonPolling{
						PollID: database.Text("poll-1"),
					}, nil).Once()
				lessonRepo.
					On("UpdateLessonRoomState", ctx, tx, database.Text("lesson-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.Nil(t, state.CurrentPolling)
					}).
					Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypePollingAnswer)),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
			},
		},
		{
			name: "execute end polling command when none polling",
			command: &EndPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute end polling command when started polling",
			command: &EndPollingCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.
					On("FindByID", ctx, tx, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STARTED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &EndPollingCommandHandler{
				command:           tc.command,
				DB:                db,
				LessonRepo:        lessonRepo,
				LessonMemberRepo:  lessonMemberRepo,
				LessonPollingRepo: lessonPollingRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonMemberRepo, lessonPollingRepo)
		})
	}
}

func TestSubmitPollingAnswerCommandHandler(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	db := &mock_database.Ext{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)
	tcs := []struct {
		name     string
		command  *SubmitPollingAnswerCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute submit polling answer command successfully",
			command: &SubmitPollingAnswerCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				UserID:      "student-1",
				Answers:     []string{"A"},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STARTED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
				lessonMemberRepo.
					On(
						"GetLessonMemberStatesWithParams",
						ctx,
						db,
						mock.Anything,
					).
					Return(
						entities.LessonMemberStates{},
						nil,
					).
					Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*entities.LessonMemberState)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "student-1", state.UserID.String)
						assert.Equal(t, string(LearnerStateTypePollingAnswer), state.StateType.String)
					}).
					Return(nil).
					Once()
			},
		},
		{
			name: "execute submit polling answer command when none polling",
			command: &SubmitPollingAnswerCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				UserID:      "student-1",
				Answers:     []string{"A"},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute submit polling answer command which belong to options of current polling",
			command: &SubmitPollingAnswerCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				UserID:      "student-1",
				Answers:     []string{"D"},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STARTED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute submit polling answer command when stopped polling",
			command: &SubmitPollingAnswerCommand{
				CommanderID: "teacher-1",
				LessonID:    "lesson-1",
				UserID:      "student-1",
				Answers:     []string{"A"},
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
						RoomState: database.JSONB(`
						{
							"current_polling": {
								"options": [
									{
										"answer": "A",
										"is_correct": true
									},
									{
										"answer": "B",
										"is_correct": false
									},
									{
										"answer": "C",
										"is_correct": false
									}
								],
								"status": "POLLING_STATE_STOPPED",
								"created_at": "` + string(nowString) + `"
							}
						}`),
					}, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &SubmitPollingAnswerCommandHandler{
				command:          tc.command,
				DB:               db,
				LessonRepo:       lessonRepo,
				LessonMemberRepo: lessonMemberRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo, lessonMemberRepo)
		})
	}
}

func TestRequestRecordingHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	lessonRepo := &mock_repositories.MockLessonRepo{}
	tcs := []struct {
		name     string
		command  *RequestRecordingCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "request recording successfully",
			command: &RequestRecordingCommand{
				CommanderID: "user-id-1",
				LessonID:    "lesson-id-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("GrantRecordingPermission", ctx, tx, database.Text("lesson-id-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.True(t, state.Recording.IsRecording)
						assert.Equal(t, "user-id-1", *state.Recording.Creator)
					}).
					Return(nil).Once()
			},
		},
		{
			name: "request recording failed",
			command: &RequestRecordingCommand{
				CommanderID: "user-id-1",
				LessonID:    "lesson-id-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.On("GrantRecordingPermission", ctx, tx, database.Text("lesson-id-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(3).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.True(t, state.Recording.IsRecording)
						assert.Equal(t, "user-id-1", *state.Recording.Creator)
					}).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &RequestRecordingHandler{
				command:    tc.command,
				DB:         db,
				LessonRepo: lessonRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo)
		})
	}
}

func TestStopRecordingHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	lessonRepo := &mock_repositories.MockLessonRepo{}
	tcs := []struct {
		name     string
		command  *StopRecordingCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "stop recording successfully",
			command: &StopRecordingCommand{
				CommanderID: "user-id-1",
				LessonID:    "lesson-id-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("StopRecording", ctx, tx, database.Text("lesson-id-1"), database.Text("user-id-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(4).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.False(t, state.Recording.IsRecording)
						assert.Nil(t, state.Recording.Creator)
					}).
					Return(nil).Once()
			},
		},
		{
			name: "stop recording failed",
			command: &StopRecordingCommand{
				CommanderID: "user-id-1",
				LessonID:    "lesson-id-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRepo.On("StopRecording", ctx, tx, database.Text("lesson-id-1"), database.Text("user-id-1"), mock.Anything).
					Run(func(args mock.Arguments) {
						stateJsonb := args.Get(4).(pgtype.JSONB)
						state := &LessonRoomState{}
						err := stateJsonb.AssignTo(state)
						require.NoError(t, err)
						assert.False(t, state.Recording.IsRecording)
						assert.Nil(t, state.Recording.Creator)
					}).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &StopRecordingHandler{
				command:    tc.command,
				DB:         db,
				LessonRepo: lessonRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo)
		})
	}
}

func TestWhiteboardZoomStateCommandHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lessonRoomStateRepo := &mock_repo_lessonmgmt.MockLessonRoomStateRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	w := &domain.WhiteboardZoomState{
		PdfScaleRatio: 23.32,
		CenterX:       243.5,
		CenterY:       -432.034,
		PdfWidth:      234.43,
		PdfHeight:     -0.33424,
	}
	tcs := []struct {
		name     string
		command  *WhiteboardZoomStateCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute whiteboard zoom state user successfully",
			command: &WhiteboardZoomStateCommand{
				CommanderID:         "teacher-1",
				LessonID:            "lesson-1",
				WhiteboardZoomState: w,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("UpsertWhiteboardZoomState", ctx, tx, "lesson-1", w).
					Return(nil).Once()
			},
		},
		{
			name: "execute whiteboard zoom state user failed",
			command: &WhiteboardZoomStateCommand{
				CommanderID:         "teacher-1",
				LessonID:            "lesson-1",
				WhiteboardZoomState: w,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("UpsertWhiteboardZoomState", ctx, tx, "lesson-1", w).
					Return(pgx.ErrTxClosed).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &WhiteboardZoomStateCommandHandler{
				command:             tc.command,
				DB:                  db,
				LessonRoomStateRepo: lessonRoomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, tx, lessonRoomStateRepo)
		})
	}

}

func TestSpotlightCommandHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lessonRoomStateRepo := &mock_repo_lessonmgmt.MockLessonRoomStateRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	tcs := []struct {
		name     string
		command  *SpotlightCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute spotlight user successfully",
			command: &SpotlightCommand{
				CommanderID:     "teacher-1",
				LessonID:        "lesson-1",
				SpotlightedUser: "user-1",
				IsEnable:        true,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("Spotlight", ctx, tx, database.Text("lesson-1"), database.Text("user-1")).
					Return(nil).Once()
			},
		},
		{
			name: "execute spotlight user failed",
			command: &SpotlightCommand{
				CommanderID:     "teacher-1",
				LessonID:        "lesson-1",
				SpotlightedUser: "user-1",
				IsEnable:        true,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("Spotlight", ctx, tx, database.Text("lesson-1"), database.Text("user-1")).
					Return(pgx.ErrTxClosed).Once()
			},
			hasError: true,
		},
		{
			name: "execute unspotlight user successfully",
			command: &SpotlightCommand{
				CommanderID:     "teacher-1",
				LessonID:        "lesson-1",
				SpotlightedUser: "user-1",
				IsEnable:        false,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("UnSpotlight", ctx, tx, database.Text("lesson-1")).
					Return(nil).Once()
			},
		},
		{
			name: "execute unspotlight user failed",
			command: &SpotlightCommand{
				CommanderID:     "teacher-1",
				LessonID:        "lesson-1",
				SpotlightedUser: "user-1",
				IsEnable:        false,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("UnSpotlight", ctx, tx, database.Text("lesson-1")).
					Return(pgx.ErrTxClosed).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &SpotlightCommandHandler{
				command:             tc.command,
				DB:                  db,
				LessonRoomStateRepo: lessonRoomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, tx, lessonRoomStateRepo)
		})
	}

}

func TestUpdateChatCommandHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	lessonRepo := &mock_repositories.MockLessonRepo{}

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	learners := []string{"learner-1"}
	tcs := []struct {
		name     string
		command  *UpdateChatCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute enable chat command successfully",
			command: &UpdateChatCommand{
				CommanderID: "teacher-1",
				UserIDs:     learners,
				LessonID:    "lesson-1",
				State: &UserChat{
					Value: true,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeChat)),
						database.TextArray(learners),
						&entities.StateValue{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
			},
		},
		{
			name: "execute disable annotation command successfully",
			command: &UpdateChatCommand{
				CommanderID: "teacher-1",
				UserIDs:     learners,
				LessonID:    "lesson-1",
				State: &UserChat{
					Value: false,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						database.Text("lesson-1"),
						database.Text(string(LearnerStateTypeChat)),
						database.TextArray(learners),
						&entities.StateValue{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := &UpdateChatCommandHandler{
				command:          tc.command,
				DB:               db,
				LessonMemberRepo: lessonMemberRepo,
				LessonRepo:       lessonRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonMemberRepo, lessonRepo)
		})
	}
}
