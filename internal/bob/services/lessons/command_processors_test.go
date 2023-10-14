package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommandPermissionChecker(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := &mock_repositories.MockUserRepo{}
	db := &mock_database.Ext{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)
	tcs := []struct {
		name     string
		command  StateModifyCommand
		lesson   *entities.Lesson
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "teacher execute share material command",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "learner execute share material command",
			command: &ShareMaterialCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher who not belong to lesson execute share material command",
			command: &ShareMaterialCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "teacher execute stop sharing material command",
			command: &StopSharingMaterialCommand{
				CommanderID: "teacher-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "learner execute stop sharing material command",
			command: &StopSharingMaterialCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher who not belong to lesson execute stop sharing material command",
			command: &StopSharingMaterialCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "teacher execute fold hand all command",
			command: &FoldHandAllCommand{
				CommanderID: "teacher-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "learner execute fold hand all command",
			command: &FoldHandAllCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher who not belong to lesson execute fold hand all command",
			command: &FoldHandAllCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "teacher execute update hands up command",
			command: &UpdateHandsUpCommand{
				CommanderID: "teacher-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute update hands up command",
			command: &UpdateHandsUpCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner execute update hands up command for to update self-state",
			command: &UpdateHandsUpCommand{
				CommanderID: "learner-1",
				UserID:      "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "learner execute update hands up command for to update other learner's state",
			command: &UpdateHandsUpCommand{
				CommanderID: "learner-1",
				UserID:      "learner-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "learner who not belong to lesson execute update hands up command",
			command: &UpdateHandsUpCommand{
				CommanderID: "learner-5",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-5")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher execute reset all room's state",
			command: &ResetAllStatesCommand{
				CommanderID: "teacher-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "learner execute reset all room's state",
			command: &ResetAllStatesCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher who not belong to lesson execute reset all room's state",
			command: &ResetAllStatesCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "teacher execute update annotation command",
			command: &UpdateAnnotationCommand{
				CommanderID: "teacher-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
				RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute update annotation command",
			command: &UpdateAnnotationCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
				RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner try execute update annotation command",
			command: &UpdateAnnotationCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher execute start polling command",
			command: &StartPollingCommand{
				CommanderID: "teacher-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute start polling command",
			command: &StartPollingCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
				RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner try execute start polling command",
			command: &StartPollingCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher execute stop polling command",
			command: &StopPollingCommand{
				CommanderID: "teacher-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute stop polling command",
			command: &StopPollingCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
				RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner try execute stop polling command",
			command: &StopPollingCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher execute end polling command",
			command: &EndPollingCommand{
				CommanderID: "teacher-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute end polling command",
			command: &EndPollingCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
				RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner try execute end polling command",
			command: &EndPollingCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher execute submit polling answer command",
			command: &SubmitPollingAnswerCommand{
				CommanderID: "teacher-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "learner try execute submit polling answer polling command",
			command: &SubmitPollingAnswerCommand{
				CommanderID: "learner-1",
				UserID:      "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "learner try execute submit polling answer polling command of other learner's",
			command: &SubmitPollingAnswerCommand{
				CommanderID: "learner-1",
				UserID:      "learner-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "teacher execute request recording command",
			command: &RequestRecordingCommand{
				CommanderID: "teacher-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute request recording command",
			command: &RequestRecordingCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner try execute request recording command",
			command: &RequestRecordingCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher execute stop recording command",
			command: &StopRecordingCommand{
				CommanderID: "teacher-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute stop recording command",
			command: &StopRecordingCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner try execute stop recording command",
			command: &StopRecordingCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name: "teacher execute update chat command",
			command: &UpdateChatCommand{
				CommanderID: "teacher-2",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
				RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "teacher who not belong to lesson execute update chat command",
			command: &UpdateChatCommand{
				CommanderID: "teacher-3",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
				RoomState: database.JSONB(`
				{
					"current_material": {
						"media_id": "media-1",
						"updated_at": "` + string(nowString) + `"
					}
				}`),
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("teacher-3")).
					Return(entities.UserGroupTeacher, nil).
					Once()
			},
		},
		{
			name: "learner try execute update chat command",
			command: &UpdateChatCommand{
				CommanderID: "learner-1",
			},
			lesson: &entities.Lesson{
				LessonID: database.Text("lesson-1"),
				TeacherIDs: entities.TeacherIDs{
					TeacherIDs: database.TextArray([]string{"teacher-1", "teacher-2"}),
				},
				LearnerIDs: entities.LearnerIDs{
					LearnerIDs: database.TextArray([]string{"learner-1", "learner-2", "learner-3"}),
				},
			},
			setup: func(ctx context.Context) {
				userRepo.
					On("UserGroup", ctx, db, database.Text("learner-1")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			checker := CommandPermissionChecker{
				lesson:   tc.lesson,
				DB:       db,
				UserRepo: userRepo,
			}
			err := checker.Execute(ctx, tc.command)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, userRepo)
		})
	}
}
