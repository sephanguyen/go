package vr

import "fmt"

// DatabaseNamePrefix returns the prefix of databases in a specific environment.
//
// Example: for prod.tokyo, prefix is "tokyo_" -> database names are "tokyo_bob", "tokyo_eureka", etc...
func DatabaseNamePrefix(p P, e E) string {
	switch p {
	case PartnerGA, PartnerAIC, PartnerTokyo:
		if e == EnvProduction {
			return p.String() + "_"
		}
	case PartnerJPREP:
		switch e {
		case EnvStaging:
			return "stag_"
		case EnvUAT, EnvProduction:
			return ""
		}
	case PartnerE2E:
		if e == EnvLocal {
			return ""
		}
	case PartnerManabie:
		switch e {
		case EnvLocal:
			return ""
		case EnvStaging:
			return ""
		case EnvUAT:
			return "uat_"
		}
	case PartnerRenseikai:
		if e == EnvProduction {
			return ""
		}
	case PartnerSynersia:
		if e == EnvProduction {
			return PartnerTokyo.String() + "_" // see https://manabie.atlassian.net/browse/LT-13324
		}
	}
	panic(fmt.Sprintf("invalid partner-environment combination: %s-%s", p, e))
}

// DatabaseName returns the full database name in a specific environment.
//
// Example: for bob: tokyo_bob in prod.tokyo, or uat_bob in uat.manabie.
func DatabaseName(p P, e E, s S) string {
	return DatabaseNamePrefix(p, e) + s.String()
}
