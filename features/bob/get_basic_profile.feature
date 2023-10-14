Feature: Retrieve basic profile Profile

    Scenario: user retrieves basic profile
        Given "<role>" signin system
        And a "valid" userId GetBasicProfileRequest
        When user retrieves basic profile
        Then returns "OK" status code
        And Bob must returns 1 basic profile

    Examples:
        | role         |
        | school admin |
        | student      |
        | parent       |
        | teacher      |

    Scenario: user retrieves basic profile with invalid userID
        Given "<role>" signin system
        And a "invalid" userId GetBasicProfileRequest
        When user retrieves basic profile
        And Bob must returns 0 basic profile

    Examples:
        | role         |
        | school admin |
        | student      |
        | parent       |
        | teacher      |

    Scenario: user retrieves basic profile with missing userID
        Given "<role>" signin system
        And a "missing id" userId GetBasicProfileRequest
        When user retrieves basic profile
        And user cannot retrieves basic profile when missing "userId"
    
    Examples:
        | role         |
        | school admin |
        | student      |
        | parent       |
        | teacher      |

    Scenario: user retrieves basic profile with missing metadata
        Given a signed in as "<role>"
        And a "valid" userId GetBasicProfileRequest
        When user retrieves basic profile with missing metadata
        And user cannot retrieves basic profile when missing "metadata"
    
    Examples:
        | role         |
        | school admin |
        | student      |
        | parent       |
        | teacher      |