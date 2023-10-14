@quarantined
Feature: Update student name

  Background:
    Given have some centers
    And have some teacher accounts
    And have some courses

Scenario Outline: Update a student account successfully with first name last name and phonetic name
    Given "school admin" signin system
    And user creates a new lesson with all required fields in lessonmgmt
    And returns "OK" status code
    And only student info with first name last name and phonetic name
    And create new student account
    And returns "OK" status code
    And assign student to a student subscription
    And assign student to a lesson
    And student account data to update with first name lastname and phonetic name
    When update student account
    Then returns "OK" status code
    And student name is updated correctly
