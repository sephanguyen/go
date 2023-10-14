@blocker
Feature: Download Payment Request From Cloud Storage
    As an HQ manager or admin
    I am able to download a payment request file from cloud storage

    Background:
        Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_CreatePaymentRequest_GCloud_File_Upload"
        And unleash feature flag is "enable" with feature name "Invoice_InvoiceManagement_ConvenienceStoreCsvMessageFields"

    Scenario Outline: Admin download a CSV payment file from cloud storage with Convenience Store payment method
        Given there are "<payment-count>" existing "PENDING" payments with payment method "CONVENIENCE STORE"
        And partner has existing convenience store master record
        And students has payment detail and billing address
        And these payments already belong to payment request file with payment method "CONVENIENCE STORE"
        And there is an existing payment file in cloud storage
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

    Scenario Outline: Admin download a CSV payment file from cloud storage with Convenience Store payment method and invoice adjustment
        Given there are "<payment-count>" existing "PENDING" payments with payment method "CONVENIENCE STORE"
        And the invoices have invoice adjustment with amount "500"
        And partner has existing convenience store master record
        And students has payment detail and billing address
        And these payments already belong to payment request file with payment method "CONVENIENCE STORE"
        And there is an existing payment file in cloud storage
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
        And there is an existing payment file in cloud storage
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

    Scenario Outline: The payment file has no existing file in cloud storage
        Given there is a payment request file with payment method "<payment-method>" that is not in cloud storage
        When "<signed-in user>" logins to backoffice app
        And admin is at create payment request table
        And admin select and downloads the payment request file
        Then receives "Internal" status code

        Examples:
            | signed-in user | payment-method    |
            | school admin   | CONVENIENCE STORE |
            | hq staff       | DIRECT DEBIT      |

    Scenario Outline: Admin download a CSV payment file with Convenience Store payment method and adjustment type bill items
        Given there are "<payment-count>" existing "PENDING" payments with payment method "CONVENIENCE STORE"
        And "<all-or-one>" billing items of students are adjustment billing type
        And partner has existing convenience store master record
        And students has payment detail and billing address
        And these payments already belong to payment request file with payment method "CONVENIENCE STORE"
        And there is an existing payment file in cloud storage
        When "<signed-in user>" logins to backoffice app
        And admin is at create payment request table
        And admin select and downloads the payment request file
        Then receives "OK" status code
        And the data byte returned is not empty
        And the payment request file has a correct CSV format

        Examples:
            | signed-in user | payment-count | all-or-one |
            | school admin   | 3             | all        |
            | hq staff       | 2             | one        |
