package vr

import "fmt"

// ServiceAccountPrefix returns the common prefix of the service account in a
// specific environment.
func ServiceAccountPrefix(p P, e E) string {
	switch e {
	case EnvStaging:
		switch p {
		case PartnerManabie:
			return "stag-"
		case PartnerJPREP:
			return "stag-jprep-"
		}
	case EnvUAT:
		return "uat-"
	case EnvProduction:
		if p == PartnerJPREP {
			return "prod-jprep-"
		}
		return "prod-"
	}

	panic(fmt.Sprintf("invalid partner-environment combination: %v-%v", p, e))
}

// ServiceAccountDBUser returns the database username for a service in a specific environment.
//
// Example: stag-bob@staging-manabie-online.iam
func ServiceAccountDBUser(p P, e E, s S) string {
	return ServiceAccountPrefix(p, e) + s.String() + "@" + GCPProjectID(p, e) + ".iam"
}

// ServiceAccountDBMigrationUser returns the migration database username for a service in a specific environment.
//
// Example: stag-bob-m@staging-manabie-online.iam
func ServiceAccountDBMigrationUser(p P, e E, s S) string {
	return ServiceAccountPrefix(p, e) + s.String() + "-m@" + GCPProjectID(p, e) + ".iam"
}

// ServiceAccountHasuraDBUser returns the database username of Hasura for a service in a specific environment.
//
// Example: stag-bob-h@staging-manabie-online.iam
func ServiceAccountHasuraDBUser(p P, e E, s S) string {
	return ServiceAccountPrefix(p, e) + s.String() + "-h@" + GCPProjectID(p, e) + ".iam"
}

// ServiceAccountEmail returns the service account email for a service in a specific environment.
//
// Example: stag-bob@staging-manabie-online.iam.gserviceaccount.com
func ServiceAccountEmail(p P, e E, s S) string {
	return ServiceAccountDBUser(p, e, s) + ".gserviceaccount.com"
}

// MigrationServiceAccountEmail returns the service account email used in database migration for a service in a specific environment.
//
// Example: stag-bob-m@staging-manabie-online.iam.gserviceaccount.com
func MigrationServiceAccountEmail(p P, e E, s S) string {
	return ServiceAccountDBMigrationUser(p, e, s) + ".gserviceaccount.com"
}
