package agora

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/entities"

	"github.com/stretchr/testify/assert"
)

func Test_GetAppInfo(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(dto.GetAgoraAppInfoResponse{
				Duration:  1,
				Timestamp: 1,
				Entities: []entities.AppInfo{
					{
						ApplicationName:  DataMockAppName,
						OrganizationName: DataMockOrgName,
					},
				},
			})
		}))
		defer ts.Close()

		agoraClient := newAgoraClientForUnitTest(ts.URL)
		orgName, appName, err := agoraClient.getAgoraAppInfo()
		assert.Equal(t, DataMockOrgName, orgName)
		assert.Equal(t, DataMockAppName, appName)
		assert.Nil(t, err)
	})
}
