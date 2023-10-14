Feature: Multiple Student interact with school qrcode scanner

    Scenario: Student scans qr code on scanner
        Given scanner is setup on "organization -2147483648"
        And there is an existing student with qr code from "<organization>"
        When this student scans qr code
        Then scanner should return "<result>"

        Examples:
            | result         | organization              |
            | successfully   | organization -2147483648  |
            | unsuccessfully | organization -2147483646  |