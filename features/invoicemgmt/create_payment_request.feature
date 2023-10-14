Feature: Create Payment Request
  As an HQ manager or admin
  I am able to create payment request

  Background:
    Given unleash feature flag is "disable" with feature name "BACKEND_Invoice_InvoiceManagement_CreatePaymentRequest_GCloud_File_Upload"

  @quarantined
  Scenario Outline: Admin create a payment request with Convenience Store as payment method
    Given there are "<payment-count>" existing "PENDING" payments with payment method "CONVENIENCE STORE"
    And partner has existing convenience store master record
    And students has payment detail and billing address
    And "<signed-in user>" logins to backoffice app
    And admin is at create payment request modal
    And admin chooses "CONVENIENCE STORE" as payment method
    And admin adds payment due date from at day "<payment-due-date-from>" and due date until at day "<payment-due-date-until>"
    When admin clicks save create payment request
    Then receives "OK" status code
    And the payments are associated to a payment request file
    And there are "<payment-file-generated-count>" payment file with correct file name saved on database
    And the payments and invoices isExported field was set to "true"

    # The value of payment-due-date-from and payment-due-date-until means that it is the number of days before/after today
    # If the value is positive, it will be the days after today
    # If negative, it will be the days before today
    # If 0, it means today

    # Limits the payment count to less than 1k since high data count leads to kafka sync delay in CI that causes flaky error
    Examples:
      | signed-in user | payment-count | payment-file-generated-count | payment-due-date-from | payment-due-date-until |
      | school admin   | 10            | 1                            | 0                     | 1                      |
      | hq staff       | 10            | 1                            | 2                     | 2                      |

  @quarantined
  Scenario Outline: Admin create a payment request with Convenience Store with already exported payment
    Given there are "3" existing "PENDING" payments with payment method "CONVENIENCE STORE"
    And a payment is already exported
    And partner has existing convenience store master record
    And students has payment detail and billing address
    And "<signed-in user>" logins to backoffice app
    And admin is at create payment request modal
    And admin chooses "CONVENIENCE STORE" as payment method
    And admin adds payment due date from at day "1" and due date until at day "2"
    When admin clicks save create payment request
    Then receives "Internal" status code

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |

  @quarantined
  Scenario Outline: Admin create a payment request with Direct Debit as payment method
    Given there are "<payment-count>" existing "PENDING" payments with payment method "DIRECT DEBIT"
    And there is an existing bank mapped to partner bank
    And students has payment and bank account detail
    And "<signed-in user>" logins to backoffice app
    And admin is at create payment request modal
    And admin chooses "DIRECT DEBIT" as payment method
    And admin adds payment due date at day "<payment-due-date>"
    When admin clicks save create payment request
    Then receives "OK" status code
    And the payments are associated to a payment request file
    And there are "<payment-file-generated-count>" payment file with correct file name saved on database
    And the payments and invoices isExported field was set to "true"
    And students have new customer code history record

    # The value of payment-due means that it is the number of days before/after today
    # If the value is positive, it will be the days after today
    # If negative, it will be the days before today
    # If 0, it means today

    Examples:
      | signed-in user | payment-count | payment-file-generated-count | payment-due-date |
      | school admin   | 10            | 1                            | 0                |
      | hq staff       | 10            | 1                            | 2                |

  @quarantined
  Scenario Outline: Admin create a payment request with Direct Debit as payment method and there are multiple bank mapping
    Given there are "<payment-count>" existing "PENDING" payments with payment method "DIRECT DEBIT"
    And there are banks mapped to different partner bank
    And students have bank account in either of the banks
    And "<signed-in user>" logins to backoffice app
    And admin is at create payment request modal
    And admin chooses "DIRECT DEBIT" as payment method
    And admin adds payment due date at day "<payment-due-date>"
    When admin clicks save create payment request
    Then receives "OK" status code
    And the payments are associated to a payment request file
    And there are "<payment-file-generated-count>" payment file with correct file name saved on database
    And the payments and invoices isExported field was set to "true"
    And students have new customer code history record

    # The value of payment-due means that it is the number of days before/after today
    # If the value is positive, it will be the days after today
    # If negative, it will be the days before today
    # If 0, it means today

    Examples:
      | signed-in user | payment-count | payment-file-generated-count | payment-due-date |
      | school admin   | 10            | 2                            | 0                |
      | hq staff       | 10            | 2                            | 2                |

  @quarantined
  Scenario Outline: Admin create a payment request with Direct Debit as payment method with partner bank record limit
    Given unleash feature flag is "enable" with feature name "Invoice_InvoiceManagement_BackOffice_BulkAddAndValidatePayments"
    And there are "<payment-count>" existing "PENDING" payments with payment method "DIRECT DEBIT"
    And there is an existing bank mapped to partner bank
    And this partner bank record limit is "2"
    And students has payment and bank account detail
    And "<signed-in user>" logins to backoffice app
    And admin is at create payment request modal
    And admin chooses "DIRECT DEBIT" as payment method
    And admin adds payment due date at day "<payment-due-date>"
    When admin clicks save create payment request
    Then receives "OK" status code
    And the payments are associated to a payment request file
    And there are "<payment-file-generated-count>" payment file with correct file name saved on database
    And the payments and invoices isExported field was set to "true"
    And students have new customer code history record

    # The value of payment-due means that it is the number of days before/after today
    # If the value is positive, it will be the days after today
    # If negative, it will be the days before today
    # If 0, it means today

    Examples:
      | signed-in user | payment-count | payment-file-generated-count | payment-due-date |
      | school admin   | 10            | 5                            | 0                |
      | hq staff       | 10            | 5                            | 2                |
