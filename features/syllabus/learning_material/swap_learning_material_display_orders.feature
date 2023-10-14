Feature: Swap learning material display orders
  Background:
    Given <learning_material>a signed in "school admin"
    And <learning_material>a valid book content
    And some existing learning materials in an arbitrary topic of the book

  Scenario Outline: authenticate <role> swap display order of two learning material
    Given <learning_material>a signed in "<role>"
    When user swap LM display order
    Then <learning_material>returns "<msg>" status code
    Examples:
      | role           | msg              |
      | parent         | PermissionDenied |
      | student        | PermissionDenied |
      | school admin   | OK               |
      | hq staff       | OK               |
      | teacher        | PermissionDenied |
      | centre lead    | PermissionDenied |
      | centre manager | PermissionDenied |
      | teacher lead   | PermissionDenied |
  # this scenario, use random to make fair, insrease the coverage
  Scenario: Swap LM's display orders
    When user swap LM display order
    Then our system must swap display orders of learning material correctly

