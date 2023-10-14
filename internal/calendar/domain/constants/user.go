package constants

type (
	StaffWorkingStatus string
)

const (
	Available StaffWorkingStatus = "AVAILABLE"
	OnLeave   StaffWorkingStatus = "ON_LEAVE"
	Resigned  StaffWorkingStatus = "RESIGNED"
)
