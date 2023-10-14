Feature: Retrieve brightcove profile data

    Scenario Outline: Retrieve brightcove profile data
        Given a signed in "<signed as>"
        When api v2 get brightcove profile data
        Then returns "<msg>" status code

        Examples:
            | signed as       | msg             |
            | unauthenticated | Unauthenticated |
            | student         | OK              |
            | parent          | OK              |
