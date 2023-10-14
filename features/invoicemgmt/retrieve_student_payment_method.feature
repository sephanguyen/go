@blocker
Feature: Retrieve Student Payment Method Record
    Payment service method to retrieve student payment record that will be use for creating order
    # Currently we only support Convenience Store and Direct Debit for Student's default payment method

  Scenario Outline: Retrieves student with with default payment method record successfully
    Given an existing student with default payment method "<payment-method>"
    And "<signed-in user>" logins to backoffice app
    When the RetrieveStudentPaymentMethod endpoint is called for this student
    Then receives "OK" status code
    And "<payment-method>" payment method for this student is retrieve successfully

    Examples:
      | signed-in user | payment-method    |
      | school admin   | CONVENIENCE_STORE |
      | hq staff       | DIRECT_DEBIT      |
      | centre staff   | DIRECT_DEBIT      |
      | centre manager | CONVENIENCE_STORE |

  Scenario: Retrieves student with no default payment method record successfully
    Given an existing student with no student payment method
    And "<signed-in user>" logins to backoffice app
    When the RetrieveStudentPaymentMethod endpoint is called for this student
    Then receives "OK" status code
    And empty payment method for this student is retrieve successfully

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |
      | centre staff   |
      | centre manager |
