@blocker
Feature: Update student payment method
  As an HQ manager, admin, center manager and staff
  I am able to update payment method of a student

  # selecting payment method button is only enabled if bank account is verified
  Scenario Outline: Admin updates student payment method successfully
    Given an existing student with student payment "billing address and bank account" info
    And student bank account is set to "verified" status
    And student payment detail has "<existing-payment-method>" payment method
    When "<signed-in user>" logins to backoffice app
    And updates "<payment-method>" payment method of the student
    Then receives "OK" status code
    And student payment method is updated successfully to "<payment-method>"
    And student payment information updated successfully with "UPDATED_PAYMENT_METHOD" student payment detail action log record

    Examples:
      | signed-in user | existing-payment-method | payment-method    |
      | school admin   | DIRECT_DEBIT            | CONVENIENCE_STORE |
      | hq staff       | CONVENIENCE_STORE       | DIRECT_DEBIT      |
      | centre staff   | DIRECT_DEBIT            | CONVENIENCE_STORE |
      | centre manager | CONVENIENCE_STORE       | DIRECT_DEBIT      |

  Scenario Outline: Admin updates student payment method unsuccessfully with unverified bank account
    Given an existing student with student payment "billing address and bank account" info
    And student bank account is set to "unverified" status
    And student payment detail has "<existing-payment-method>" payment method
    When "<signed-in user>" logins to backoffice app
    And updates "<payment-method>" payment method of the student
    Then receives "Internal" status code

    Examples:
      | signed-in user | existing-payment-method | payment-method    |
      | school admin   | DIRECT_DEBIT            | CONVENIENCE_STORE |
      | hq staff       | CONVENIENCE_STORE       | DIRECT_DEBIT      |
      | centre staff   | DIRECT_DEBIT            | CONVENIENCE_STORE |
      | centre manager | CONVENIENCE_STORE       | DIRECT_DEBIT      |


  Scenario Outline: Admin cannot update student payment method with invalid student payment records
    Given an existing student with student payment "<invalid-student-payment-info>" info
    When "<signed-in user>" logins to backoffice app
    And updates "<payment-method>" payment method of the student
    Then receives "Internal" status code

    Examples:
      | signed-in user | payment-method    | invalid-student-payment-info        |
      | school admin   | CONVENIENCE_STORE | non existing student payment detail |
      | hq staff       | DIRECT_DEBIT      | no billing address detail           |
      | centre staff   | CONVENIENCE_STORE | billing address                     |
