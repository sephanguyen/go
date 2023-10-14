Feature: User updates scheduling status

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario Outline: School admin scheduling status
    Given user signed in as school admin
    And an existing lesson
    When user updates scheduling status in the lesson is "<value>"
    Then returns "OK" status code
    And the lesson scheduling status was updated
    And Lessonmgmt must push msg "UpdateLesson" subject "Lesson.Updated" to nats

    Examples:
      | value                              |
      | LESSON_SCHEDULING_STATUS_CANCELED  |
      | LESSON_SCHEDULING_STATUS_COMPLETED |

   Scenario Outline: School admin cannot update lesson status when lesson is locked
    Given user signed in as school admin
    And an existing lesson
    And the lesson is locked "true"
    When user updates scheduling status in the lesson is "<value>"
    Then returns "Internal" status code
  
    Examples:
      | value                              |
      | LESSON_SCHEDULING_STATUS_CANCELED  |
      | LESSON_SCHEDULING_STATUS_PUBLISHED |

   Scenario Outline: School admin can update lesson status
    Given user signed in as school admin
    And user have created recurring lesson
    When user change status to "<status>" by saving "<saving_type>"
    Then returns "OK" status code
    And the lesson scheduling status was updated
    Examples:
      | status            | saving_type         |
      | draft             | only this           | 
      | draft             | this and following  |