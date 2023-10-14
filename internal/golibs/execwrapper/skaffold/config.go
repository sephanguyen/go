package skaffoldwrapper

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type EnvSet struct {
	Env                         string `env:"ENV"`
	Org                         string `env:"ORG"`
	APPSMITH_DEPLOYMENT_ENABLED string `env:"APPSMITH_DEPLOYMENT_ENABLED"` //nolint:revive,stylecheck
	CAMEL_K_ENABLED             string `env:"CAMEL_K_ENABLED"`             //nolint:revive,stylecheck
}

// Environ is similar to os.Environ, but also overrides
// environment variable outputs with its own values.
func (e *EnvSet) Environ() []string {
	res := os.Environ()

	val := reflect.ValueOf(e).Elem()
	typ := reflect.TypeOf(e).Elem()
	for i := 0; i < typ.NumField(); i++ {
		if envKey, ok := typ.Field(i).Tag.Lookup("env"); ok {
			envVal := val.Field(i).String()
			envKeyVal := fmt.Sprintf("%s=%s", envKey, envVal)
			res = append(res, envKeyVal)
		}
	}
	return res
}

func (e *EnvSet) EnvironMap() map[string]string {
	envList := e.Environ()
	res := make(map[string]string, len(envList))
	for _, v := range envList {
		ss := strings.SplitN(v, "=", 2) // TODO: is this correct?
		if len(ss) < 2 {
			panic(fmt.Errorf("invalid key-value environment set %q", v))
		}
		res[ss[0]] = ss[1]
	}
	return res
}

type FlagSet struct {
	F                     string // is -f or --filename
	P                     string // is -p or --profile
	ProfileAutoActivation *bool  // is --profile-auto-activation
}

// Args returns the list of arguments intended for skaffold commands.
func (f FlagSet) Args(extra ...string) []string {
	args := make([]string, 0, len(extra)+2)
	args = append(args, extra...)
	if f.F != "" {
		args = append(args, "-f", f.F)
	}
	if f.P != "" {
		args = append(args, "-p", f.P)
	}
	if f.ProfileAutoActivation != nil {
		args = append(args, "--profile-auto-activation="+strconv.FormatBool(*f.ProfileAutoActivation))
	}
	return args
}
