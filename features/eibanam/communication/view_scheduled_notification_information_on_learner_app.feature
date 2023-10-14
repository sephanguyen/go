@cms @learner @parent
@communication
@scheduled-notification
@ignore

Feature: View scheduled notification information on Learner App

    Background:
        Given "school admin" logins CMS
        And "school admin" has created a student with grade, course and parent info
        And "student" logins Learner App
        And "parent" of "student" logins Learner App
        And "school admin" has created a scheduled notification
        And "school admin" is at "Notification" page on CMS

    Scenario Outline: Badge number displays <behaviour> if <userAccount> <action>
        Given scheduled notification has sent to "<userAccount>"
        When "<userAccount>" with "<action>" the scheduled notification
        Then "<userAccount>" receives notification with badge number of notification bell displays "<behaviour>" on Learner App
        Examples:
            | behaviour      | userAccount            | action |
            | 1              | 1 of [student, parent] | unread |
            | 0              | 1 of [student, parent] | read   |