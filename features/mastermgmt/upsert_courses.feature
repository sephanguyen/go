Feature: Upsert courses
    Background:
        Given a random number
        And some centers
        And some course types

    Scenario Outline: user update/insert courses with 2 locations
        Given "<signed-in user>" signin system
        When user upsert courses "<course data a>" data with 2 locations and teaching method "<teaching_method>"
        Then returns "<msg>" status code
        And course access paths already exist in DB with 2 locations
        And course saved DB correct with teaching method "<teaching_method>"
        Examples:
            | signed-in user | course data a | msg | teaching_method |
            | school admin   | valid         | OK  | individual      |

