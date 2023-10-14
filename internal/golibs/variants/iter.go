package vr

import (
	"fmt"
	"sync"
	"testing"
)

// EP represents one combination of E and P.
type EP struct {
	E
	P
}

// AllEP returns all EP combinations available in manaverse. It should
// return the same result as PartnerListByEnv, only in a different format.
func AllEP() []EP {
	return allEP()
}

var allEP = func() func() []EP {
	res := make([]EP, 0, 20)
	var once sync.Once
	return func() []EP {
		once.Do(func() {
			for e, plist := range PartnerListByEnv() {
				for _, p := range plist {
					res = append(res, EP{E: e, P: p})
				}
			}
		})
		return res
	}
}()

// EPS represents one combination of E, P, and S.
type EPS struct {
	E
	P
	S
}

// AllEP returns all EP combinations available in manaverse. It should
// return the processed result from PartnerListByEnv and BackendServices.
func AllEPS() []EPS {
	return allEPS()
}

var allEPS = func() func() []EPS {
	res := make([]EPS, 0, 20*20)
	var once sync.Once
	return func() []EPS {
		once.Do(func() {
			for _, ep := range AllEP() {
				for _, s := range BackendServices() {
					res = append(res, EPS{E: ep.E, P: ep.P, S: s})
				}
			}
		})
		return res
	}
}()

type Iterator struct {
	t *testing.T

	// test configs
	parallel bool

	// P/E/S iteration configs
	e                   []E
	p                   []P
	skipE               []E
	skipP               []P
	skipDisableServices bool
}

// Iter returns an iterator object whose IterPE or IterPES method can be used
// to easily iterate through all available P/E (and maybe S) combinations.
func Iter(t *testing.T) *Iterator {
	return &Iterator{
		t: t,

		// default parameters
		parallel:            true,
		skipDisableServices: false,
	}
}

// DisableParallel disables calling t.Parallel() on subtests.
func (it *Iterator) DisableParallel() *Iterator {
	it.parallel = false
	return it
}

// SkipE skips all input environments when iterating.
func (it *Iterator) SkipE(e ...E) *Iterator {
	if len(it.e) > 0 {
		panic("cannot use SkipE along with E")
	}
	it.skipE = e
	return it
}

// SkipP skips all input partners when iterating.
func (it *Iterator) SkipP(p ...P) *Iterator {
	if len(it.p) > 0 {
		panic("cannot use SkipP along with P")
	}
	it.skipP = p
	return it
}

// E makes it iterate through only input environments.
func (it *Iterator) E(e ...E) *Iterator {
	if len(it.skipE) > 0 {
		panic("cannot use E along with SkipE")
	}
	it.e = e
	return it
}

// P makes it iterate through only input partners.
func (it *Iterator) P(p ...P) *Iterator {
	if len(it.skipP) > 0 {
		panic("cannot use P along with SkipP")
	}
	it.p = p
	return it
}

// SkipDisabledServices skips iterating through P/E/S where that service is disabled.
func (it *Iterator) SkipDisabledServices() *Iterator {
	it.skipDisableServices = true
	return it
}

// shouldSkip returns true if val is:
//   - not in ignored, and
//   - in allowed, or if allowed is empty
func shouldSkip[T comparable](allowed, ignored []T, val T) bool {
	for _, v := range ignored {
		if v == val {
			return false
		}
	}
	if len(allowed) == 0 {
		return true
	}
	for _, v := range allowed {
		if v == val {
			return true
		}
	}
	return false
}

// ep returns the list of PE. Use E or SkipE to include or remove environments,
// and similarly, P/SkipP for partners.
func (it *Iterator) ep() []EP {
	res := make([]EP, 0, 18)

	for _, ep := range AllEP() {
		if !shouldSkip(it.e, it.skipE, ep.E) {
			continue
		}
		if !shouldSkip(it.p, it.skipP, ep.P) {
			continue
		}
		res = append(res, ep)
	}
	return res
}

// IterPE runs f for all iterable P/E combinations.
// The P/E combination can be adjusted with other configuration methods.
//
// By default, t.Parallel() is used (unless it.WithParallel is used otherwise).
func (it *Iterator) IterPE(f func(*testing.T, P, E)) {
	for _, pe := range it.ep() {
		p := pe.P
		e := pe.E
		it.t.Run(fmt.Sprintf("%v.%v", p, e), func(t *testing.T) {
			if it.parallel {
				t.Parallel()
			}
			f(t, p, e)
		})
	}
}

// IterPES runs f for all iterable P/E/S combinations.
// The P/E/S combination can be adjusted with other configuration methods.
//
// By default, t.Parallel() is used (unless it.WithParallel is used otherwise).
func (it *Iterator) IterPES(f func(*testing.T, P, E, S)) {
	for _, pe := range it.ep() {
		p := pe.P
		e := pe.E
		for _, s := range BackendServices() {
			s := s
			if it.skipDisableServices && !IsBackendServiceEnabled(p, e, s) {
				continue
			}
			it.t.Run(fmt.Sprintf("%v.%v.%v", p, e, s), func(t *testing.T) {
				if it.parallel {
					t.Parallel()
				}
				f(t, p, e, s)
			})
		}
	}
}
