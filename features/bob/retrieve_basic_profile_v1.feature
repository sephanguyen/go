Feature: Retrieve Basic Profile

    Scenario Outline: a valid user retrieves basic profile
        Given "<role>" signin system
        And a "valid" userId RetrieveBasicProfileRequest
        When a user retrieves basic profile
        Then returns "OK" status code
        And a Bob must returns 1 basic profile

        Examples:
            | role         |
            | school admin |
            | student      |
            | parent       |
            | teacher      |
    
    Scenario: user retrieves basic profile with invalid userId
        Given "<role>" signin system
        And a "invalid" userId RetrieveBasicProfileRequest
        When a user retrieves basic profile
        Then a Bob must returns 0 basic profile

        Examples:
            | role         |
            | school admin |
            | student      |
            | parent       |
            | teacher      |

    Scenario: user retrieves basic profile without userId
        Given "<role>" signin system
        And a "missing" userId RetrieveBasicProfileRequest
        When a user retrieves basic profile
        Then returns "InvalidArgument" status code

        Examples:
            | role         |
            | school admin |
            | student      |
            | parent       |
            | teacher      |

    Scenario: student retrieves others role basic profile
        Given a signed in student
        And a user retrieves "<user role>" basic profile request
        When a user retrieves basic profile
        Then returns "<status code>" status code
        
        Examples:
            | user role    | status code     |
            | student      | OK              |
            | school admin | InvalidArgument |
            | parent       | InvalidArgument |
            | teacher      | InvalidArgument |
    
    Scenario: a valid user retrieves basic profile without metadata
        Given "<role>" signin system
        And a "valid" userId RetrieveBasicProfileRequest
        When a user retrieves basic profile without metadata
        Then returns "InvalidArgument" status code

        Examples:
            | role         |
            | school admin |
            | student      |
            | parent       |
            | teacher      | 