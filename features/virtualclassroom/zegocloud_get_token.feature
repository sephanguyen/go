Feature: Get Authentication Token for ZegoCloud

   Scenario: Teacher gets authentication token for zegocloud
    Given "teacher" signin system
    When user gets authentication token for zegocloud
    Then returns "OK" status code
    And user receives authentication token

    Given "teacher" signin system
    When user gets authentication token for zegocloud using v2
    Then returns "OK" status code
    And user receives authentication token from v2

   Scenario: Student gets authentication token for zegocloud
    Given "student" signin system
    When user gets authentication token for zegocloud
    Then returns "OK" status code
    And user receives authentication token

    Given "student" signin system
    When user gets authentication token for zegocloud using v2
    Then returns "OK" status code
    And user receives authentication token from v2