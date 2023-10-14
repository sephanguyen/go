@blocker
Feature: Scan QR code Multiple

    Background:
        Given there is an existing student
        And this student has "Existing" qr code record

    Scenario Outline: Scenario Outline name: Student scans qrcode for entry successfully with multiple request
        Given student has "<entry-exit>" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode "<request-count>" with "Present" date in time zone "<time-zone>"
        Then student has no multiple record
        And "TOUCH_ENTRY" touch type is recorded
        And parent receives notification status "<notif-status>"
        And name of the student is displayed on welcome screen

        Examples:
            | entry-exit        | parent-info | time-zone        | notif-status   | signed-in user | request-count |
            | no entry and exit | Existing    | Asia/Tokyo       | Successfully   | school admin   | 2             |

    Scenario Outline: Student scans qrcode for exit successfully with multiple request
        Given student has "entry" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode "<request-count>" with "Present" date in time zone "<time-zone>"
        Then student has no multiple record
        And "TOUCH_EXIT" touch type is recorded
        And parent receives notification status "<notif-status>"
        And name of the student is displayed on welcome screen

        Examples:
            | parent-info | notif-status   | time-zone        | signed-in user | request-count |
            | Existing    | Successfully   | Asia/Tokyo       | school admin   | 2             |