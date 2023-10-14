Feature: Retrieve list of notifications
    @major
    Scenario: Retrieve list of important notifications
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student logins to Learner App
        And school admin sends "<number notifications>" of notifications with "<number important notifications>" of important notifications to a student
        When student retrieves list of notifications with important only filter is "<is important only filter>"
        Then returns correct list of notifications with counting is "<return number notifications>"
        Examples:
            | number notifications | number important notifications | is important only filter | return number notifications |
            | 5                    | 1                              | false                    | 5                           |
            | 15                   | 7                              | true                     | 7                           |
