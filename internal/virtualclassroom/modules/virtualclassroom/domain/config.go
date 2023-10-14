package domain

import "time"

type Country string

const (
	CountryNone   Country = "COUNTRY_NONE"
	CountryMaster Country = "COUNTRY_MASTER"
	CountryVN     Country = "COUNTRY_VN"
	CountryID     Country = "COUNTRY_ID"
	CountrySG     Country = "COUNTRY_SG"
	CountryJP     Country = "COUNTRY_JP"
)

type Config struct {
	Key       string
	Group     string
	Country   Country
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
