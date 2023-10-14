Feature: trigger migrate to flash card submission once submitted

    Scenario: trigger migrate to flash card submission once submitted
        Given user create a study plan of flash card to database
        When user create a shuffle quiz set for flash card with <numOfQuizzes> of quizzes
        And user submitted with <numOfAnswers> of flash card answers
        Then database has <numOfSubmissions> of record in flash card submission and <numOfRecordAnswers> answers in flash card submission answer table

        Examples:
            | numOfQuizzes | numOfAnswers | numOfSubmissions | numOfRecordAnswers |
            | 8            | 3            | 1                | 3                  |
            | 5            | 0            | 0                | 0                  |
            | 5            | 5            | 1                | 5                  |