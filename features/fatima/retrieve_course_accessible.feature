@quarantined
Feature: Retrieve course accessible
  Background:
    Given some package data in db

  Scenario: Unauthenticated user retrieves course accessible
    Given an invalid authentication token
    When user retrieve accessible course
    Then returns "Unauthenticated" status code

  Scenario Outline: Authenticate user retrieves course accessible
    Given a signed in "<signed-in user>"
    And this user has package "<package_name>" is "<status>"
    When user retrieve accessible course
    Then returns "OK" status code
    And returns all CourseAccessibleResponse of this user
    Examples:
      | signed-in user | package_name                     | status          |
      | school admin   | free_package                     | valid           |
      | school admin   | free_package                     | expired         |
      | school admin   | free_package,basic_trial_package | valid,expired   |
      | school admin   | free_package,basic_trial_package | expired,valid   |
      | school admin   | free_package,basic_trial_package | expired,expired |
