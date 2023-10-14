Feature: Upsert study plan
    Background:
        Given a valid "school admin" token
        And a valid course student background
        And add a valid book with some learning objectives to course

    Scenario: update invalid study plan
        When user update a study plan with invalid study_plan_id
        Then returns "NotFound" status code

    Scenario Outline: update invalid study plan
        Given user create a valid study plan with "<study_plan_type>"
        When user update study plan
        Then study plans have been updated
        Examples:
            | study_plan_type |
            | course          |
            | individual      |

    Scenario: upsert valid study plan
        When user create a valid study plan
        Then study plans and related items have been stored

    Scenario: upsert valid study plan when course does not have any students
        Given user add a book to course does not have any students
        When user create a valid study plan
        Then study plans and related items have been stored

    Scenario Outline: upsert valid study plan when book does not have los or assingments
        Given user add a valid book does not have any "<type>" to course
        When user create a valid study plan
        Then study plans and related items have been stored
        Examples:
            | type                              |
            | learning_objective                |
            | assignment                        |
            | learning_objective and assignment |