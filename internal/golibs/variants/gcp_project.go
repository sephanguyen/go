package vr

import "fmt"

// GCPProjectID returns the ID of the GCP project for a partner/environment
// combination.
//
// The GCP project of a environment is the project
// in which the service accounts of that environment are defined.
func GCPProjectID(p P, e E) string {
	// nolint:goconst
	switch e {
	case EnvStaging:
		switch p {
		case PartnerManabie:
			return "staging-manabie-online"
		case PartnerJPREP:
			return "staging-manabie-online"
		}
	case EnvUAT:
		switch p {
		case PartnerManabie:
			return "uat-manabie"
		case PartnerJPREP:
			return "staging-manabie-online"
		}
	case EnvProduction:
		switch p {
		case PartnerAIC:
			return "production-aic"
		case PartnerGA:
			return "production-ga"
		case PartnerJPREP:
			return "student-coach-e1e95"
		case PartnerRenseikai:
			return "production-renseikai"
		case PartnerSynersia:
			return "synersia"
		case PartnerTokyo:
			return "student-coach-e1e95"
		}
	}

	panic(fmt.Sprintf("invalid partner-environment combination: %v-%v", p, e))
}
