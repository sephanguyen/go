@quarantined
Feature: Student retrieve grades

  Scenario: Student retrieve their grades
    Given some student has submission with status "SUBMISSION_STATUS_RETURNED"
    When teacher retrieve student grade base on submission grade id
    Then returns "OK" status code
    And eureka must return correct grades for each submission

  Scenario: Student retrieve their grades
    Given some student has submission with status "SUBMISSION_STATUS_MARKED"
    And "student" retrieve their own submissions
    When student retrieve their grade
    Then returns "OK" status code