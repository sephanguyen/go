@cms
@communication
@ignore

Feature: Discard draft and scheduled notification on CMS

    Background:
        Given "school admin" logins CMS
        And "school admin" has created a student with grade, course and parent info
        And "student" logins Learner App
        And "school admin" is at "Notification" page on CMS
        And "school admin" has created 1 "draft" notifications
        And "school admin" has created 1 "scheduled" notifications

    Scenario Outline: Discard a <type> notification successfully
        Given "school admin" has opened editor full-screen dialog of "<type>" notification
        When "school admin" clicks "Discard" button
        And "school admin" confirms to discard
        Then "school admin" sees "<type>" notification has been deleted on CMS
        Examples:
            | type      |
            | draft     |
            | scheduled |

    Scenario Outline: Cancel discard a <type> notification
        Given "school admin" has opened editor full-screen dialog of "<type>" notification
        When "school admin" clicks "Discard" button
        And "school admin" cancels to discard
        Then "school admin" still sees "<type>" notification on CMS
        Examples:
            | type      |
            | draft     |
            | scheduled |