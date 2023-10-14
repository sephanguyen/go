package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockLocationIDRequest struct {
	locationIds []string
}

func (m *mockLocationIDRequest) GetLocationIds() []string {
	return m.locationIds
}

func TestLocationReaderService_NewLocationRestricted(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	methods := map[string]struct{}{
		"/master/methodWillBeCheck": {},
	}
	db := new(mock_database.Ext)
	locationRepo := new(mock_repositories.MockLocationRepo)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	interceptor := NewLocationRestricted(methods, db, locationRepo)
	t.Run("Test when method is not restricted", func(t *testing.T) {
		req := &mockLocationIDRequest{}
		info := &grpc.UnaryServerInfo{
			FullMethod: "/HeheService/OtherMethod",
		}

		resp, err := interceptor.UnaryServerInterceptor(ctx, req, info, handler)

		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	})

	t.Run("when method is restricted and no location IDs are sent", func(t *testing.T) {
		req := &mockLocationIDRequest{
			locationIds: []string{},
		}
		info := &grpc.UnaryServerInfo{
			FullMethod: "/master/methodWillBeCheck",
		}

		locationRepo.On("GetRootLocation", ctx, db).Return("", nil).Once()

		_, err := interceptor.UnaryServerInterceptor(ctx, req, info, handler)

		assert.Equal(t, "permission denied: user is not granted org level", err.Error())
	})

	t.Run("when method is restricted and locations invalid", func(t *testing.T) {
		req := &mockLocationIDRequest{
			locationIds: []string{"loc-1", "loc-2"},
		}
		info := &grpc.UnaryServerInfo{
			FullMethod: "/master/methodWillBeCheck",
		}

		locationRepo.On("GetLocationsByLocationIDs", ctx, db, database.TextArray(req.locationIds), false).Once().Return(
			[]*domain.Location{
				{
					LocationID: "loc-1",
					Name:       "name-1",
				},
			}, nil,
		)
		_, err := interceptor.UnaryServerInterceptor(ctx, req, info, handler)

		assert.Equal(t, "permission denied: some locations of [loc-1 loc-2] are not granted for user", err.Error())
	})

	t.Run("when method is restricted and locations is valid", func(t *testing.T) {
		req := &mockLocationIDRequest{
			locationIds: []string{"loc-1", "loc-2"},
		}
		info := &grpc.UnaryServerInfo{
			FullMethod: "/master/methodWillBeCheck",
		}

		locationRepo.On("GetLocationsByLocationIDs", ctx, db, database.TextArray(req.locationIds), false).Once().Return(
			[]*domain.Location{
				{
					LocationID: "loc-1",
					Name:       "name-1",
				},
				{
					LocationID: "loc-2",
					Name:       "name-2",
				},
			}, nil,
		)
		resp, err := interceptor.UnaryServerInterceptor(ctx, req, info, handler)

		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	})

}
