package field

type Status int8

const (
	StatusUndefined = iota
	StatusNull
	StatusPresent
	StatusIgnored
)
