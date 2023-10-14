Feature: User deletes lesson

  Background:
    Given user signed in as school admin
    And some centers
    And some teacher accounts with school id
    And some student accounts with school id
    And a form's config for "individual lesson report" feature with school id
    And a form's config for "group lesson report" feature with school id
    And some courses with school id
    And some student subscriptions

  Scenario: Admin can delete lesson
    Given user signed in as school admin
    And an existing lesson in lessonmgmt
    And a lesson report
    When user deletes a lesson
    Then returns "OK" status code
    And user no longer sees the lesson
    And user no longer sees any lesson report belong to the lesson
    And Lessonmgmt must push msg "DeleteLesson" subject "Lesson.Deleted" to nats

  Scenario: Teacher can delete lesson
    Given a signed in teacher
    And an existing lesson in lessonmgmt
    And a lesson report
    When user deletes a lesson
    Then returns "OK" status code
    And user no longer sees the lesson
    And user no longer sees any lesson report belong to the lesson
    And Lessonmgmt must push msg "DeleteLesson" subject "Lesson.Deleted" to nats
    
