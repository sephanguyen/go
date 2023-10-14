@cms @learner @learner2
@communication
@scheduled-notification
@staging

Feature: Edit scheduled notification on CMS

    Background:
        Given "school admin" logins CMS
        And school admin has created a student with grade, course and parent info
        And "student" logins Learner App
        And school admin has created a scheduled notification

    Scenario Outline: Update <field> of scheduled notification successfully with <button> button
        Given school admin has opened a scheduled notification dialog
        When school admin edits "<field>" of scheduled notification
        And school admin clicks "<button>" button
        Then school admin sees updated scheduled notification on CMS
        Examples:
            | field                | button                               |
            | Title                | 1 of [Save schedule, Close schedule] |
            | Content              | 1 of [Save schedule, Close schedule] |
            | Date                 | 1 of [Save schedule, Close schedule] |
            | Time                 | 1 of [Save schedule, Close schedule] |
            | Grade                | 1 of [Save schedule, Close schedule] |
            | Course               | 1 of [Save schedule, Close schedule] |
            | Individual Recipient | 1 of [Save schedule, Close schedule] |
            | All fields           | 1 of [Save schedule, Close schedule] |