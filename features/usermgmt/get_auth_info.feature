@blocker
Feature: Get auth info

    Scenario: Get auth info
        When a user gets auth info by "available username" and "available domain name"
        Then user receives login email and tenant id successfully
        And receives "OK" status code


    Scenario Outline: Get auth info with invalid data "<username>" and "<org_domain_name>"
        When a user gets auth info by "<username>" and "<org_domain_name>"
        And receives "NotFound" status code

        Examples:
            | username                  | org_domain_name         |
            | unavailable username      | available domain name   |
            | available username        | unavailable domain name |
            | username from another org | available domain name   |
