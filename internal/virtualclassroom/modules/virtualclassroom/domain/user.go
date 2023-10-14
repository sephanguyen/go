package domain

type User struct {
	ID        string
	Name      string
	Avatar    string
	UserGroup string
	Country   string
}

type UserBasicInfo struct {
	UserID            string
	Name              string
	FirstName         string
	LastName          string
	FullNamePhonetic  string
	FirstNamePhonetic string
	LastNamePhonetic  string
}
