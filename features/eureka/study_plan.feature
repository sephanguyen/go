Feature: Create study plan

    Scenario: Create new study plan
        Given a valid "teacher" token
        When user create study plan
        Then returns "OK" status code
        Then eureka must store correct study plan

    Scenario: List student study plans
        Given some students are assigned some valid study plans
        When teacher list study plans for each student
        Then returns a list of assigned study plans of each student
