Feature: Retrieve flashcard study progress

    Background: given a quizet of an learning objective
        Given a quizset with "21" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

    Scenario Outline: retrieve flashcard study progress
        Given a valid student account
        And user create flashcard study with "keep order" and "21" flashcard study quizzes per page
        When retrieve flashcard study progress with "<type>" arguments
        Then returns "<err>" status code
        And flashcard study progress response match with the response from bob service

        Examples:
            | type                          | err               |
            | empty                         | InvalidArgument   |
            | valid                         | OK                |
            | without student_id            | InvalidArgument   |
            | without study_set_id          | InvalidArgument   |
            | without paging                | InvalidArgument   |

  