package elastic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/stretchr/testify/assert"
)

func AssertDocIsValid(t *testing.T, d Doc) {
	resourcePath := idutil.ULIDNow()

	mapAlias1 := makeMapFromAny(t, d.Inner)
	// marshaling given type must not have resource_path
	_, ok := mapAlias1["resource_path"]
	assert.False(t, ok)

	d._mandatory.ResourcePath = resourcePath
	mapAlias2 := makeMapFromAny(t, d)
	assert.Equal(t, resourcePath, mapAlias2["resource_path"])
	delete(mapAlias2, "resource_path")
	assert.True(t, reflect.DeepEqual(mapAlias1, mapAlias2))
}

func makeMapFromAny(t *testing.T, i interface{}) map[string]interface{} {
	bs, err := json.Marshal(i)
	assert.NoError(t, err)
	contMap := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(bs, &contMap))

	return contMap
}
