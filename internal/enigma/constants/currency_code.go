package constants

var (
	currencyCodeMap = map[string]string{
		"344": "HKD",
		"840": "USD",
		"702": "SGD",
		"156": "CNY",
		"392": "JPY",
		"901": "TWD",
		"036": "AUD",
		"978": "EUR",
		"826": "GBP",
		"124": "CAD",
		"446": "MOP",
		"608": "PHP",
		"764": "THB",
		"458": "MYR",
		"360": "IDR",
		"410": "KRW",
		"682": "SAR",
		"554": "NZD",
		"784": "AED",
		"096": "BND",
		"704": "VND",
		"356": "INR",
	}
)

func GetCurrency(cur string) string {
	return currencyCodeMap[cur]
}
