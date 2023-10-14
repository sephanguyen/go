Feature: Upsert question group

  Background:
    Given <learning_objective>a signed in "school admin"
    And <learning_objective>a valid book content
    And a learning object

  Scenario Outline: Update display order for all quizzes quizset
    Given <learning_objective>a signed in "<role>"
    And a "valid" quiz
    And a "valid" quiz
    When update display order of index 0 and index 1
    Then <learning_objective>returns "<code>" status code
    And new display order is updated

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

  Scenario Outline: Update display order for quizzes and question group quizset
    Given <learning_objective>a signed in "<role>"
    And a "valid" quiz
    And insert a new question group
    And a "questionGroup" quiz
    When update display order of index 0 and index 1
    Then <learning_objective>returns "<code>" status code
    And new display order is updated

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

  Scenario Outline: Update display order for all question group quizset
    Given <learning_objective>a signed in "<role>"
    And insert a new question group
    And a "questionGroup" quiz
    And insert a new question group
    And a "questionGroup" quiz
    When update display order of index 0 and index 1
    Then <learning_objective>returns "<code>" status code
    And new display order is updated

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

  Scenario Outline: Update display order for quizzes inside question group quizset
    Given <learning_objective>a signed in "<role>"
    And insert a new question group
    And a "questionGroup" quiz
    And a "questionGroup" quiz
    When update display order of index 0 and index 1 in question group
    Then <learning_objective>returns "<code>" status code
    And new display order is updated

    Examples:
      | role         | code             |
      | hq staff     | OK               |
      | school admin | OK               |
      | teacher      | PermissionDenied |
      | student      | PermissionDenied |

