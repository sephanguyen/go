package constant

import "time"

const (
	UserGroupStudent             = "student"
	UserGroupAdmin               = "admin"
	UserGroupTeacher             = "teacher"
	UserGroupParent              = "parent"
	UserGroupSchoolAdmin         = "school admin"
	UserGroupOrganizationManager = "organization manager"
	UserGroupHQStaff             = "hq staff"
	UserGroupCentreManager       = "centre manager"
	UserGroupCentreStaff         = "centre staff"
	UserGroupUnauthenticated     = "unauthenticated"
	DefaultPageLimit             = 3

	// Sleep duration constant. You can tweak the durations here easily to improve or fix some scenarios that requires delay or sleep.
	KafkaSyncSleepDuration = 2000 * time.Millisecond
	DuplicateSleepDuration = 300 * time.Millisecond
	ReselectSleepDuration  = 300 * time.Millisecond

	InvoiceFileFolderUploadPath = "invoicemgmt-upload"
)
