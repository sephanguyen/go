package vr

import (
	"fmt"
	"os"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
)

// S, service, is the name of the service (business or platform).
type S int

// List of (almost) all available services.
const (
	ServiceNotDefined S = iota
	ServiceAuth
	ServiceBob
	ServiceEnigma
	ServiceEntryExitMgmt
	ServiceEureka
	ServiceFatima
	ServiceFink
	ServiceHephaestus
	ServiceInvoiceMgmt
	ServiceJerry
	ServiceLessonMgmt
	ServiceMasterMgmt
	ServiceNotificationMgmt
	ServicePayment
	ServiceShamir
	ServiceTimesheet
	ServiceTom
	ServiceUnleash
	ServiceUserMgmt
	ServiceVirtualClassroom
	ServiceYasuo
	ServiceZeus
	ServiceCalendar
	ServiceDiscount

	// test services
	ServiceDraft
	ServiceGandalf

	// client services
	ServiceBackoffice
	ServiceLearnerWeb
	ServiceTeacherWeb

	// platform services
	ServiceHasura
	ServiceKafkaConnect
	ServiceElasticsearch
	ServiceNATSJetstream
	ServiceConversationmgmt
	ServiceSpike
)

var serviceToString = map[S]string{
	ServiceAuth:             "auth",
	ServiceBob:              "bob",
	ServiceEnigma:           "enigma",
	ServiceEntryExitMgmt:    "entryexitmgmt",
	ServiceEureka:           "eureka",
	ServiceFatima:           "fatima",
	ServiceFink:             "fink",
	ServiceHephaestus:       "hephaestus",
	ServiceInvoiceMgmt:      "invoicemgmt",
	ServiceJerry:            "jerry",
	ServiceLessonMgmt:       "lessonmgmt",
	ServiceMasterMgmt:       "mastermgmt",
	ServiceNotificationMgmt: "notificationmgmt",
	ServicePayment:          "payment",
	ServiceShamir:           "shamir",
	ServiceTimesheet:        "timesheet",
	ServiceCalendar:         "calendar",
	ServiceTom:              "tom",
	ServiceUnleash:          "unleash",
	ServiceUserMgmt:         "usermgmt",
	ServiceVirtualClassroom: "virtualclassroom",
	ServiceYasuo:            "yasuo",
	ServiceZeus:             "zeus",
	ServiceConversationmgmt: "conversationmgmt",
	ServiceSpike:            "spike",
	ServiceDiscount:         "discount",

	ServiceDraft:   "draft",
	ServiceGandalf: "gandalf",

	ServiceBackoffice: "backoffice",
	ServiceLearnerWeb: "learner-web",
	ServiceTeacherWeb: "teacher-web",

	ServiceHasura:        "hasura",
	ServiceKafkaConnect:  "kafka-connect",
	ServiceElasticsearch: "elastic",
	ServiceNATSJetstream: "nats-jetstream",
}

// String implements the Stringer interface.
func (s S) String() string {
	str, ok := serviceToString[s]
	if !ok {
		panic(fmt.Errorf("invalid service %d (not found)", s))
	}
	return str
}

// ToService returns the matching S from input s.
// It panics if s is invalid.
func ToService(s string) S {
	res, err := ToServiceErr(s)
	if err != nil {
		panic(err)
	}
	return res
}

// IsService checks whether s is a valid service.
func IsService(s string) bool {
	_, err := ToServiceErr(s)
	return err == nil
}

// ToServiceErr is similar to ToServiceErr, but returns an
// error instead of panicking.
func ToServiceErr(s string) (S, error) {
	for k, v := range serviceToString {
		if v == s {
			return k, nil
		}
	}
	return ServiceNotDefined, fmt.Errorf("invalid service string %q", s)
}

type chartYAML struct {
	Dependencies []struct {
		Name string `yaml:"name"`
	} `yaml:"dependencies"`
}

var getStdServices = func() func() ([]S, error) {
	f := func() ([]S, error) {
		fp := execwrapper.Abs("deployments/helm/manabie-all-in-one/Chart.yaml")
		c := chartYAML{}
		if err := loadyaml(fp, &c); err != nil {
			return nil, err
		}
		res := make([]S, 0, len(c.Dependencies))
		for _, v := range c.Dependencies {
			svcName := v.Name

			// some services are ignored
			if svcName == "backoffice" || svcName == "learner-web" ||
				svcName == "teacher-web" || svcName == "util" || svcName == "gandalf" {
				continue
			}

			if !IsService(svcName) {
				return nil, fmt.Errorf("service %q found in Chart.yaml but not defined in this package", svcName)
			}
			res = append(res, ToService(svcName))
		}
		return res, nil
	}
	return f
}()

// StdServices returns a list containing all standard backend services.
func StdServices() []S {
	res, err := getStdServices()
	if err != nil {
		panic(err)
	}
	return res
}

var getBackendServices = func() func() []S {
	res := make([]S, 0, 5)
	var once sync.Once
	return func() []S {
		once.Do(func() {
			backendChartDir := execwrapper.Abs("deployments/helm/backend")
			dirEntries, err := os.ReadDir(backendChartDir)
			if err != nil {
				panic(err)
			}
			for _, d := range dirEntries {
				if !d.IsDir() {
					continue
				}
				svcName := d.Name()
				if svcName == "common" {
					continue
				}
				svc := ToService(svcName)
				res = append(res, svc)
			}
		})
		return res
	}
}()

// BackendServices returns the backend services located at `deployments/helm/backend`
// directory. Note that `common` chart is skipped.
func BackendServices() []S {
	return getBackendServices()
}
