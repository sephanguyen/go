Feature: Retrieve learning objectives

    Background:a valid book content
        Given "school admin" logins "CMS"
        And a valid book content
        And user create learning objectives

    Scenario Outline: authenticate retrieve learning objectives
        Given "<user>" logins "CMS"
        When retrieve learning objectives with "TopicIds"
        Then returns "<status code>" status code

        Examples:
            | user           | status code |
            | admin          | OK          |
            | school admin   | OK          |
            | hq staff       | OK          |
            | teacher        | OK          |
            | student        | OK          |
            | parent         | OK          |
            | center lead    | OK          |
            | center manager | OK          |
            | center staff   | OK          |

    Scenario Outline:
        Given some lo completenesses existed in db
        When retrieve learning objectives with "<params>"
        Then our system must return learning objectives correctly
        Examples:
            | params               |
            | TopicIds             |
            | LoIds                |
            | WithCompleteness     |
            | WithAchievementCrown |