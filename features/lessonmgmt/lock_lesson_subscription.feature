@quarantined
Feature: Lock lesson subscription
    Scenario: Lock lesson subscription
        Given "staff granted role school admin" signin system
        And "disable" Unleash feature with feature name "Lesson_Student_UseUserBasicInfoTable"
        And an existing "SUBMITTED" timesheet for current staff
        And timesheet has lesson records with "CANCELLED-CANCELLED-COMPLETED"
        When current staff approves this timesheet
        Then returns "OK" status code
        And timesheet status changed to approve "successfully"
        And update lock lesson successfully