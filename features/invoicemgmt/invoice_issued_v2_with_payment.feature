@major
Feature: Issue Invoice without payment
    As an HQ manager or admin
    I am able to update the status of the status of the invoice to INVOICE_ISSUED and will not create a payment

    Background:
        Given there is an existing invoice
        And unleash feature flag is "enable" with feature name "Invoice_InvoiceManagement_BackOffice_SingleIssueInvoiceWithPayment"

    Scenario Outline: Admin issues an invoice succesfully using CASH/BANK_TRANSFER/CONVENIENCE_STORE payment method
        Given invoice has draft invoice status
        When "<signed-in user>" issues invoice using v2 endpoint with "<payment-method>" payment method and dates "<due-date>" "<expiry-date>"
        Then receives "OK" status code
        And invoice status is updated to "ISSUED" status
        And invoice exported tag is set to "false"
        And payment exported tag is set to "false"
        And payment history is recorded with pending status
        And action log record is recorded with "INVOICE_ISSUED" action log type
        And action log record is recorded with "PAYMENT_ADDED" action log type

        Examples:
            | signed-in user | payment-method    | due-date | expiry-date |
            | school admin   | CONVENIENCE_STORE | TODAY    | TODAY+1     |
            | school admin   | CASH              | TODAY    | TODAY+1     |
            | hq staff       | BANK_TRANSFER     | TODAY    | TODAY+1     |

    Scenario Outline: Admin issues an invoice succesfully using DIRECT_DEBIT payment method
        Given invoice has draft invoice status
        And this student has payment and bank account detail
        And student bank account is set to "verified" status
        When "<signed-in user>" issues invoice using v2 endpoint with "DIRECT_DEBIT" payment method and dates "<due-date>" "<expiry-date>"
        Then receives "OK" status code
        And invoice status is updated to "ISSUED" status
        And invoice exported tag is set to "false"
        And payment exported tag is set to "false"
        And payment history is recorded with pending status
        And action log record is recorded with "INVOICE_ISSUED" action log type
        And action log record is recorded with "PAYMENT_ADDED" action log type

        Examples:
            | signed-in user | due-date | expiry-date |
            | school admin   | TODAY    | TODAY+1     |
            | hq staff       | TODAY    | TODAY+1     |

    Scenario Outline: Admin issues an invoice with DIRECT_DEBIT payment method with unverified bank account
        Given invoice has draft invoice status
        And this student has payment and bank account detail
        And student bank account is set to "not verified" status
        When "school admin" issues invoice using v2 endpoint with "DIRECT_DEBIT" payment method and dates "<due-date>" "<expiry-date>"
        Then receives "InvalidArgument" status code

    Scenario Outline: Admin issues an invoice that has negative amount succesfully
        Given invoice has draft invoice status
        And this invoice has "-100" total amount
        When "school admin" issues invoice using v2 endpoint
        Then receives "OK" status code
        And invoice status is updated to "ISSUED" status
        And this invoice has "no existing" payment with "PENDING" status
        And action log record is recorded with "INVOICE_ISSUED" action log type

    Scenario Outline: Admin issues an invoice that has zero amount and update invoice status to PAID succesfully
        Given invoice has draft invoice status
        And this invoice has "0" total amount
        When "school admin" issues invoice using v2 endpoint
        Then receives "OK" status code
        And invoice status is updated to "PAID" status
        And this invoice has "no existing" payment with "PENDING" status
        And action log record is recorded with "INVOICE_ISSUED" action log type