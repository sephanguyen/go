Feature: Multiple Parent Interact with student entry exit records

  Scenario: Display entry exit records on learner app from parent with different organization
    Given "<parent>" logins learner App with a resource path from "<organization>"
    And this parent has existing student with entry and exit record
    When "<parent>" visits its student's entry and exit record on learner App
    Then "<parent>" only sees records from "<organization>"

    Examples:
      | parent    | organization             |
      | parent P1 | organization -2147483648 |
      | parent P2 | organization -2147483646 |