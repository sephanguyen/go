Feature: get courses by ids

    Scenario: get courses by ids
        Given "school admin" signin system
        When user get courses by "<id-type>" ids
        Then must return a correct list of "<course-type>" courses
        Examples:
            | id-type      | course-type |
            | existing     | valid       |
            | non-existing | empty       |
            | mixed        | mixed       |
