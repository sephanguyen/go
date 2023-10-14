@quarantined
Feature: User deletes lesson recurring

  Background:
    And some centers
    And some teacher accounts with school id
    And some student accounts with school id
    And a form's config for "individual lesson report" feature with school id
    And a form's config for "group lesson report" feature with school id
    And some courses with school id
    And some student subscriptions

  Scenario: User delete from base lesson or following lessons
    Given user signed in as school admin
    When user creates recurring lesson
    And submit many lesson reports from many lessons recurring "<existing_lesson>"
    And the lesson "<locked_lessons>" will locked
    When user deletes recurring lesson from "<delete_from>" with "<method>"
    Then returns "OK" status code
    And user no longer sees the lessons from many lessons recurring "<deleted_lesson>"
    And user no longer sees any lessons report belong to the lessons from many lessons recurring "<deleted_lesson>"
    And user still sees the lesson from many lessons recurring "<remaining_lesson>"
    And user still sees lesson report belong to the lesson from many lessons recurring "<remaining_lesson>"
    And Lessonmgmt must push msg "DeleteLesson" subject "Lesson.Deleted" to nats
    
    Examples:
      | method    | existing_lesson | delete_from | locked_lessons | deleted_lesson | remaining_lesson |
      | one_time  | 0,1,2,3         | 2           |                | 2              | 0,1,3            |
      | recurring | 0,1,2,3         | 0           |                | 0,1,2,3        |                  |
      | recurring | 0,1,2,3         | 2           |                | 2,3            | 0,1              |
      | recurring | 0,1,2,3         | 3           |                | 3              | 0,1,2            |
      |           | 0,1,2,3         | 2           |                | 2              | 0,1,3            |
      | recurring | 0,1,2,3         | 0           | 2,3            | 0,1            | 2,3              |