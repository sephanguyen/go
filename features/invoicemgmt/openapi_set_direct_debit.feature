@major
Feature: OpenAPI Set Direct Debit
  As a school staff
  I need to be able to create/update bank information of a student

  Background:
    Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_SetDirectDebit"

  Scenario: Upsert bank info of student that has no initial bank account successfully
    Given an existing student with student payment "billing address" info
    And there are existing bank and bank branch
    And this student is included on bank OpenAPI valid payload
    And admin already setup an api user
    When admin submits the bank OpenAPI payload
    Then bank info of the student was upserted successfully by OpenAPI
    And no student payment detail action log recorded

  Scenario: Upsert bank info of student that has initial bank account successfully
    Given an existing student with student payment "billing address and bank account" info
    And there are existing bank and bank branch
    And this student is included on bank OpenAPI valid payload
    And admin already setup an api user
    When admin submits the bank OpenAPI payload
    Then bank info of the student was upserted successfully by OpenAPI
    And student payment information updated successfully with "UPDATED_BANK_DETAILS" student payment detail action log record

  Scenario Outline: Upsert bank info of student unsuccessfully
    Given an existing student with student payment "billing address" info
    And there are existing bank and bank branch
    And this student is included on bank OpenAPI invalid "<conditions>" payload
    And admin already setup an api user
    When admin submits the bank OpenAPI payload
    Then receives failed "<response-code>" response code from OpenAPI

    Examples:
      | conditions                                     | response-code |
      | internal error                                 | 50000         |
      | missing field                                  | 40001         |
      | invalid public key                             | 40302         |
      | invalid private key                            | 40301         |
      | invalid account number with verified account   | 40004         |
      | invalid account number with unverified account | 40004         |
      | invalid account type with verified account     | 40004         |
      | invalid account type with unverified account   | 40004         |


