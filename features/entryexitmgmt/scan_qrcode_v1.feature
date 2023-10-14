@quarantined
Feature: Scan QR code

    Background:
        Given there is an existing student
        And student has "v1" qr version

    Scenario: Student scans v1 qrcode
        Given student has "no entry and exit" record
        And student has "Existing" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode with "Present" date in time zone "<time-zone>"
        Then scan returns "OK" status code
        And "TOUCH_ENTRY" touch type is recorded
        And parent receives notification status "Successfully"
        And name of the student is displayed on welcome screen

        Examples:
            | time-zone        | signed-in user |
            | Asia/Tokyo       | school admin   |
            | Asia/Ho_Chi_Minh | centre lead    |
            | Asia/Tokyo       | centre manager |
            | Asia/Ho_Chi_Minh | centre staff   |
            | Asia/Tokyo       | hq staff       |
            | Asia/Ho_Chi_Minh | teacher        |
