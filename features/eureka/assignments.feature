Feature: Assignments

  Scenario: List student available contents
    Given some students are assigned some valid study plans
    When students list available contents
    Then returns a list of study plan items content

  Scenario: List student available contents with course id
    Given some students are assigned some valid study plans
    When students list available contents with "valid" course id
    Then returns a list of study plan items content

  Scenario: List student available contents with invalid course id
    Given some students are assigned some valid study plans
    When students list available contents with "invalid" course id
    Then returns a list of empty study plan items content

  Scenario: List student available contents with incorrect filters
    Given some students are assigned some valid study plans
    When students list available contents with incorrect filters
    Then returns a list of empty study plan items content

  Scenario: List active study plan items
    Given some students are assigned some valid study plans
    When student list active study plan items
    Then returns a list of "incompleted" active study plan items

  Scenario Outline: List completed study plan items
    Given some students are assigned some valid study plans
    And student submit their "existed" content assignment "<times>" times
    When student list completed study plan items
    Then returns a list of completed study plan items
    Examples:
        | times    |
        | single   |
        | multiple |

  Scenario: List overdue study plan items
    Given some students are assigned some valid study plans
    And student haven't completed any study plan items
    When student list overdue study plan items
    Then returns a list of overdue study plan items

  Scenario: List active study plan items of students with different courses
    Given some students of different courses are assigned some valid study plans
    When student list active study plan items
    Then returns a list of "incompleted" active study plan items

  Scenario Outline: List completed study plan items with empty start date
    Given some students are assigned some valid study plans
      And all study plan item has empty start date
    And student submit their "existed" content assignment "<times>" times
    When student list completed study plan items
    Then returns a list of completed study plan items
    Examples:
        | times    |
        | single   |
        | multiple |
