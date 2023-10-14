@cms @learner
@communication
@scheduled-notification

Feature: Cannot edit scheduled notification on CMS

    Background:
        Given "school admin" logins CMS
        And school admin has created a student with grade, course and parent info
        And "student S1" logins Learner App
        And school admin has created a scheduled notification
        And school admin is at "Notification" page on CMS

    Scenario Outline: Fail to update new scheduled notification when missing <requiredField> using "Save scheduled" button
        Given school admin has opened a scheduled notification dialog
        When school admin clear value of "<requiredField>" field and "Save schedule" notification
        Then school admin sees "<numberOfError>" of required errors validation message in form of "<requiredField>"
        Examples:
            | requiredField   | numberOfError |
            | Title           | 1             |
            | Content         | 1             |
            | Time            | 1             |
            | Recipient Group | 3             |

    Scenario Outline: Fail to update new scheduled notification when missing <requiredField> using "Close" button
        Given school admin has opened a scheduled notification dialog
        When school admin clear value of "<requiredField>" field and "Close schedule" notification
        Then school admin sees "<numberOfError>" of required errors validation message in form of "<requiredField>"
        Examples:
            | requiredField   | numberOfError |
            | Time            | 1             |
            | Recipient Group | 3             |
