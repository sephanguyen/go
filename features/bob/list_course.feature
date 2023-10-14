Feature: List Course
  Background:
    Given "staff granted role school admin" signin system
    And a random number
    And a list of courses are existed in DB of "manabie"
    And all courses are belong to "current" academicYear

  Scenario Outline: user of invalid roles try to list courses
    Given "<signed as>" signin system
    And a ListCoursesRequest message "manabie" school
    When user list courses
    Then returns "<msg>" status code
    Examples:
      | signed as                   | msg              |
      | unauthenticated             | Unauthenticated  |
      | staff granted role hq staff | PermissionDenied |

  Scenario: teacher list courses of school
    Given "staff granted role teacher" signin system
    And above teacher belong to "manabie" school
    And a ListCoursesRequest message "empty filter" school
    When user list courses
    Then returns "OK" status code
    And returns courses in ListCoursesResponse matching filter of ListCoursesRequest

  Scenario: student list courses of school
    Given "student" signin system
    And a ListCoursesRequest message "manabie" school
    When user list courses
    Then returns "OK" status code
    And returns response for student list courses have to correctly