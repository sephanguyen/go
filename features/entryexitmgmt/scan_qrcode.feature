@major
Feature: Scan QR code

    Background:
        Given there is an existing student
        And this student has "Existing" qr code record

    Scenario: Student scans qrcode for entry successfully
        Given student has "<entry-exit>" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode with "Present" date in time zone "<time-zone>"
        Then scan returns "OK" status code
        And "TOUCH_ENTRY" touch type is recorded
        And parent receives notification status "<notif-status>"
        And name of the student is displayed on welcome screen

        Examples:
            | entry-exit        | parent-info | time-zone        | notif-status   | signed-in user |
            | no entry and exit | Existing    | Asia/Tokyo       | Successfully   | school admin   |
            | exit              | No          | Asia/Tokyo       | Unsuccessfully | centre manager |
            | no entry and exit | No          | Asia/Ho_Chi_Minh | Unsuccessfully | centre staff   |
            | exit              | No          | Asia/Tokyo       | Unsuccessfully | hq staff       |
            | exit              | Existing    | Asia/Tokyo       | Successfully   | teacher        |

    Scenario: Student's first scan of the day is Entry
        Given student has "<entry-exit>" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode with "Present" date in time zone "<time-zone>"
        Then scan returns "OK" status code
        And "TOUCH_ENTRY" touch type is recorded
        And parent receives notification status "<notif-status>"
        And name of the student is displayed on welcome screen

        Examples:
            | entry-exit        | parent-info | time-zone        | notif-status   | signed-in user |
            | no entry and exit | Existing    | Asia/Tokyo       | Successfully   | school admin   |
            | past entry        | No          | Asia/Ho_Chi_Minh | Unsuccessfully | centre manager |
            | past completed    | No          | Asia/Tokyo       | Unsuccessfully | centre staff   |
            | past entry        | No          | Asia/Ho_Chi_Minh | Unsuccessfully | hq staff       |
            | past completed    | Existing    | Asia/Ho_Chi_Minh | Successfully   | teacher        |

    Scenario: Student scans qrcode for exit successfully
        Given student has "entry" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode with "Present" date in time zone "<time-zone>"
        Then scan returns "OK" status code
        And "TOUCH_EXIT" touch type is recorded
        And parent receives notification status "<notif-status>"
        And name of the student is displayed on welcome screen

        Examples:
            | parent-info | notif-status   | time-zone        | signed-in user |
            | Existing    | Successfully   | Asia/Tokyo       | school admin   |
            | No          | Unsuccessfully | Asia/Ho_Chi_Minh | centre lead    |
            | Existing    | Successfully   | Asia/Tokyo       | centre manager |
            | No          | Unsuccessfully | Asia/Ho_Chi_Minh | centre staff   |
            | Existing    | Successfully   | Asia/Tokyo       | hq staff       |
            | No          | Unsuccessfully | Asia/Tokyo       | teacher        |

    Scenario: Student scans qrcode for exit successfully when the equivalent UTC time of the entry date is previous date or yesterday
        Given student has "entry date equivalent to previous date in UTC" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode with "Fixed Time" date in time zone "<time-zone>"
        Then scan returns "OK" status code
        And "TOUCH_EXIT" touch type is recorded
        And parent receives notification status "<notif-status>"
        And name of the student is displayed on welcome screen

        Examples:
            | parent-info | time-zone        | notif-status   | signed-in user |
            | Existing    | Asia/Tokyo       | Successfully   | school admin   |
            | No          | Asia/Tokyo       | Unsuccessfully | centre manager |
            | No          | Asia/Ho_Chi_Minh | Unsuccessfully | centre staff   |
            | No          | Asia/Tokyo       | Unsuccessfully | hq staff       |
            | Existing    | Asia/Tokyo       | Successfully   | teacher        |

    Scenario: Student scans with past date
        Given student has "entry" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode with "Past" date in time zone "<time-zone>"
        Then scan returns "InvalidArgument" status code
        And parent receives notification status "Unsuccessfully"

        Examples:
            | parent-info | time-zone        | signed-in user |
            | Existing    | Asia/Tokyo       | school admin   |
            | No          | Asia/Ho_Chi_Minh | centre lead    |
            | Existing    | Asia/Tokyo       | centre manager |
            | No          | Asia/Ho_Chi_Minh | centre staff   |
            | Existing    | Asia/Tokyo       | hq staff       |
            | No          | Asia/Tokyo       | teacher        |

    Scenario: Student scans qrcode within 1 minute should failed
        Given student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        And student just scanned qrcode with "Present" date in time zone "<time-zone>"
        When student scans again
        Then scan returns "PermissionDenied" status code
        And parent receives notification status "Unsuccessfully"

        Examples:
            | parent-info | time-zone        | signed-in user |
            | Existing    | Asia/Tokyo       | school admin   |
            | No          | Asia/Ho_Chi_Minh | centre lead    |
            | Existing    | Asia/Tokyo       | centre manager |
            | No          | Asia/Ho_Chi_Minh | centre staff   |
            | Existing    | Asia/Tokyo       | hq staff       |
            | No          | Asia/Tokyo       | teacher        |

    Scenario: Student scans qrcode with invalid encryption key
        Given student has "entry" record
        And student has "<parent-info>" parent
        And student parent has existing device
        And "<signed-in user>" logins to backoffice app
        When student scans qrcode with "Present" date in time zone "<time-zone>" with invalid encryption
        Then scan returns "InvalidArgument" status code
        And parent receives notification status "Unsuccessfully"

        Examples:
            | parent-info | time-zone        | signed-in user |
            | Existing    | Asia/Tokyo       | school admin   |
            | No          | Asia/Ho_Chi_Minh | centre lead    |
            | Existing    | Asia/Tokyo       | centre manager |
            | No          | Asia/Ho_Chi_Minh | centre staff   |
            | Existing    | Asia/Tokyo       | hq staff       |
            | No          | Asia/Tokyo       | teacher        |
