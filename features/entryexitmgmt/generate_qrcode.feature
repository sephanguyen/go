@major
Feature: Generate Batch QR Codes

  Scenario Outline: Generate qrcodes for "<count>" student successfully
    Given a qrcode request payload with "<count>" student ids
    And unleash feature flag is "enable" with feature name "BACKEND_EntryExit_EntryExitManagement_SDK_Upload"
    When "<signed-in user>" generates qrcode for these student ids
    Then receives "OK" status code
    And response has no errors

    Examples:
      | count | signed-in user |
      | 5     | school admin   |
      | 5     | hq staff       |
      | 5     | centre lead    |
      | 5     | centre manager |
      | 5     | centre staff   |

  Scenario: Generate new version of qrcode when student has v1 qrcode version
    Given a qrcode request payload with "5" student ids
    And student has "v1" qr version
    When "<signed-in user>" generates qrcode for these student ids
    Then receives "OK" status code
    And response has no errors
    And student should have updated qrcode version

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |
      | centre lead    |
      | centre manager |
      | centre staff   |

  Scenario: Generate qrcodes with Invalid payload
    Given a qrcode request payload with "0" student ids
    When "<signed-in user>" generates qrcode for these student ids
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |
      | centre lead    |
      | centre manager |
      | centre staff   |
