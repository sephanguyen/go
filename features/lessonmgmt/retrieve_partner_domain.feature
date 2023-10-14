Feature: Retrieve Partner Domain
    Scenario Outline: user retrieve partner domain
        Given a signed in "<role>" with a school
        And user retrieve partner domain type "<domain type>"
        Then returns "OK" status code

        Examples:
            | role         | domain type |
            | school admin | Bo          |
            | school admin | Teacher     |
            | school admin | Learner     |
            | teacher      | Bo          |
            | teacher      | Teacher     |
            | teacher      | Learner     |
            | student      | Bo          |
            | student      | Teacher     |
            | student      | Learner     |
