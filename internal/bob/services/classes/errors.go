package classes

import "google.golang.org/genproto/googleapis/rpc/errdetails"

const (
	errSubjectClass = "class"
)

var (
	errWrongClassCodeMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        "wrong-class-code",
				Subject:     errSubjectClass,
				Description: "Wrong class code",
			},
		},
	}

	errClassDoesNotBelongToYourSchoolMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        "class-does-not-belong-to-your-school",
				Subject:     errSubjectClass,
				Description: "Class doest not belong to you school",
			},
		},
	}

	errClassCodeExpired = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        "class-code-expired",
				Subject:     errSubjectClass,
				Description: "Class code expired",
			},
		},
	}
)
