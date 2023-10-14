package vr

// IsHasuraEnabled reports whether the input service has Hasura
// enabled based on the Terraform definition file.
func IsHasuraEnabled(p P, e E, s S) (bool, error) {
	if b, err := IsServiceEnabled(p, e, s); !b || err != nil {
		return b, err
	}

	enabledServices := []S{
		ServiceCalendar,
		ServiceEntryExitMgmt,
		ServiceEureka,
		ServiceFatima,
		ServiceInvoiceMgmt,
		ServiceLessonMgmt,
		ServiceMasterMgmt,
		ServiceTimesheet,
	}
	for _, v := range enabledServices {
		if v == s {
			return true, nil
		}
	}
	return false, nil
}
