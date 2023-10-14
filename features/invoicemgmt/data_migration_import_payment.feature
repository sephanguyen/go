@quarantined
Feature: Data Migration Import Payment
  As an HQ manager or admin
  I am able to import data migration for payment on master management

  Scenario Outline: Admin successfully imports payment csv file successfully
    Given there are "<invoice-count>" migrated invoices with "<invoice-status>" status
    And there is a payment csv file with "<payment-method>" payment method for these invoices
    And this payment csv file has payment data with "<payment-status>" payment status
    When "<signed-in user>" logins to backoffice app 
    And imports the payment csv file
    Then receives "OK" status code 
    And payment csv file is imported successfully
    And there are payment records with correct invoice created successfully

    Examples:
      | signed-in user   | invoice-count | invoice-status | payment-method    | payment-status     |
      | school admin     | 2             | ISSUED         | CASH              | PAYMENT_PENDING    |
      | hq staff         | 1             | PAID           | CONVENIENCE_STORE | PAYMENT_SUCCESSFUL |
      | hq staff         | 1             | REFUNDED       | BANK_TRANSFER     | PAYMENT_SUCCESSFUL |
      | school admin     | 1             | FAILED         | DIRECT_DEBIT      | PAYMENT_FAILED     |

  Scenario Outline: Admin failed to import payment csv file with payment status not valid
    Given there are "<invoice-count>" migrated invoices with "<invoice-status>" status
    And there is a payment csv file with "<payment-method>" payment method for these invoices
    And this payment csv file has payment data with "<payment-invalid-status>" payment status
    When "<signed-in user>" logins to backoffice app 
    And imports the payment csv file
    Then receives "OK" status code 
    And payment csv file is imported unsuccessfully
    And response has error on "<payment-invalid-status>" payment status that should be "<payment-valid-status>"

    Examples:
      | signed-in user   | invoice-count | invoice-status | payment-method    | payment-invalid-status | payment-valid-status |
      | school admin     | 1             | ISSUED         | CASH              | PAYMENT_FAILED         | PAYMENT_PENDING      | 
      | hq staff         | 1             | PAID           | CONVENIENCE_STORE | PAYMENT_FAILED         | PAYMENT_SUCCESSFUL   |
      | hq staff         | 1             | REFUNDED       | BANK_TRANSFER     | PAYMENT_FAILED         | PAYMENT_SUCCESSFUL   |
      | school admin     | 1             | FAILED         | DIRECT_DEBIT      | PAYMENT_PENDING        | PAYMENT_FAILED       |

  Scenario Outline: Admin failed to import payment csv file with invalid student
    Given there are "<invoice-count>" migrated invoices with "<invoice-status>" status
    And there is a payment csv file with "<payment-method>" payment method for these invoices
    And payment csv file contains invalid students
    And this payment csv file has payment data with "<payment-status>" payment status
    When "<signed-in user>" logins to backoffice app 
    And imports the payment csv file
    Then receives "OK" status code 
    And payment csv file is imported unsuccessfully
    And response has error for invalid payment student

    Examples:
      | signed-in user   | invoice-count | invoice-status | payment-method    | payment-status     |
      | school admin     | 1             | ISSUED         | CASH              | PAYMENT_PENDING    |
      | hq staff         | 1             | PAID           | CONVENIENCE_STORE | PAYMENT_SUCCESSFUL |
      | hq staff         | 1             | REFUNDED       | BANK_TRANSFER     | PAYMENT_SUCCESSFUL |
      | school admin     | 1             | FAILED         | DIRECT_DEBIT      | PAYMENT_FAILED     |
