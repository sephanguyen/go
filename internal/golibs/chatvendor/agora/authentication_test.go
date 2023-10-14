package agora

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetAppToken(t *testing.T) {
	agoraClient := newAgoraClientForUnitTest("")
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		token, err := agoraClient.GetAppToken()
		assert.Equal(t, nil, err)
		assert.NotEmpty(t, token)
	})
}

func Test_GetUserToken(t *testing.T) {
	agoraClient := newAgoraClientForUnitTest("")
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		token, _, err := agoraClient.GetUserToken(DataMockUserUUID)
		assert.Equal(t, err, nil)
		assert.NotEmpty(t, token)
	})
}

func Test_GetAppKey(t *testing.T) {
	agoraClient := newAgoraClientForUnitTest("")
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		expectedAppKey := DataMockOrgName + "#" + DataMockAppName
		appKey := agoraClient.GetAppKey()
		assert.Equal(t, expectedAppKey, appKey)
	})
}
