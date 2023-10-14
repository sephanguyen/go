Feature: Create parents With Username
    As a school staff
    I'm able to create new parents with username

    Scenario: Create "1" new parents by staff granted "<role>"
        Given a signed in "staff granted role school admin"
        And new "2" parents data "available username"
        When "staff granted role school admin" create "2" new parents
        Then new parents were created successfully

    Scenario Outline: Cannot create if parent data to create has empty or invalid <field>
        Given a signed in "staff granted role school admin"
        And parent data has empty or invalid "<field>"
        When "staff granted role school admin" create new parents
        Then "staff granted role school admin" cannot create that account
        And receives "InvalidArgument" status code
        Examples:
            | field                            |
            | empty username                   |
            | username has spaces              |
            | username has special characters  |
            | existing username                |
            | existing username and upper case |
