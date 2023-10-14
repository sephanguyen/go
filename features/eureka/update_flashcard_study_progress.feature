Feature: update The Flashcard study progress
    In order to finish learning an objective or doing an exam
    As a student
    I need to take the quiz

    Background: given a quizet of an learning objective
        Given a quizset with "21" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

    Scenario Outline: a student update the flashcard study progress
        Given a signed in "student"
        And a study plan item id
        And user create flashcard study with valid request and limit "20" in the first time
        And retrieve last flashcard study progress with "without is_completed" arguments
        And returns "valid" study_set_id
        When update flashcard study progress with "<type>" arguments
        Then returns "<err>" status code

        Examples:
            | type                   | err  |
            | without study_set_id   | InvalidArgument |
            | without student_id     | InvalidArgument |
            | without studying_index | InvalidArgument |
            | empty                  | InvalidArgument |
            | valid                  | OK              |

    Scenario: a student update the flashcard study progress
        Given a signed in "student"
        And a study plan item id
        And user create flashcard study with valid request and limit "20" in the first time
        And retrieve last flashcard study progress with "without is_completed" arguments
        And returns "valid" study_set_id
        When update flashcard study progress with "valid" arguments
        Then returns "OK" status code
        And flashcard study progress must be updated
