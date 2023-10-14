package domain

type UserBasicInfo struct {
	UserID            string
	Name              string
	FirstName         string
	LastName          string
	FullNamePhonetic  string
	FirstNamePhonetic string
	LastNamePhonetic  string
	Email             string
}

func (u *UserBasicInfo) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_id",
		"name",
		"first_name",
		"last_name",
		"full_name_phonetic",
		"first_name_phonetic",
		"last_name_phonetic",
		"email",
	}
	values = []interface{}{
		&u.UserID,
		&u.Name,
		&u.FirstName,
		&u.LastName,
		&u.FullNamePhonetic,
		&u.FirstNamePhonetic,
		&u.LastNamePhonetic,
		&u.Email,
	}
	return
}
