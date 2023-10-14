Feature: Retrieve study plan progress

  Scenario: Retrieve study plan progress
    Given some students are assigned some valid study plans
    And student submit their "exited" content assignment "<times>" times
    When teacher retrieves study plan progress of students
    Then returns study plan progress of students correctly
    Examples:
        | times    |
        | single   |
        | multiple |

  Scenario Outline: Retrieve study plan progress
    Given some students are assigned some study plan with available from "<time>"
    When teacher retrieves study plan progress of students
    Then returns study plan progress of students are 0
    Examples:
      | time                      |
      | empty                     |
      | 2906-01-02                |

