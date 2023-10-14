Feature: trigger migrate to exam lo submission and answer once submitted

    Scenario: trigger migrate to exam lo submission and answer once submitted
        Given user create a study plan of exam lo to database
        When user create a shuffle quiz set for exam lo with <numOfQuizzes> of quizzes
        And user submitted with <numOfAnswers> of answers
        Then database has a record in exam lo submission and <numOfQuizzes> records in exam lo submissions answer

        Examples:
            | numOfQuizzes | numOfAnswers |
            | 8            | 3            |
            | 0            | 0            |
            | 5            | 0            |