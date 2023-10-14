package vr

import (
	"fmt"
)

// E, environment, indicates whether the cluster is for staging, uat, or production.
type E int

// List of all available environments.
// Preproduction is not included, because its config is always cloned from production's configs.l
const (
	EnvNotDefined E = iota
	EnvLocal
	EnvStaging
	EnvUAT
	EnvProduction
	EnvPreproduction
)

var envToString = map[E]string{
	EnvLocal:         "local",
	EnvStaging:       "stag",
	EnvUAT:           "uat",
	EnvProduction:    "prod",
	EnvPreproduction: "dorp",
}

// String implememts the Stringer interface.
func (e E) String() string {
	s, ok := envToString[e]
	if !ok {
		panic(fmt.Errorf("invalid environment %d", e))
	}
	return s
}

// ToEnv returns the matching E from input s.
// It panics if s is invalid.
func ToEnv(s string) E {
	for k, v := range envToString {
		if v == s {
			return k
		}
	}
	panic(fmt.Errorf("invalid environment string %q", s))
}

// PartnerListByEnv returns a map listing the partners of every environment.
func PartnerListByEnv() map[E][]P {
	return map[E][]P{
		EnvLocal:         {PartnerManabie, PartnerE2E},
		EnvStaging:       {PartnerManabie, PartnerJPREP},
		EnvUAT:           {PartnerManabie, PartnerJPREP},
		EnvPreproduction: {PartnerAIC, PartnerGA, PartnerJPREP, PartnerRenseikai, PartnerSynersia, PartnerTokyo},
		EnvProduction:    {PartnerAIC, PartnerGA, PartnerJPREP, PartnerRenseikai, PartnerSynersia, PartnerTokyo},
	}
}
