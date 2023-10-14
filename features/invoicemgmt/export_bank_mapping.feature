@major
Feature: Export bank mapping
    As an HQ manager or admin
    I am able to export bank mapping data

    Scenario Outline: Admin exports master bank mapping data
        Given the organization "<organization>" has existing bank mappings
        And "<signed-in user>" logins to backoffice app
        When admin export the bank mapping data
        Then receives "OK" status code
        And the bank mapping CSV has a correct content

        Examples:
            | signed-in user | organization |
            | school admin   | -2147483630  |
            | hq staff       | -2147483630  |

    Scenario Outline: Admin exports master bank mapping with no record
        Given the organization "<organization>" has no existing bank mapping
        And "<signed-in user>" logins to backoffice app
        When admin export the bank mapping data
        Then receives "OK" status code
        And the bank mapping CSV only contains the header record

        Examples:
            | signed-in user | organization |
            | school admin   | -2147483629  |
            | hq staff       | -2147483629  |
