@quarantined
# Deprioritized 
# If continued, new permission_role record required to associate payment.invoice.read to Parent role
Feature: Retrieve Invoice records
  As a parent
  I am able to see invoice records of my children on learner app
  Background:
    Given "parent" logins Learner App
    And this parent has an existing student

  Scenario: Parent selects student with no invoice records
    Given student has "no existing" invoice records
    When parent is at the invoice list screen
    And parent selects this existing student
    Then no records found displayed successfully
    And receives "OK" status code

  Scenario Outline: Parent selects student with invoice records
    Given student has "<invoices>" invoice records
    When parent is at the invoice list screen
    And parent selects this existing student
    Then records found with default limit are displayed successfully
    And no invoice draft records found
    And receives "OK" status code

    Examples:
        | invoices                    |
        | ISSUED-VOID-PAID-DRAFT      |
        | ISSUED-DRAFT-VOID-VOID      |
        | PAID-ISSUED-REFUNDED-VOID   |
        | PAID-FAILED-DRAFT-VOID      |
        | ISSUED-VOID-REFUNDED-FAILED |

  Scenario Outline: Parent checks all the invoice records of the student
      Given student has "<invoices>" invoice records
      And parent is at the invoice list screen
      And parent selects this existing student
      And records found with default limit are displayed successfully
      When parent scrolls down to display all records
      Then all records found are displayed successfully
      And no invoice draft records found
      And receives "OK" status code

    Examples:
        | invoices                         |
        | ISSUED-VOID-PAID-DRAFT-PAID      |
        | ISSUED-DRAFT-VOID-VOID-PAID      |
        | PAID-ISSUED-REFUNDED-VOID-FAILED |
        | PAID-FAILED-DRAFT-VOID-PAID      |
        | ISSUED-VOID-REFUNDED-FAILED-VOID |

  Scenario Outline: Parent selects another student and invoice records displayed successfully
    Given parent has another existing student
    And this student has "<invoices>" invoice records
    And parent is at the invoice list screen
    When parent selects this existing student
    Then records found with default limit are displayed successfully
    And no invoice draft records found
    And receives "OK" status code

    Examples:
      | invoices                    |
      | ISSUED-VOID-PAID-DRAFT      |
      | ISSUED-DRAFT-VOID-VOID      |
      | PAID-ISSUED-REFUNDED-VOID   |
      | PAID-FAILED-DRAFT-VOID      |
      | ISSUED-VOID-REFUNDED-FAILED |