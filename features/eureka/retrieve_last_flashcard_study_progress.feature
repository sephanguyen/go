Feature: Retrieve last flashcard study progress

    Background: given a quizet of an learning objective
        Given a quizset with "21" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic
    Scenario Outline: retrieve last flashcard study progress when no records
        Given a signed in "student"
        And a study plan item id
        When retrieve last flashcard study progress with "<type>" arguments
        Then returns "<err>" status code

        Examples:
            | type                          | err               |
            | empty                         | InvalidArgument   |
            | valid                         | OK                |
            | without student_id            | InvalidArgument   |
            | without lo_id                 | InvalidArgument   |
            | without study_plan_item_id    | InvalidArgument   |
            | without is_completed          | OK                |

    Scenario: retrieve last flashcard study progress
        Given a signed in "student"
        And a study plan item id
        When user create flashcard study with "keep order" and "20" flashcard study quizzes per page
        And retrieve last flashcard study progress with "without is_completed" arguments
        Then returns "valid" study_set_id
        And last flashcard study progress response match with the response from bob service

    Scenario: retrieve last flashcard study progress with is_completed
        Given a signed in "student"
        And a study plan item id
        When user create flashcard study with "keep order" and "20" flashcard study quizzes per page
        And retrieve last flashcard study progress with "valid" arguments
        Then returns "empty" study_set_id