Feature: Import ClassDo Account

  Scenario Outline: Import ClassDo Account CSV
    Given user signed in as school admin
    When user imports ClassDo accounts with "<condition>" data
    Then returns "OK" status code
    And ClassDo accounts are "<existence>" in the database

    Examples:
      | condition | existence     |
      | valid     | existing      |
      | invalid   | not existing  |

  Scenario Outline: Import ClassDo Account CSV using delete action
    Given user signed in as school admin
    And user imports ClassDo accounts with "valid" data
    And returns "OK" status code
    When user imports ClassDo accounts with delete action
    Then returns "OK" status code
    And ClassDo accounts are "removed" in the database