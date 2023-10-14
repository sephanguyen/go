@cms @learner
@communication
@scheduled-notification

Feature: Show error message after send scheduled notification

    Background:
        Given "school admin" logins CMS
        And school admin has created a student with grade, course and parent info
        And "student S1" logins Learner App
        And school admin has created a scheduled notification which will be sent 2 minute later
        And school admin is at "Notification" page on CMS

    Scenario Outline: Return message "The notification has been sent, you can no longer edit this notification"
        Given school admin has opened a scheduled notification dialog
        And school admin has edited scheduled notification
        When scheduled notification sent successfully on CMS
        And school admin clicks "<action>" in notification dialog
        Then school admin sees message "The notification has been sent, you can no longer edit this notification"
        Examples:
            | action         |
            | Save schedule  |
            | Close schedule |

    Scenario: Return message "The notification has been sent, you can no longer discard this notification"
        Given school admin has opened a scheduled notification dialog
        When scheduled notification sent successfully on CMS
        And school admin clicks "Discard and confirm" in notification dialog
        Then school admin sees message "The notification has been sent, you can no longer discard this notification"

    Scenario: Return message "The notification has been sent"
        Given school admin has opened a scheduled notification dialog
        When school admin selects "Now"
        And scheduled notification sent successfully on CMS
        And school admin clicks "Send" in notification dialog
        Then school admin sees message "The notification has been sent"
