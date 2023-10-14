package agora

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	"github.com/manabie-com/backend/internal/golibs/try"
)

func (a *agoraClientImpl) doRequest(ctx context.Context, method Method, endpoint string, header map[string]string, body io.Reader, holder any) error {
	url := GetAgoraRESTAPI(a.AgoraConfig) + endpoint
	if body == nil {
		body = strings.NewReader("")
	}
	request, err := http.NewRequestWithContext(ctx, string(method), url, body)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	isHasToken := false
	for key, element := range header {
		request.Header.Set(key, element)
		if key == AuthHeaderKey {
			isHasToken = true
		}
	}

	if !isHasToken {
		appToken, err := a.GetAppToken()
		if err != nil {
			return fmt.Errorf("[agora]: cannot get app token: [%v]", err)
		}

		request.Header.Set(AuthHeaderKey, appToken)
	}

	if err := try.Do(func(attempt int) (bool, error) {
		res, err := a.httpClient.Do(request)
		if err != nil {
			return false, fmt.Errorf("[agora] cannot exec req: %w", err)
		}
		defer res.Body.Close()

		code := res.StatusCode

		errResStr := make([]byte, 0)
		if code != 200 {
			// get error response
			errRes := &dto.ErrorResponse{}
			_ = json.NewDecoder(res.Body).Decode(&errRes)
			errResStr, _ = json.Marshal(errRes)
		} else {
			// sucess -> assign to holder
			if holder != nil {
				err := json.NewDecoder(res.Body).Decode(&holder)
				if err != nil {
					return false, fmt.Errorf("json.Decode: %w", err)
				}
			}
			return false, nil
		}

		// unauthorized -> re-new token
		if code == 401 {
			appToken, err := a.GetAppToken()
			if err != nil {
				return false, fmt.Errorf("[agora]: cannot get app token: [%v]", err)
			}

			request.Header.Set(AuthHeaderKey, appToken)
			return attempt < 5, fmt.Errorf("[agora]: http request failed with status code: [%d], error: [%v]", res.StatusCode, string(errResStr))
		}

		// not found
		if code == 404 {
			return false, nil
		}

		return false, fmt.Errorf("[agora]: http request failed with status code: [%d], error: [%v]", res.StatusCode, string(errResStr))
	}); err != nil {
		return fmt.Errorf("try.Do: %w", err)
	}

	return nil
}
