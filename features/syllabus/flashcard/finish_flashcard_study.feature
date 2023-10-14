Feature: Finish flashcard study

  Background:a valid book content and flashcard study
    Given <flashcard>a signed in "school admin"
    And <flashcard>a valid book content
    And a valid flashcard with quizzes
    And <flashcard>a signed in "student"
    And <flashcard>a course and study plan with "current" student
    And user create flashcard study

  Scenario: Finish flashcard study with role
    Given <flashcard>an exists "<role>" signed in
    When user finish flashcard study with "no restart"
    Then <flashcard>returns "<msg>" status code
    Examples:
      | role         | msg              |
      | parent       | PermissionDenied |
      | student      | OK               |
      | school admin | PermissionDenied |
      | hq staff     | PermissionDenied |
      | teacher      | PermissionDenied |

  Scenario: Student finish the flashcard study
    When user finish flashcard study with "<option>"
    Then our system updates flashcard study correctly with "<option>"
    Examples:
      | option     |
      | restart    |
      | no restart |