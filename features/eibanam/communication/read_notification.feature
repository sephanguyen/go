@cms @learner @parent
@communication

Feature: Read notification

    Background:
        Given "school admin" logins CMS
        And "school admin" has created 1 course
        And "school admin" has created a student with grade and parent info
        And "school admin" has added created course for student
        And "student" logins Learner app
        And "parent" logins Learner app
        And "school admin" is at "Notification" page on CMS

    Scenario Outline: Update read number in <section> list on CMS when <userAccount> read the notification
        Given "school admin" sends notification with required fields to student and parent
        # send notification for student and parent only
        When "<userAccount>" has read the notification
        Then school admin sees "<readNumber>" people display in "<section>" notification list on CMS
        Examples:
            | userAccount            | readNumber | section          |
            | 1 of [student, parent] | 1/2        | 1 of [All, Sent] |
            | student & parent       | 2/2        | 1 of [All, Sent] |

    @ignore
    Scenario Outline: Update read status in notification detail on CMS when <userAccount> read the notification
        Given "school admin" sends notification with required fields to student and parent
        # send notification for student and parent only
        When "<userAccount>" has read the notification
        Then school admin sees the status of "<userAccount>" is changed to "Read"
        Examples:
            | userAccount      |
            | student          |
            | parent           |
            | student & parent |

    Scenario: Student and parent open hyperlink in the content of notification
        Given school admin has created notification that content includes hyperlink
        And school admin has sent notification to student and parent
        When "student" interacts the hyperlink in the content on Learner App
        And "parent" interacts the hyperlink in the content on Learner App
        Then "student" redirects to web browser
        And "parent" redirects to web browser