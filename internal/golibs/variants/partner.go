package vr

import "fmt"

// P, partner, is the name of our business partner for whom the service is running.
// Sometimes referred to as Organization or Vendor in other places.
// For a multitenant cluster, this is the name of the entire cluster (which comprises
// of multiple partners).
type P int

// List of all available partners.
const (
	PartnerNotDefined P = iota
	PartnerAIC
	PartnerGA
	PartnerJPREP
	PartnerManabie
	PartnerRenseikai
	PartnerSynersia
	PartnerTokyo // multitenant cluster
	PartnerE2E   // end-to-end test partner
)

var partnerToString = map[P]string{
	PartnerAIC:       "aic",
	PartnerGA:        "ga", // bestco?
	PartnerJPREP:     "jprep",
	PartnerManabie:   "manabie",
	PartnerRenseikai: "renseikai",
	PartnerSynersia:  "synersia",
	PartnerTokyo:     "tokyo",
	PartnerE2E:       "e2e",
}

// String implements the Stringer interface.
func (p P) String() string {
	s, ok := partnerToString[p]
	if !ok {
		panic(fmt.Errorf("invalid partner %d", p))
	}
	return s
}

// ToPartner returns the matching P from input s.
// It panics if s is invalid.
func ToPartner(s string) P {
	for k, v := range partnerToString {
		if v == s {
			return k
		}
	}
	panic(fmt.Errorf("invalid partner string %q", s))
}
