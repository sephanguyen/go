Feature: trigger migrate to lo submission from shuffled quiz set once student submitted

    Scenario: trigger migrate to lo submission from shuffled quiz set once student submitted
        Given a valid learning objective in database
        And a study plan of lo in database
        When student create a valid shuffle quiz set for lo
        And student submit with <numOfAnswers> of answers
        And student submit with <numOfRetryAnswers> of answers
        Then database must have <numOfSubmission> records lo submission and <numOfExpectedAnswers> records lo submission answer correctly

        Examples:
            | numOfAnswers | numOfRetryAnswers | numOfSubmission | numOfExpectedAnswers |
            | 2            | 0                 | 1               | 2                    |
            | 2            | 1                 | 1               | 2                    |
            | 2            | 3                 | 1               | 3                    |
            | 0            | 0                 | 0               | 0                    |
            | 0            | 2                 | 1               | 2                    |

