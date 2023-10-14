@quarantined
Feature: The Admin/schoolAdmin retrieve student accessible
Background:
    Given some package data in db
Scenario Outline: The user of invalid roles try to retrieve user accessible
    Given a signed as "<signed as>"
    When the user retrieve student accessible course
    Then returns "PermissionDenied" status code
    Examples:
      | signed as       |
      | student         |
      | teacher         |

Scenario Outline: Authenticate user retrieves student course accessible
    Given a signed as "admin"
        And a student has package "<package_name>" is "<status>"
    When the user retrieve student accessible course
    Then returns "OK" status code
    And returns all CourseAccessibleResponse of this student
    Examples:
      | package_name                     | status          |
      | free_package                     | valid           |
      | free_package                     | expired         |
      | free_package,basic_trial_package | valid,expired   |
      | free_package,basic_trial_package | expired,valid   |
      | free_package,basic_trial_package | expired,expired |
