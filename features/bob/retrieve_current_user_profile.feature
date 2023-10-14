Feature: Retrieve current user Profile

    Scenario: admin user retrieves user profile
        Given "staff granted role school admin" signin system
        When user retrieves his own profile
        Then Bob must returns user own profile

    Scenario: student user retrieves user profile
        Given "student" signin system
        When user retrieves his own profile
        Then Bob must returns user own profile
