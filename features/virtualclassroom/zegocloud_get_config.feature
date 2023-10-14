Feature: Get Chat Config for ZegoCloud

   Scenario: Teacher gets chat config for zegocloud
    Given "teacher" signin system
    When user gets chat config for zegocloud
    Then returns "OK" status code
    And user receives chat configurations

   Scenario: Student gets chat config for zegocloud
    Given "student" signin system
    When user gets chat config for zegocloud
    Then returns "OK" status code
    And user receives chat configurations