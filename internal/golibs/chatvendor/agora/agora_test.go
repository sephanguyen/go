package agora

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_NewAgoraClient(t *testing.T) {
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

		config := configs.AgoraConfig{
			AppID:              DataMockAppID,
			PrimaryCertificate: DataMockAppCertificate,
			AppName:            DataMockAppName,
			OrgName:            DataMockOrgName,
			RestAPI:            ts.URL,
		}

		_, err := NewAgoraClient(config, zap.NewExample())
		assert.Nil(t, err)
	})

	t.Run("missing app_id", func(t *testing.T) {
		t.Parallel()
		config := configs.AgoraConfig{
			AppID:              "",
			PrimaryCertificate: DataMockAppCertificate,
			OrgName:            DataMockOrgName,
			AppName:            DataMockAppName,
			RestAPI:            "example.com",
		}

		_, err := NewAgoraClient(config, zap.NewExample())
		assert.Equal(t, fmt.Errorf("[agora]: missing app_id config"), err)
	})

	t.Run("missing app_certificate", func(t *testing.T) {
		t.Parallel()
		config := configs.AgoraConfig{
			AppID:              DataMockAppID,
			PrimaryCertificate: "",
			OrgName:            DataMockOrgName,
			AppName:            DataMockAppName,
			RestAPI:            "example.com",
		}

		_, err := NewAgoraClient(config, zap.NewExample())
		assert.Equal(t, fmt.Errorf("[agora]: missing app_certificate config"), err)
	})

	t.Run("missing org_name", func(t *testing.T) {
		t.Parallel()
		config := configs.AgoraConfig{
			AppID:              DataMockAppID,
			PrimaryCertificate: DataMockAppCertificate,
			OrgName:            "",
			AppName:            DataMockAppName,
			RestAPI:            "example.com",
		}

		_, err := NewAgoraClient(config, zap.NewExample())
		assert.Equal(t, fmt.Errorf("[agora]: missing org_name config"), err)
	})

	t.Run("missing app_name", func(t *testing.T) {
		t.Parallel()
		config := configs.AgoraConfig{
			AppID:              DataMockAppID,
			PrimaryCertificate: DataMockAppCertificate,
			OrgName:            DataMockOrgName,
			AppName:            "",
			RestAPI:            "example.com",
		}

		_, err := NewAgoraClient(config, zap.NewExample())
		assert.Equal(t, fmt.Errorf("[agora]: missing app_name config"), err)
	})

	t.Run("missing REST API address config", func(t *testing.T) {
		t.Parallel()
		config := configs.AgoraConfig{
			AppID:              DataMockAppID,
			PrimaryCertificate: DataMockAppCertificate,
			OrgName:            DataMockOrgName,
			AppName:            DataMockAppName,
			RestAPI:            "",
		}

		_, err := NewAgoraClient(config, zap.NewExample())
		assert.Equal(t, fmt.Errorf("[agora]: missing REST API address config"), err)
	})
}
