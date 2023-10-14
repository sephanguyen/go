Feature: Multiple Student Interact with own entry exit records

  Scenario: Only display student's entry & exit records within the organization
    Given "<student>" logins learner App with a resource path from "<organization>"
    And this student has existing entry and exit record
    When "<student>" visits its student's entry and exit record on learner App
    Then "<student>" only sees records from "<organization>"

    Examples:
      | student    | organization             |
      | student S1 | organization -2147483648 |
      | student S2 | organization -2147483646 |