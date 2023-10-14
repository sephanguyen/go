@quarantined
Feature: Upsert Preset Study Plans
    Update preset study plans

    Scenario: admin upsert preset study plan
        Given a signed in admin
        And 4 valid preset study plan
        When user upsert preset study plan
        Then returns a list of stored study plan
        And Bob must store all preset study plan

    Scenario: admin upsert preset study plan without id
        Given a signed in admin
        And 4 valid preset study plan without Id
        When user upsert preset study plan
        Then returns a list of stored study plan
        And Bob must store all preset study plan

    Scenario: student upsert preset study plan
        Given a signed in student
        And 4 valid preset study plan
        When user upsert preset study plan
        Then returns "PermissionDenied" status code
