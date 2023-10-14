@quarantined
Feature: Issue Invoice
  As an HQ manager or admin
  I am able to update the status of the status of the invoice and create payment records
  Background:
    Given there is an existing invoice

  Scenario Outline: Admin issues an invoice successfully
    Given invoice has draft invoice status
    When "<signed-in user>" issues invoice with "<payment-method>" payment method
    Then receives "OK" status code
    And invoice status is updated to "ISSUED" status
    And invoice exported tag is set to "false"
    And payment exported tag is set to "false"
    And payment history is recorded with pending status
    And action log record is recorded with "INVOICE_ISSUED" action log type

    Examples:
      | signed-in user | payment-method    |
      | school admin   | CONVENIENCE_STORE |
      | school admin   | CONVENIENCE_STORE |
      | school admin   | CASH              |
      | hq staff       | CONVENIENCE_STORE |
      | hq staff       | BANK_TRANSFER     |

  Scenario Outline: Admin failed to issue an invoice with invalid payment method
    Given invoice has draft invoice status
    When "<signed-in user>" issues invoice with "<payment-method>" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-method |
      | school admin   | DIRECT_DEBIT   |
      | hq staff       | DIRECT_DEBIT   |

  Scenario Outline: Admin failed to issue an invoice with non-existing invoice ID
    Given invoice ID is non-existing
    When "<signed-in user>" issues invoice with "<payment-method>" payment method
    Then receives "Internal" status code
    And no payment history is recorded

    Examples:
      | signed-in user | payment-method    |
      | school admin   | CASH              |
      | hq staff       | CONVENIENCE_STORE |
