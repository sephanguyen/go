@cms
@communication
@scheduled-notification

Feature: Cannot create scheduled notification on CMS

    Background:
        Given school admin logins CMS
        And school admin has created a student with grade, course and parent info
        And school admin is at "Notification" page on CMS

    Scenario Outline: Fail to save new scheduled notification when missing <requiredField> using "Save schedule" button
        Given school admin has opened compose new notification full-screen dialog
        When school admin selects "Schedule"
        And school admin leaves "<requiredField>" blank and click "Save schedule" button
        Then school admin sees scheduled notification is not created with "<numberOfError>" error validation
        Examples:
            | requiredField   | numberOfError |
            | Title           | 1             |
            | Content         | 1             |
            | Time            | 1             |
            | Recipient Group | 3             |

    Scenario Outline: Fail to save new scheduled notification when missing <requiredField> using "Close schedule" button
        Given school admin has opened compose new notification full-screen dialog
        When school admin selects "Schedule"
        And school admin leaves "<requiredField>" blank and click "Close schedule" button
        Then school admin sees new notification full-screen dialog still opened
        And school admin sees scheduled notification is not created with "<numberOfError>" error validation
        Examples:
            | requiredField   | numberOfError |
            | Time            | 1             |
            | Recipient Group | 3             |
