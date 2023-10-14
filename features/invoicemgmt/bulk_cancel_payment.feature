@major
Feature: Bulk Cancel Payment
    As an HQ manager or admin
    I can bulk cancel the payment that has PENDING status

    Scenario Outline: Admin bulk cancel payment succesfully
        Given there are "<payment-count>" existing "PENDING" payments with payment method "CONVENIENCE STORE"
        And these "<payment-count>" payments belong to a bulk payment
        And "<signed-in user>" logins to backoffice app
        When admin cancel the bulk payment
        Then receives "OK" status code
        And bulk payment record status is updated to "BULK_PAYMENT_CANCELLED"
        And each payments has "PAYMENT_FAILED" payment status
        And action log record for each invoice is recorded with "PAYMENT_CANCELLED" action log type

        Examples:
            | signed-in user | payment-count |
            | school admin   | 5             |
            | hq staff       | 5             |

    Scenario Outline: Admin bulk cancel only the PENDING payments
        Given there are "<payment-count>" existing "PENDING" payments with payment method "CONVENIENCE STORE"
        And there are "<payment-diff-status-count>" existing "<payment-diff-status>" payments with payment method "CONVENIENCE STORE"
        And these "<payments-in-bulk-count>" payments belong to a bulk payment
        And "<signed-in user>" logins to backoffice app
        When admin cancel the bulk payment
        Then receives "OK" status code
        And bulk payment record status is updated to "BULK_PAYMENT_CANCELLED"
        And only pending payments were updated to "PAYMENT_FAILED" payment status
        And only pending payments invoice are recorded with "PAYMENT_CANCELLED" action log type

        Examples:
            | signed-in user | payment-count | payment-diff-status-count | payments-in-bulk-count | payment-diff-status |
            | school admin   | 5             | 3                         | 8                      | FAILED              |
            | hq staff       | 5             | 3                         | 8                      | SUCCESSFUL          |
            | hq staff       | 5             | 3                         | 8                      | REFUNDED            |

    Scenario Outline: Admin cannot bulk cancel payment since there is one payment that is already exported
        Given there are "<payment-count>" existing "PENDING" payments with payment method "DIRECT DEBIT"
        And these "<payment-count>" payments belong to a bulk payment
        And a payment is already exported
        And "<signed-in user>" logins to backoffice app
        When admin cancel the bulk payment
        Then receives "InvalidArgument" status code

        Examples:
            | signed-in user | payment-count |
            | school admin   | 5             |
            | hq staff       | 5             |

    Scenario Outline: Admin cannot bulk cancel payment since the bulk payment status is not PENDING
        Given there are "<payment-count>" existing "PENDING" payments with payment method "DIRECT DEBIT"
        And these "<payment-count>" payments belong to a bulk payment
        And this bulk payment has status "<bulk-payment-status>"
        And "<signed-in user>" logins to backoffice app
        When admin cancel the bulk payment
        Then receives "InvalidArgument" status code

        Examples:
            | signed-in user | payment-count | bulk-payment-status    |
            | school admin   | 5             | BULK_PAYMENT_CANCELLED |
            | hq staff       | 5             | BULK_PAYMENT_EXPORTED  |

    Scenario Outline: Admin bulk cancel payments that are already manually cancelled
        Given there are "<payment-count>" existing "FAILED" payments with payment method "CONVENIENCE STORE"
        And these "<payment-count>" payments belong to a bulk payment
        And "<signed-in user>" logins to backoffice app
        When admin cancel the bulk payment
        Then receives "OK" status code
        And bulk payment record status is updated to "BULK_PAYMENT_CANCELLED"
        And each payments has "PAYMENT_FAILED" payment status
        And no invoice action log with "PAYMENT_CANCELLED" action log type recorded for each invoice

        Examples:
            | signed-in user | payment-count |
            | school admin   | 10            |
            | hq staff       | 5             |