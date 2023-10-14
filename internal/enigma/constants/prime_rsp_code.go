package constants

var (
	primeRspCodeMap = map[int64]string{
		0:  "Sucess",
		1:  "Rejected by Payment Bank",
		3:  "Rejected due to Payer Authentication Failure (3D)",
		-1: "Rejected due to Input Parameters Incorrect",
		-2: "Rejected due to Server Access Error",
		-8: "Rejected due to PayDollar Internal/Fraud Prevention Checking",
		-9: "Rejected by Host Access Error",
	}
)

func GetDescriptionWithPrc(prc int64) string {
	return primeRspCodeMap[prc]
}
