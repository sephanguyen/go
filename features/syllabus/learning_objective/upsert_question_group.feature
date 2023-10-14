Feature: Upsert question group

  Background:
    Given <learning_objective>a signed in "school admin"
    And <learning_objective>a valid book content
    And a learning object

  Scenario Outline: Insert a new question group when already have quiz ids and question ids
    Given <learning_objective>a signed in "<role>"
    And a "valid" quiz
    And existing question group
    When insert a new question group
    Then <learning_objective>returns "<code>" status code
    And new question group was added at the end of list

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

  Scenario Outline: Insert a new question group when have no any quiz or question
    Given <learning_objective>a signed in "<role>"
    When insert a new question group
    Then <learning_objective>returns "<code>" status code
    And new question group was added

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

  Scenario Outline: Update a question group when already have quiz ids and question ids
    Given <learning_objective>a signed in "<role>"
    And a "valid" quiz
    And existing question group
    When update a question group
    Then <learning_objective>returns "<code>" status code
    And question group was updated

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

