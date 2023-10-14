Feature: Check Profile

    As a user in middle of registration process,
    I want to Check if anyone already using my email or phone number on Manabie platform

    # Invalid. Request must come from user in goranization
    # Scenario Outline: check with existed field
    #     Given an invalid authentication token
    #     When user check an "<field>" that "existed" in DB
    #     Then returns "OK" status code
    #     And that basic profile
    #     Examples:
    #         | field        |
    #         | email        |
    #         | phone_number |

    Scenario Outline: check with nonexisted field
        Given an invalid authentication token
        When user check an "<field>" that "nonexisted" in DB
        Then returns "NotFound" status code
        Examples:
            | field        |
            | email        |
            | phone_number |

    Scenario Outline: check with empty request
        Given an invalid authentication token
        When user check an "<field>" with empty value
        Then returns "InvalidArgument" status code
        Examples:
            | field        |
            | email        |
            | phone_number |
