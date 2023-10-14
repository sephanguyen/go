@quarantined
Feature: Upload existing old payment request file to cloud storage

    Scenario Outline: Job script upload existing payment files
        Given admin is logged-in back office on organization "<organization-id>"
        And there are "3" existing "<payment-status>" payments with payment method "CONVENIENCE STORE"
        And there are "3" existing "<payment-status>" payments with payment method "DIRECT DEBIT"
        And partner has existing convenience store master record
        And students has payment detail and billing address
        And there is an existing bank mapped to partner bank
        And students has payment and bank account detail
        And these payments belong to old payment request files
        And there is invoice date scheduled at day "-10"
        When an admin runs the upload payment request file job script
        Then the upload payment request file script has no error
        And the file_url of payment files is not empty
        And these payment files are uploaded successfully
        And the payment request files have a correct format

        Examples:
            | organization-id | payment-status |
            | -2147483648     | PENDING        |
            | -2147483642     | FAILED         |
            | -2147483639     | SUCCESSFUL     |
            | -2147483635     | PENDING        |

    Scenario Outline: Job script failed to upload existing payment files
        Given admin is logged-in back office on organization "<organization-id>"
        And there is "<payment-method>" payment request file that has no associated payment
        And there is invoice date scheduled at day "-10"
        When an admin runs the upload payment request file job script
        Then the upload payment request file script returns an error

        Examples:
            | organization-id | payment-method    |
            | -2147483645     | DIRECT DEBIT      |
            | -2147483643     | CONVENIENCE STORE |
