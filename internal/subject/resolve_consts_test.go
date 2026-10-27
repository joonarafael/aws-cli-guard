package subject_test

const (
	awsBin             = "aws"
	profileOpt         = "--profile"
	profileNameDefault = "default"
	profileNameProd    = "prod"
	reasonAmbient      = "ambient-credentials"
	regionOpt          = "--region"
	outputOpt          = "--output"
	outputJSON         = "json"

	reasonProfileFlagMissingValue  = "profile-flag-missing-value"
	reasonGlobalOptionMissingValue = "global-option-missing-value"
	reasonAmbiguousGlobalOption    = "ambiguous-global-option"
	reasonGlobalOptionUnexpected   = "global-option-unexpected-value"
	reasonGlobalOptionInvalidValue = "global-option-invalid-value"

	fixtureAccessKey = "AKIAFIXTUREACCESSKEY123456789"
	fixtureSecretKey = "wJalrFIXTURESECRETNOTREALKEY12"
	fixtureSession   = "FIXTURESESSIONTOKENVALUE"
)
