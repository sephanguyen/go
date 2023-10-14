package entities

import (
	"testing"

	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/stretchr/testify/assert"
)

func TestLearningMaterial_SetDefaultVendorType(t *testing.T) {
	t.Run("should set vendor type is MANABIE", func(t *testing.T) {
		lm := LearningMaterial{}
		lm.SetDefaultVendorType()

		assert.Equal(t, sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String(), lm.VendorType.Get())
	})
}
