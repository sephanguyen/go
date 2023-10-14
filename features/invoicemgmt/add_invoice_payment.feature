@major
Feature: Add Payment
    As an HQ manager or admin
    I am able to add payment to an invoice

    Background:
        Given there is a student that has bill item with status "INVOICED"
        And bill item exists in invoicemgmt database
        And there is an existing invoice with "ISSUED" invoice status with bill item

    Scenario Outline: Admin add CONVENIENCE_STORE, CASH or BANK_TRANSFER payment to an invoice successfully
        Given "<signed-in user>" logins to backoffice app
        And there is "<payment-history>" payment history
        And admin adds payment to invoice
        And sets payment method to "<payment-method>" in add payment request
        And sets due date to "<due-date>" and expiry date to "<expiry-date>" in add payment request
        And sets amount same with invoice outstanding balance in add payment request
        When admin submits the add payment form with remarks "<remarks>"
        Then receives "OK" status code
        And payment exported tag is set to "false"
        And payment history is recorded with pending status
        And action log record is recorded with "PAYMENT_ADDED" action log type

        Examples:
            | signed-in user | payment-history | payment-method    | due-date | expiry-date | remarks        |
            | school admin   | none            | CONVENIENCE_STORE | TODAY+1  | TODAY+1     | sample-remarks |
            | school admin   | none            | CASH              | TODAY+2  | TODAY+3     | sample-remarks |
            | school admin   | none            | BANK_TRANSFER     | TODAY+3  | TODAY+3     | sample-remarks |
            | hq staff       | FAILED          | CONVENIENCE_STORE | TODAY+1  | TODAY+1     |                |
            | hq staff       | FAILED          | CASH              | TODAY+2  | TODAY+3     |                |
            | hq staff       | FAILED          | BANK_TRANSFER     | TODAY+3  | TODAY+3     |                |

    Scenario Outline: Admin add DIRECT_DEBIT payment to an invoice successfully
        Given "<signed-in user>" logins to backoffice app
        And there is "<payment-history>" payment history
        And there is an existing bank mapped to partner bank
        And this student has payment and bank account detail
        And admin adds payment to invoice
        And sets payment method to "DIRECT_DEBIT" in add payment request
        And sets due date to "<due-date>" and expiry date to "<expiry-date>" in add payment request
        And sets amount same with invoice outstanding balance in add payment request
        When admin submits the add payment form with remarks "<remarks>"
        Then receives "OK" status code
        And payment exported tag is set to "false"
        And payment history is recorded with pending status
        And action log record is recorded with "PAYMENT_ADDED" action log type

        Examples:
            | signed-in user | payment-history | due-date | expiry-date | remarks        |
            | school admin   | none            | TODAY+1  | TODAY+1     | sample-remarks |
            | hq staff       | FAILED          | TODAY+2  | TODAY+3     |                |

    Scenario Outline: Admin add payment for invoice that has invalid latest payment status
        Given "school admin" logins to backoffice app
        And there is "failed" payment history
        And there is "<payment-history>" payment history
        And admin adds payment to invoice
        And sets payment method to "CASH" in add payment request
        And sets due date to "TODAY+1" and expiry date to "TODAY+1" in add payment request
        And sets amount same with invoice outstanding balance in add payment request
        When admin submits the add payment form with remarks "sample-remarks"
        Then receives "InvalidArgument" status code

        Examples:
            | payment-history |
            | SUCCESSFUL      |
            | REFUNDED        |
            | PENDING         |

    Scenario Outline: Admin add DIRECT_DEBIT payment when student has no bank account detail
        Given "school admin" logins to backoffice app
        And admin adds payment to invoice
        And sets payment method to "DIRECT_DEBIT" in add payment request
        And sets due date to "TODAY+1" and expiry date to "TODAY+1" in add payment request
        And sets amount same with invoice outstanding balance in add payment request
        When admin submits the add payment form with remarks "sample-remarks"
        Then receives "InvalidArgument" status code

    Scenario Outline: Admin add DIRECT_DEBIT payment when student bank account is unverified
        Given "school admin" logins to backoffice app
        And there is an existing bank mapped to partner bank
        And this student has payment and bank account detail
        And this student bank account is not verified
        And admin adds payment to invoice
        And sets payment method to "DIRECT_DEBIT" in add payment request
        And sets due date to "TODAY+1" and expiry date to "TODAY+1" in add payment request
        And sets amount same with invoice outstanding balance in add payment request
        When admin submits the add payment form with remarks "sample-remarks"
        Then receives "InvalidArgument" status code
