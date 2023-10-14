package objects

import "github.com/manabie-com/backend/internal/usermgmt/pkg/field"

type AccountRepo struct {
}

type AccountAttribute struct {
	ID          field.String `json:"Id"`
	Name        field.String `json:"Name"`
	Description field.String `json:"Description"`
	Phone       field.String `json:"Phone"`
}

type Account struct {
	AccountAttribute
	CreatedDate    field.Time `json:"CreatedDate"`
	SystemModStamp field.Time `json:"SystemModStamp"`
}

func (a *Account) Attrs() []string {
	return []string{"Id", "Name", "Description", "Phone"}
}

func (a *Account) Object() string {
	return "Account"
}

func (a *Account) Objects() string {
	return "Accounts"
}
