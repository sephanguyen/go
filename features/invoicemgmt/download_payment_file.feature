@quarantined
Feature: Download Payment Request
  As an HQ manager or admin
  I am able to download a payment request file

  Background:
    Given unleash feature flag is "disable" with feature name "BACKEND_Invoice_InvoiceManagement_CreatePaymentRequest_GCloud_File_Upload"

  Scenario Outline: Admin download a CSV payment file with Convenience Store payment method
    Given there are "<payment-count>" existing "PENDING" payments with payment method "CONVENIENCE STORE"
    And partner has existing convenience store master record
    And students has payment detail and billing address
    And these payments already belong to payment request file with payment method "CONVENIENCE STORE"
    When "<signed-in user>" logins to backoffice app
    And admin is at create payment request table
    And admin select and downloads the payment request file
    Then receives "OK" status code
    And the data byte returned is not empty
    And the payment request file has a correct CSV format

    # Since the step here use the create payment request endpoint,
    # the maximum payment-count that can be included here is 1000 to prevent error in CSV row count
    Examples:
      | signed-in user | payment-count |
      | school admin   | 10            |
      | hq staff       | 10            |

  Scenario Outline: Admin download a TXT payment file with Direct Debit payment method
    Given there are "<payment-count>" existing "PENDING" payments with payment method "DIRECT DEBIT"
    And there is an existing bank mapped to partner bank
    And students has payment and bank account detail
    And these payments already belong to payment request file with payment method "DIRECT DEBIT"
    When "<signed-in user>" logins to backoffice app
    And admin is at create payment request table
    And admin select and downloads the payment request file
    Then receives "OK" status code
    And the data byte returned is not empty
    And the payment request file has a correct bank TXT format

    Examples:
      | signed-in user | payment-count |
      | school admin   | 10            |
      | hq staff       | 10            |

  Scenario Outline: Sends a download payment request with empty file ID
    Given "<signed-in user>" logins to backoffice app
    When admin is at create payment request table
    And send a download file request with empty file ID
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |

  Scenario Outline: The payment file has no associated payments
    Given "<signed-in user>" logins to backoffice app
    And there is a payment file that has no associated payments
    When admin is at create payment request table
    And admin select and downloads the payment request file
    Then receives "Internal" status code

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |