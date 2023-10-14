Feature: Parent leaves conversation

    Background:
        Given "school admin" logins CMS
        And "school admin" has created "student S1" with "parent P1, parent P2" info
        And "teacher" has joined "parent P1" chat group
        And "parent P1, parent P2" logins on Learner App

    @wip
    Scenario Outline: Parent leaves parent chat group
        Given "parent P1" has sent "<messageType>" message to parent chat group
        When school admin removes the relationship of "parent P1" and "student S1"
        Then "parent P1" sees parent chat group is removed on Learner App
        And teacher sees "<messageType>" message of "parent P1" with name and avatar
        And "parent P2" sees "<messageType>" message of "parent P1" with name and avatar
        Examples:
            | messageType             |
            | 1 of [text, image, pdf] |

    @wip
    Scenario Outline: Parent leaves and rejoins parent chat group
        Given school admin removes the relationship of "parent P1" and "student S1"
        And school admin add the relationship of "parent P1" and "student S1"
        Given "parent P1" has sent "<messageType>" message to parent chat group
        Then teacher sees "<messageType>" message of "parent P1" with name and avatar
        And "parent P2" sees "<messageType>" message of "parent P1" with name and avatar
        Examples:
            | messageType             |
            | 1 of [text, image, pdf] |