@major
Feature: Export bank
    As an HQ manager or admin
    I am able to export bank data

    Scenario Outline: Admin exports master bank data
        Given the organization "<organization>" has existing "<is-archived>" bank data
        And "<signed-in user>" logins to backoffice app
        When admin export the bank data
        Then receives "OK" status code
        And the bank CSV has a correct content

        Examples:
            | signed-in user | organization | is-archived  |
            | school admin   | -2147483630  | archived     |
            | hq staff       | -2147483630  | not archived |

    Scenario Outline: Admin exports master bank with no record
        Given the organization "<organization>" has no existing bank
        And "<signed-in user>" logins to backoffice app
        When admin export the bank data
        Then receives "OK" status code
        And the bank CSV only contains the header record

        Examples:
            | signed-in user | organization |
            | school admin   | -2147483629  |
            | hq staff       | -2147483629  |
