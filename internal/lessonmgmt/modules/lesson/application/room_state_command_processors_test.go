package application

import (
	"context"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_user_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/usermodadapter"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoomStateCommandPermissionChecker_Execute(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userModuleAdapter := &mock_user_module_adapter.MockUserModuleAdapter{}
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	tcs := []struct {
		name     string
		command  StateModifyCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "teacher execute share material command",
			command: &ModifyCurrentMaterialCommand{
				CommanderID: "teacher-2",
			},
			setup: func(ctx context.Context) {
				userModuleAdapter.On("GetUserGroup", ctx, "teacher-2").
					Return(entities_bob.UserGroupTeacher, nil).Once()
			},
			hasError: false,
		},
		{
			name: "learner execute share material command",
			command: &ModifyCurrentMaterialCommand{
				CommanderID: "learner-1",
			},
			setup: func(ctx context.Context) {
				userModuleAdapter.On("GetUserGroup", ctx, "learner-1").
					Return(entities_bob.UserGroupStudent, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			checker := RoomStateCommandPermissionChecker{
				WrapperConnection: wrapperConnection,
				UserModule:        userModuleAdapter,
			}
			err := checker.Execute(ctx, tc.command)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, userModuleAdapter, mockUnleashClient)
		})
	}
}
