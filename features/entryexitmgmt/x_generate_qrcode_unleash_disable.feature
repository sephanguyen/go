@quarantined
Feature: Generate Batch QR Codes With Unleash BACKEND_EntryExit_EntryExitManagement_SDK_Upload disabled

  Scenario Outline: Generate qrcodes for "<count>" student successfully with SDK upload feature flag disabled
    Given a qrcode request payload with "10" student ids
    And unleash feature flag is "disable" with feature name "BACKEND_EntryExit_EntryExitManagement_SDK_Upload"
    When entryexitmgmt generates qrcode for these student ids
    Then receives "OK" status code
    And response has no errors