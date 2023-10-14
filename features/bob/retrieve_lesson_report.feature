Feature: User Retrieve LessonReports
    Scenario Outline: user retrieve partner domain
        Given a signed in user "<role>" with school: <school signed>
        And user retrieve partner domain type "<domain type>"
        Then returns "OK" status code

        Examples:
            | role         | school signed | domain type |
            | school admin | 32            | Bo          |
            | school admin | 37            | Teacher     |
            | school admin | 34            | Learner     |
            | teacher      | 132           | Bo          |
            | teacher      | 134           | Teacher     |
            | teacher      | 136           | Learner     |
            | student      | 3             | Bo          |
            | student      | 3             | Teacher     |
            | student      | 3             | Learner     |
