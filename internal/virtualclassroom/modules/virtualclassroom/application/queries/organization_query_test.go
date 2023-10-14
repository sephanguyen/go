package queries

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrganizationQuery_GetOrganizationMap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	orgRepo := &mock_virtual_repo.MockOrganizationRepo{}

	t.Run("success", func(t *testing.T) {
		query := OrganizationQuery{
			WrapperDBConnection: wrapperConnection,
			OrganizationRepo:    orgRepo,
		}
		ids := []string{"id-1", "id-2", "id-3"}
		expectedResult := map[string]string{
			"000": ids[0],
			"001": ids[1],
			"002": ids[2],
		}
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		orgRepo.On("GetIDs", ctx, db).Once().Return(ids, nil)
		result, err := query.GetOrganizationMap(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("fail with err in repo", func(t *testing.T) {
		query := OrganizationQuery{
			WrapperDBConnection: wrapperConnection,
			OrganizationRepo:    orgRepo,
		}
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		orgRepo.On("GetIDs", ctx, db).Once().Return(nil, errors.New("error"))
		result, err := query.GetOrganizationMap(ctx)
		assert.Nil(t, result)
		assert.Equal(t, fmt.Errorf("error in OrganizationRepo.GetIDs: %w", errors.New("error")), err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}
