Feature: List tag by student in course
  As teacher group
  I need to be able to get a list of tags by student in course

  Background:
    Given "1" student login
    And <course_statistical>a school admin login
    And <course_statistical>a teacher login
    And "school admin" has created a book with each "1" los, "1" assignments, "1" topics, "1" chapters, "2" quizzes
    And "school admin" has created a course with a book
    And "school admin" has created a studyplan for all student
    And "school admin" has updated course duration for student

  Scenario Outline: Authentication for listing tag by student in course
    Given <course_statistical>a signed in "<role>"
    When user get list tag by student in course
    Then <course_statistical>returns "<msg>" status code

    Examples:
      | role           | msg              |
      | teacher        | OK               |
      | school admin   | OK               |
      | parent         | PermissionDenied |
      | student        | PermissionDenied |
      | hq staff       | OK               |
      | centre lead    | PermissionDenied |
      | centre manager | PermissionDenied |
      | teacher lead   | PermissionDenied |

  Scenario Outline: List tag by student in course
    Given <course_statistical>a signed in "school admin"
    And user creates tagged user
    When user get list tag by student in course
    Then our system must returns list tags correctly
    And <course_statistical>returns "OK" status code