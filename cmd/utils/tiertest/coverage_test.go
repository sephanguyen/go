package tiertest

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_countSecnariosInFolderWithTier(t *testing.T) {
	folder := "./testdata"
	type testCase struct {
		tier    Tier
		matched int
	}
	tcases := []testCase{
		{
			tier:    TierMinor,
			matched: 5,
		},
		{
			tier:    TierMajor,
			matched: 4,
		},
		{
			tier:    TierCritical,
			matched: 2,
		},
		{
			tier:    TierBlocker,
			matched: 8,
		},
	}
	for _, tcase := range tcases {
		total, matched, err := countScenariosInFolderWithTier(folder, tcase.tier)
		assert.NoError(t, err)
		assert.Equal(t, tcase.matched, matched)
		assert.Equal(t, 19, total)
	}
}

func Test_countScenarioWithTier(t *testing.T) {
	featFile := "./testdata/1.feature"
	file, err := os.Open(featFile)
	require.NoError(t, err)
	defer file.Close()
	bs, err := io.ReadAll(file)
	assert.NoError(t, err)
	type testCase struct {
		tier    Tier
		matched int
	}
	tcases := []testCase{
		{
			tier:    TierMinor,
			matched: 2,
		},
		{
			tier:    TierMajor,
			matched: 2,
		},
		{
			tier:    TierCritical,
			matched: 1,
		},
		{
			tier:    TierBlocker,
			matched: 4,
		},
	}
	for _, tcase := range tcases {
		total, matched, err := countScenariosWithTier(bytes.NewReader(bs), tcase.tier)
		assert.NoError(t, err)
		assert.Equal(t, tcase.matched, matched)
		assert.Equal(t, 9, total)
	}
}
