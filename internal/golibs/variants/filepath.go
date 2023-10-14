package vr

import (
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
)

// CommonConfigFilepath returns the absolute path to the
// common config file of service s.
func CommonConfigFilepath(s S) string {
	return execwrapper.Absf("deployments/helm/backend/%v/configs/%v.common.config.yaml", s, s)
}

// ConfigFilepath returns the absolute path to the config
// file of service s in environment e of partner p.
func ConfigFilepath(p P, e E, s S) string {
	return execwrapper.Absf("deployments/helm/backend/%v/configs/%v/%v/%v.config.yaml", s, p, e, s)
}

// SecretFilePath returns the absolute path to the encrypted
// secret file of service s in environment e of partner p.
func SecretFilePath(p P, e E, s S) string {
	return execwrapper.Absf("deployments/helm/backend/%v/secrets/%v/%v/%v.secrets.encrypted.yaml", s, p, e, s)
}

// MigrationSecretFilePath returns the absolute path to the
// encrypted migration secret file of service s in environment e of partner p.
//
// Note that it does not check whether service s should have
// any migration secrets.
func MigrationSecretFilePath(p P, e E, s S) string {
	return execwrapper.Absf("deployments/helm/backend/%v/secrets/%v/%v/%v_migrate.secrets.encrypted.yaml", s, p, e, s)
}
