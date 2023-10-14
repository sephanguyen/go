Feature: List Course By Locations
  Background:
    Given "staff granted role school admin" signin system
    And a random number
    And some class members
    And a list of courses are existed in DB of "manabie"
    And some centers
    And a list of course_access_paths are existed in DB
    And all courses are belong to "current" academicYear

  Scenario Outline: teacher list courses of school
    Given "staff granted role teacher" signin system
    And above teacher belong to "manabie" school
    And a ListCoursesByLocationsRequest message "<message>" school and keyword "course" with <number> locations
    When user list courses by locations
    Then returns "OK" status code
    And returns courses in ListCoursesResponse matching filter of ListCoursesRequest
    And locations of courses matching with course_access_paths with <number> locations

    Examples:
      | number           | message       |
      | 0                | manabie       |
      | 1                | manabie       |
      | 2                | empty filter  |


