package vr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type existence int

const (
	exNotFound existence = iota
	exFound
)

func TestIter_IterPE(t *testing.T) {
	t.Parallel()

	mapKeys := func(in map[E][]P) []E {
		res := make([]E, 0, len(in))
		for k := range in {
			res = append(res, k)
		}
		return res
	}

	compareMap := func(t *testing.T, expected, actual map[E][]P) {
		actualKeys := mapKeys(actual)
		expectedKeys := mapKeys(expected)
		require.ElementsMatch(t, expectedKeys, actualKeys)
		for k := range expected {
			require.ElementsMatch(t, expected[k], actual[k], "elements mismatched for environment %v", k)
		}
	}

	t.Run("default case", func(t *testing.T) {
		actual := make(map[E][]P, 5)
		Iter(t).DisableParallel().
			IterPE(func(_ *testing.T, p P, e E) { actual[e] = append(actual[e], p) })
		compareMap(t, PartnerListByEnv(), actual)
	})

	t.Run("for some environments only", func(t *testing.T) {
		actual := make(map[E][]P, 5)
		expected := make(map[E][]P, 5)
		Iter(t).DisableParallel().
			E(EnvLocal, EnvStaging).
			IterPE(func(_ *testing.T, p P, e E) { actual[e] = append(actual[e], p) })
		for k, v := range PartnerListByEnv() {
			if k == EnvLocal || k == EnvStaging {
				expected[k] = v
			}
		}
		compareMap(t, expected, actual)
	})

	t.Run("skip some environments", func(t *testing.T) {
		actual := make(map[E][]P, 5)
		expected := make(map[E][]P, 5)
		Iter(t).DisableParallel().
			SkipE(EnvLocal, EnvStaging).
			IterPE(func(_ *testing.T, p P, e E) { actual[e] = append(actual[e], p) })
		for k, v := range PartnerListByEnv() {
			if !(k == EnvLocal || k == EnvStaging) {
				expected[k] = v
			}
		}
		compareMap(t, expected, actual)
	})
}

func TestIter_IterPES(t *testing.T) {
	t.Parallel()

	t.Run("default case", func(t *testing.T) {
		actual := make([]S, 0, 20)
		Iter(t).DisableParallel().E(EnvProduction).P(PartnerTokyo).
			IterPES(func(_ *testing.T, _ P, _ E, s S) { actual = append(actual, s) })
		require.ElementsMatch(t, BackendServices(), actual)
	})

	t.Run("skip disabled services", func(t *testing.T) {
		actual := make([]S, 0, 20)
		Iter(t).DisableParallel().E(EnvProduction).P(PartnerTokyo).
			SkipDisabledServices().
			IterPES(func(_ *testing.T, _ P, _ E, s S) { actual = append(actual, s) })
		require.NotContains(t, actual, ServiceDraft)
	})
}
