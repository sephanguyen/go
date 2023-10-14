package auth

import "github.com/manabie-com/backend/internal/golibs/constants"

// LocalTenants map tenant id with school id
var LocalTenants = map[int]string{
	constants.ManabieSchool:      "manabie-0nl6t",
	constants.RenseikaiSchool:    "renseikai-yu9y7",
	constants.SynersiaSchool:     "synersia-24rue",
	constants.TestingSchool:      "end-to-end-dopvo",
	constants.GASchool:           "ga-school-jhe90",
	constants.KECSchool:          "kec-school-ovmgv",
	constants.AICSchool:          "aic-school-5qbbu",
	constants.NSGSchool:          "nsg-school-6osx0",
	constants.ManagaraBase:       "withus-managara-base-0wf23",
	constants.ManagaraHighSchool: "withus-managara-hs-t5fuk",
	constants.KECDemo:            "kec-demo-4ybj3",

	//JPREP doesn't has tenant on stag/uat/prod, this is for testing on local env
	constants.JPREPSchool: "jprep-eznr7",
}
