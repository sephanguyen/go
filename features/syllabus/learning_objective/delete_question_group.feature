Feature: Delete question group

  Background:
    Given <learning_objective>a signed in "school admin"
    And <learning_objective>a valid book content
    And a learning object

  Scenario Outline: Delete question group that not contains questions
    Given <learning_objective>a signed in "<role>"
    And a "valid" quiz
    And existing question group
    When delete existing question group
    Then <learning_objective>returns "<code>" status code
    And question group is deleted

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

  Scenario Outline: Delete question group that contains questions
    Given <learning_objective>a signed in "<role>"
    And a "valid" quiz
    And existing question group
    And a "questionGroup" quiz
    When delete existing question group
    Then <learning_objective>returns "<code>" status code
    And question group is deleted

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

    Scenario Outline: Delete question group in quizset that just have only empty question group
    Given <learning_objective>a signed in "<role>"
    And existing question group
    When delete existing question group
    Then <learning_objective>returns "<code>" status code
    And question group is deleted

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |
