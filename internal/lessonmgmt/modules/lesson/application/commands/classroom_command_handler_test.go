package commands

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClassroomCommandHandler_ImportClassroom(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	classroomRepo := new(mock_repositories.MockClassroomRepo)

	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		input    interface{}
		hasError bool
	}{
		{
			name: "import classroom successfully",
			input: &lpb.ImportClassroomRequest{
				Payload: []byte(fmt.Sprintf(`location_id,location_name,classroom_id,classroom_name,room_area,seat_capacity,remarks
					location-1,location-name-1,,classroom name 1,floor 1,24,
					location-2,location-name-2,classroom-id-1,teacher room 1,floor 1,30,teacher seat`)),
			},

			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				locationName := map[string]string{
					"location-1": "location-name-1",
					"location-2": "location-name-2",
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				masterDataRepo.On("CheckLocationByIDs", ctx, mock.Anything, []string{"location-1", "location-2"}, locationName).Once().Return(nil)
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1"}).Return(nil).Once()
				classroomRepo.On("UpsertClassrooms", ctx, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}

	handler := ClassroomCommandHandler{
		WrapperConnection: wrapperConnection,
		ClassroomRepo:     classroomRepo,
		MasterDataPort:    masterDataRepo,
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			res, err := handler.ImportClassroom(ctx, tc.input.(*lpb.ImportClassroomRequest))
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
				require.NoError(t, err)
				require.NotNil(t, res)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
