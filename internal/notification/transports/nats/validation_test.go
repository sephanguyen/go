package nats

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/subscribers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_ValidateNatsMessage(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		natsNoti := utils.GenSampleNatsNotification()

		for _, clientId := range ClientIDsAccepted {
			natsNoti.ClientId = clientId
			err := ValidateNatsMessage(natsNoti)
			assert.Nil(t, err)
		}
	})

	t.Run("prevent case", func(t *testing.T) {
		t.Parallel()
		natsNoti := utils.GenSampleNatsNotification()
		natsNoti.ClientId = idutil.ULIDNow()
		err := ValidateNatsMessage(natsNoti)
		expectedErr := fmt.Errorf("prevent client_id: %s", natsNoti.ClientId)
		assert.Equal(t, expectedErr, err)
	})
}
