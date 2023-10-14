Feature: Sync academic year

	Scenario: Jprep sync academicYear to our system
		Given some academic year message
		When jprep sync academic year to our system
		Then these academic years must be store in our system
