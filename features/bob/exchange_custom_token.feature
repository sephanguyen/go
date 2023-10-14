Feature: exchange custom token

    @quarantined
    Scenario: exchange custom token with valid authentication token
        Given an other student profile in DB
            And a valid authentication token with ID already exist in DB
        When a client exchange custom token
        Then our system need to returns a valid custom token
    @blocker 
    Scenario: exchange custom token with valid parent authentication token
        Given a new parent profile in DB
        When a client exchange custom token
        Then our system need to returns a valid custom token

    @quarantined
    Scenario: exchange tenant custom token with valid authentication token
        Given an other student profile in DB
            And an identity platform account with existed account in db
            And a valid authentication token with tenant
        When a client exchange custom token
        Then our system need to returns a valid custom token

    @quarantined
    Scenario: exchange tenant custom token with valid parent authentication token
        Given a new parent profile in DB
            And an identity platform account with existed account in db
            And a valid authentication token with tenant
        When a client exchange custom token
        Then our system need to returns a valid custom token

    @quarantined
    Scenario: exchange tenant custom token with invalid authentication token
        Given an other student profile in DB
        When a client exchange custom token
        Then our system need to returns error