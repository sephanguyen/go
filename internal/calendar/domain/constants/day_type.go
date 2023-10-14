package constants

type (
	DateTypeID string
)

const (
	RegularDay  DateTypeID = "regular"
	SeasonalDay DateTypeID = "seasonal"
	SpareDay    DateTypeID = "spare"
	ClosedDay   DateTypeID = "closed"
)
