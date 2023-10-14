@blocker
Feature: Get auth info

    Scenario Outline: Get auth info
        When a user reset password by "available username" and "available domain name" in "<language>"
        Then receives "OK" status code
        And user received reset password email in "<language>"

        Examples:
            | language |
            | japanese |
            | english  |
            | empty    |


    Scenario Outline: Get auth info with invalid data "<username>" and "<org_domain_name>"
        When a user reset password by "<username>" and "<org_domain_name>" in "<language>"
        Then receives "<status>" status code

        Examples:
            | username                  | org_domain_name         | language | status   |
            | unavailable username      | available domain name   | japanese | NotFound |
            | available username        | unavailable domain name | japanese | NotFound |
            | username from another org | available domain name   | japanese | NotFound |
