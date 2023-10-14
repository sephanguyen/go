Feature: check tag name exist
    Scenario: user checks if tag name exist
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And a tag "<is_exist>" in database
        When admin send check request
        Then check response is "<is_exist>"
        Examples:
            | is_exist |
            | true     |
            | false    |