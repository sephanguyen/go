@cms @learner
@communication
@scheduled-notification

Feature: Create scheduled notification on CMS

    Background:
        Given "school admin" logins CMS
        And "school admin" has created a student with grade, course and parent info
        And "student" logins Learner App
        And "school admin" is at "Notification" page on CMS

    Scenario Outline: Save scheduled notification successfully with required field using <button> button
        Given "school admin" has opened compose new notification full-screen dialog
        When "school admin" fills scheduled notification information
        And "school admin" clicks "<button>" button
        Then "school admin" sees new scheduled notification on CMS
        Examples:
            | button         |
            | Save schedule  |
            | Close schedule |

    Scenario Outline: Save scheduled notification successfully with all fields using <button> button
        Given "school admin" has opened compose new notification full-screen dialog
        When "school admin" fills scheduled notification information
        And "school admin" has attached PDF file
        And "school admin" clicks "<button>" button
        Then "school admin" sees new scheduled notification on CMS
        Examples:
            | button         |
            | Save schedule  |
            | Close schedule |

    Scenario: Update a draft notification to scheduled notification
        Given "school admin" has created a draft notification
        When "school admin" opens editor full-screen dialog of draft notification
        And "school admin" selects notification status "Schedule"
        And "school admin" selects date, time of schedule notification
        And "school admin" clicks "Save schedule" button
        Then "school admin" sees draft notification has been saved to scheduled notification


