@quarantined
Feature: Update billing address info when student home address was updated

    Background:
        Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_AutoSetConvenienceStore"
        And invoicemgmt internal config "enable_auto_default_convenience_store" is "on"

    Scenario: Billing address is updated when student's home address was updated
        Given there is an existing student with "existing" billing address and "existing" payment detail
        When yasuo send the update student event request with "complete updated user address info and updated payer name"
        Then student payer name successfully updated
        And student billing address record is successfully updated
        And student payment information updated successfully with "UPDATED_BILLING_DETAILS" student payment detail action log record

    Scenario: Billing address is created while payer name was updated
        Given there is an existing student with "no existing" billing address and "existing" payment detail
        When yasuo send the update student event request with "complete updated user address info and updated payer name"
        Then student payer name successfully updated
        And student billing address record is successfully created
        And student payment information updated successfully with "UPDATED_BILLING_DETAILS" student payment detail action log record

    Scenario: Student's billing address was created when student's home address was updated
        Given there is an existing student with "no existing" billing address and "no existing" payment detail
        When yasuo send the update student event request with "complete updated user address info and updated payer name"
        Then student payment detail record is successfully created
        And student billing address record is successfully created
        And no student payment detail action log recorded

    Scenario: Student's payer name is not updated
        Given there is an existing student with "no existing" billing address and "existing" payment detail
        When yasuo send the update student event request with "updated payer name"
        Then student payer name is not updated

    Scenario: No billing address was updated or created
        Given there is an existing student with "no existing" billing address and "no existing" payment detail
        When yasuo send the update student event request with "missing user address"
        Then no student payment detail record created
        And no student billing address record created

    Scenario Outline: Student's billing address was removed
        Given there is an existing student with "existing" billing address and "existing" payment detail
        When yasuo send the update student event request with "<condition>"
        Then student default payment method including payer name is successfully removed
        And student billing address record is successfully removed
        And student payment information updated successfully with "UPDATED_BILLING_DETAILS" student payment detail action log record

        Examples:
            | condition                                    |
            | missing user address                         |
            | one important billing address field is empty |
            | all important billing address field is empty |

    Scenario: Payment method was re-added when student initial payment method is empty and student has bank account
        Given there is an existing student with "existing" billing address and "existing" payment detail
        And the default payment method of this student is "EMPTY"
        And there is an existing bank mapped to partner bank
        And this student has bank account with verification "<verified-status>" status
        When yasuo send the update student event request with "complete updated user address info and updated payer name"
        Then student payer name successfully updated
        And student billing address record is successfully updated
        And student payment information updated successfully with "UPDATED_BILLING_DETAILS" student payment detail action log record
        And the student default payment method was set to "<expected-payment-method>"

        Examples:
            | verified-status | expected-payment-method |
            | verified        | DIRECT_DEBIT            |
            | not verified    | CONVENIENCE_STORE       |

    Scenario: Payment method was re-added when student initial payment method is empty
        Given there is an existing student with "existing" billing address and "existing" payment detail
        And the default payment method of this student is "EMPTY"
        When yasuo send the update student event request with "complete updated user address info and updated payer name"
        Then student payer name successfully updated
        And student billing address record is successfully updated
        And student payment information updated successfully with "UPDATED_BILLING_DETAILS" student payment detail action log record
        And the student default payment method was set to "CONVENIENCE_STORE"

    Scenario: Payment method was re-added when student initial payment method is empty and has empty billing details
        Given there is an existing student with "existing" billing address and "existing" payment detail
        And the default payment method of this student is "EMPTY"
        And this student billing address was removed
        And there is an existing bank mapped to partner bank
        And this student has bank account with verification "verified" status
        When yasuo send the update student event request with "complete updated user address info and updated payer name"
        Then student payer name successfully updated
        And student billing address record is successfully updated
        And student payment information updated successfully with "UPDATED_BILLING_DETAILS" student payment detail action log record
        And the student default payment method was set to "DIRECT_DEBIT"