Feature: User deletes lesson

  Background:
    Given a random number
    And some centers
    And some teacher accounts with school id
    And some student accounts with school id
    And a form's config for "individual lesson report" feature with school id
    And a form's config for "group lesson report" feature with school id
    And some courses with school id
    And some student subscriptions
    And some medias
    And a lesson
    And a lesson report

  Scenario: Admin can delete lesson
    Given "staff granted role school admin" signin system
    When user deletes a lesson
    Then returns "OK" status code
    And user no longer sees the lesson
    And user no longer sees any lesson report belong to the lesson

  Scenario: Teacher can delete lesson
    Given "staff granted role teacher" signin system
    When user deletes a lesson
    Then returns "OK" status code
    And user no longer sees the lesson
    And user no longer sees any lesson report belong to the lesson