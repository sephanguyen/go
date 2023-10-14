Feature: Export course access paths

    Export course access paths
    Background:
        Given some centers
        And some course types
        And seeded 20 courses
        And have some course access paths

    Scenario: Export course access paths
        Given "school admin" signin system
        When user export course access paths
        Then returns course access paths in csv with Ok status code
