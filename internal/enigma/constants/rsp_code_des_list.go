package constants

var (
	rspCodeMap = map[int64]map[int64]string{
		//Bank’s Response Code
		1: {
			1:    "Bank Decline",
			2:    "Bank Decline",
			3:    "Other",
			4:    "Other",
			5:    "Bank Decline",
			12:   "Other",
			13:   "Other",
			14:   "Input Error",
			19:   "Other",
			25:   "Other",
			30:   "Other",
			31:   "Other",
			41:   "Lost/Stolen Card",
			43:   "Lost/Stolen Card",
			51:   "Bank Decline",
			54:   "Input Error",
			55:   "Other",
			58:   "Other",
			76:   "Other",
			77:   "Other",
			78:   "Other",
			80:   "Other",
			89:   "Other",
			91:   "Other",
			94:   "Other",
			95:   "Other",
			96:   "Other",
			99:   "Other",
			2000: "Other",
		},
		//Response Code From PayDollar
		-8: {
			999:  "Other",
			1000: "Skipped transaction",
			2000: "Blacklist error",
			2001: "Blacklist card by system",
			2002: "Blacklist card by merchant",
			2003: "Black IP by system",
			2004: "Black IP by merchant",
			2005: "Invalid cardholder name",
			2006: "Same card used more than 6 times a day",
			2007: "Duplicate merchant reference no.",
			2008: "Empty merchant reference no.",
			2011: "Other",
			2012: "Card verification failed",
			2013: "Card already registered",
			2014: "High risk country",
			2016: "Same payer IP attempted more than pre-defined no. a day.",
			2017: "Invalid card number",
			2018: "Multi-card attempt",
			2019: "Issuing Bank not match",
			2020: "Single transaction limit exceeded",
			2021: "Daily transaction limit exceeded",
			2022: "Monthly transaction limit exceeded",
			2023: "Invalid channel type",
			2099: "Non testing card",
			2031: "System rejected(TN)",
			2032: "System rejected(TA)",
			2033: "System rejected(TR)",
		},
		//Other
		0: {
			0: "Success",
		},
		-1: {
			-1: "Input Parameter Error",
		},
		-2: {
			-2: "Server Access Error",
		},
		-9: {
			-9: "Host Access Error",
		},
	}
)

func GetDescriptionWithPrcAndSrc(prc int64, src int64) string {
	srcMap := rspCodeMap[prc]
	return srcMap[src]
}
