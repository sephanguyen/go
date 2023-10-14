package external

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/cerebry"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestNewCerebryRepo(t *testing.T) {
	t.Parallel()

	t.Run("should set base url and header for resty client", func(t *testing.T) {
		// arrange
		config := cerebry.Config{
			BaseURL:        "https://abc.com",
			PermanentToken: "TOKEN_A",
		}
		expectedHeaders := map[string]string{
			"Content-Type": "application/json",
			"jwt-token":    config.PermanentToken,
		}

		// act
		actual := NewCerebryRepo(config)

		// assert
		assert.Equal(t, "https://abc.com", actual.Client.BaseURL)
		assert.Equal(t, expectedHeaders["Content-Type"], actual.Client.Header.Get("Content-Type"))
		assert.Equal(t, expectedHeaders["jwt-token"], actual.Client.Header.Get("jwt-token"))
	})
}

func TestCerebryRepo_GetUserToken(t *testing.T) {
	t.Parallel()
	cfg := cerebry.Config{
		BaseURL:        "https://manabie.com",
		PermanentToken: "ABC",
	}
	rc := resty.New()
	rc.SetBaseURL(cfg.BaseURL)

	userID := idutil.ULIDNow()
	endpointURLFmt := "https://manabie.com/api/v4/partner/user/%s/token/"
	expectedURL := fmt.Sprintf(endpointURLFmt, userID)
	parentCtx := context.Background()
	ctx := context.WithValue(parentCtx, interceptors.UserIDKey(0), userID)

	t.Run("must call correct url and return token", func(t *testing.T) {
		// arrange
		httpmock.ActivateNonDefault(rc.GetClient())
		defer httpmock.DeactivateAndReset()
		sut := CerebryRepo{
			Config: cfg,
			Client: rc,
		}

		mockResponse, _ := httpmock.NewJsonResponder(200, map[string]any{
			"token": "abc",
		})
		httpmock.RegisterResponder("GET", expectedURL, mockResponse)

		// act
		tok, err := sut.GetUserToken(ctx, userID)

		// assert
		assert.Equal(t, "abc", tok)
		assert.Nil(t, err)

		// clean
		httpmock.Deactivate()
	})

	t.Run("return error when cerebry API occur an error", func(t *testing.T) {
		// arrange
		httpmock.ActivateNonDefault(rc.GetClient())
		defer httpmock.DeactivateAndReset()
		sut := CerebryRepo{
			Config: cfg,
			Client: rc,
		}

		mockResponse, _ := httpmock.NewJsonResponder(300, map[string]any{
			"detail": "something gone wrong",
		})
		httpmock.RegisterResponder("GET", expectedURL, mockResponse)
		expectedErr := errors.New("CerebryRepo.GetUserToken: API error", nil)

		// act
		tok, err := sut.GetUserToken(ctx, userID)

		// assert
		assert.Empty(t, tok)
		assert.Equal(t, expectedErr, err)

		// clean
		httpmock.Deactivate()
	})

	t.Run("return ErrAPIRespondNotFound when cerebry API return 404", func(t *testing.T) {
		// arrange
		httpmock.ActivateNonDefault(rc.GetClient())
		defer httpmock.DeactivateAndReset()
		sut := CerebryRepo{
			Config: cfg,
			Client: rc,
		}

		mockResponse, _ := httpmock.NewJsonResponder(404, map[string]any{
			"detail": "err not found",
		})
		httpmock.RegisterResponder("GET", expectedURL, mockResponse)
		expectedErr := errors.NewAppError(errors.ErrAPIRespondNotFound, "CerebryRepo.GetUserToken: User is not registered with Cerebry", nil)

		// act
		tok, err := sut.GetUserToken(ctx, userID)

		// assert
		assert.Empty(t, tok)
		assert.Equal(t, expectedErr, err)

		// clean
		httpmock.Deactivate()
	})

	t.Run("return error when body response does not contain token", func(t *testing.T) {
		// arrange
		httpmock.ActivateNonDefault(rc.GetClient())
		defer httpmock.DeactivateAndReset()
		sut := CerebryRepo{
			Config: cfg,
			Client: rc,
		}

		mockResponse, _ := httpmock.NewJsonResponder(200, map[string]any{
			"propX": "something gone wrong",
		})
		httpmock.RegisterResponder("GET", expectedURL, mockResponse)
		expectedErr := errors.New("CerebryRepo.GetUserToken: API does not return a token", nil)

		// act
		tok, err := sut.GetUserToken(ctx, userID)

		// assert
		assert.Empty(t, tok)
		assert.Equal(t, expectedErr, err)

		// clean
		httpmock.Deactivate()
	})

	t.Run("return conversion error when wrong body response", func(t *testing.T) {
		// arrange
		httpmock.ActivateNonDefault(rc.GetClient())
		defer httpmock.DeactivateAndReset()
		sut := CerebryRepo{
			Config: cfg,
			Client: rc,
		}

		mockResponse, _ := httpmock.NewJsonResponder(200, `{""`)
		httpmock.RegisterResponder("GET", expectedURL, mockResponse)

		// act
		tok, err := sut.GetUserToken(ctx, userID)

		// assert
		assert.Empty(t, tok)
		assert.Equal(t, err.(*errors.AppError).Key, errors.ErrConversion)

		// clean
		httpmock.Deactivate()
	})
}
