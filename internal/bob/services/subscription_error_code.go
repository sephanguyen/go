package services

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

const (
	packageNotFound         = "packageNotFound"
	packageEnded            = "packageEnded"
	missingUpgradeCondition = "missingUpgradeCondition"

	// promotion error
	incorrectedPromoCode           = "incorrectedPromoCode"
	expiredPromoCode               = "expiredPromoCode"
	exceedRedemptionLimitPromoCode = "exceedRedemptionLimitPromoCode"
	invalidPromoCode               = "invalidPromoCode"
	alreadyAppliedPromoCode        = "alreadyAppliedPromoCode"
	exceedPerUserLimitPromoCode    = "exceedPerUserLimitPromoCode"
)

var (
	packageNotFoundMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        packageNotFound,
				Subject:     "payment",
				Description: "requested package not found",
			},
		},
	}

	packageEndedMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        packageEnded,
				Subject:     "payment",
				Description: "requested package ended",
			},
		},
	}

	missingUpgradeConditionMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        missingUpgradeCondition,
				Subject:     "payment",
				Description: "You are not eligible to purchase this package",
			},
		},
	}

	// promotion error details
	incorrectedPromoCodeErrDetails = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        incorrectedPromoCode,
				Subject:     "promotion",
				Description: "This code has been entered incorrectly.",
			},
		},
	}

	expiredPromoCodeErrDetails = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        expiredPromoCode,
				Subject:     "promotion",
				Description: "Sorry, this code has been expired.",
			},
		},
	}

	exceedRedemptionLimitPromoCodeErrDetails = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        exceedRedemptionLimitPromoCode,
				Subject:     "promotion",
				Description: "Sorry, this code has reached the redemption limit.",
			},
		},
	}

	invalidPromoCodeErrDetails = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        invalidPromoCode,
				Subject:     "promotion",
				Description: "This code is invalid.",
			},
		},
	}

	alreadyAppliedPromoCodeErrDetails = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        alreadyAppliedPromoCode,
				Subject:     "promotion",
				Description: "This code is already applied for activation.",
			},
		},
	}

	exceedPerUserLimitPromoCodeErrDetails = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        exceedPerUserLimitPromoCode,
				Subject:     "promotion",
				Description: "Sorry, you can not use this code anymore.",
			},
		},
	}
)
