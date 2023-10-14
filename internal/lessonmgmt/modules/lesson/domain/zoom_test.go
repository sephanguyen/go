package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZoom_Empty(t *testing.T) {
	t.Run("set empty zoom", func(t *testing.T) {
		z := &Zoom{
			ZoomID:       "z1",
			ZoomLink:     "zl1",
			AccountID:    "a1",
			OccurrenceID: "oc1",
		}
		z.Empty()
		emptyS := ""
		require.Equal(t, emptyS, z.ZoomID)
		require.Equal(t, emptyS, z.ZoomLink)
		require.Equal(t, emptyS, z.AccountID)
		require.Equal(t, emptyS, z.OccurrenceID)
	})
}

func TestZoom_IsEmpty(t *testing.T) {
	t.Run("empty zoom id", func(t *testing.T) {
		z := &Zoom{}
		isEmpty := z.IsEmpty()
		require.True(t, isEmpty)
	})
	t.Run("nil pointer zoom", func(t *testing.T) {
		var z *Zoom
		isEmpty := z.IsEmpty()
		require.True(t, isEmpty)
	})
	t.Run("not empty zoom", func(t *testing.T) {
		z := &Zoom{
			ZoomID: "z1",
		}
		isEmpty := z.IsEmpty()
		require.False(t, isEmpty)
	})
}

func TestZoom_Validate(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		z := &Zoom{
			ZoomID:       "z1",
			ZoomLink:     "zl1",
			AccountID:    "a1",
			OccurrenceID: "oc1",
		}
		err := z.Validate()
		require.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		z := &Zoom{
			ZoomID:       "zid",
			ZoomLink:     "",
			AccountID:    "a1",
			OccurrenceID: "oc1",
		}
		err := z.Validate()
		require.Error(t, err)
	})
}
