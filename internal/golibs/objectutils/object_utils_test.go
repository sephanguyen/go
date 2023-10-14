package objectutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TStruct struct {
	Name     string
	Age      int
	children []string
}

func TestSafeGetObject(t *testing.T) {
	t.Parallel()

	t.Run("should not throw error when get nil struct", func(t *testing.T) {
		t.Parallel()
		expectVal := &TStruct{}
		funcReturnNil := func() *TStruct {
			return nil
		}
		safeObject := SafeGetObject(funcReturnNil)
		assert.Equal(t, expectVal.Name, safeObject.Name)
		assert.Equal(t, expectVal.children, safeObject.children)
		assert.Equal(t, expectVal.Age, safeObject.Age)

	})

	t.Run("should not throw error when get struct have value", func(t *testing.T) {
		t.Parallel()
		expectVal := &TStruct{Name: "John", Age: 3, children: []string{"mary"}}
		funcReturn := func() *TStruct {
			return expectVal
		}
		safeObject := SafeGetObject(funcReturn)
		assert.Equal(t, expectVal.Name, safeObject.Name)
		assert.Equal(t, expectVal.children[0], safeObject.children[0])
		assert.Equal(t, expectVal.Age, safeObject.Age)

	})

}
