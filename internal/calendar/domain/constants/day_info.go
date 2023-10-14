package constants

type (
	DateInfoStatus string
)

const (
	TimeLayout string = "2006-01-02"

	None      DateInfoStatus = "none"
	Draft     DateInfoStatus = "draft"
	Published DateInfoStatus = "published"
)
