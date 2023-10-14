package entity

type User struct {
	ID       string
	Group    string
	Name     string
	Email    string
	Password string
	Token    string
}

type SchoolAdmin User
type Teacher User
type Student User
type Parent User
type HQStaff User
