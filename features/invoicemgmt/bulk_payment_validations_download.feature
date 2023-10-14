@major
Feature: Bulk Payment Validations Download
  As an HQ manager or admin
  I am able to download payment validation file

  Scenario Outline: Admin downloads payment validation file successfully
    Given there is existing bulk payment validation record for "<payment-method>"
    And this record consists of "<success-validated-count>" payment "SUCCESSFUL" validated
    And another "<failed-validated-count>" payment "FAILED" validated
    When "<signed-in user>" logins to backoffice app
    And admin is at payment validation screen
    And selects the existing bulk payment validation record to download
    Then receives "OK" status code
    And has response payment data with "<total-validated-count>" correct records
    And has response validation date
    #test maximum 2000 records that can be included in Direct Debit scenario locally as can be flaky on kafka sync ci
    Examples:
      | signed-in user | success-validated-count |  failed-validated-count | payment-method    | total-validated-count |
      | school admin   | 10                      |  10                     | DIRECT DEBIT      | 20                    |
      | hq staff       | 1                       |  1                      | CONVENIENCE STORE | 2                     |
      | school admin   | 1                       |  0                      | DIRECT DEBIT      | 1                     |
      | school admin   | 0                       |  1                      | DIRECT DEBIT      | 1                     |

  Scenario Outline: The payment file has no associated payment validation details
    Given there is existing bulk payment validation record for "<payment-method>"
    When "<signed-in user>" logins to backoffice app
    And admin is at payment validation screen
    And selects the existing bulk payment validation record to download
    Then receives "Internal" status code

    Examples:
      | signed-in user | payment-method    |
      | school admin   | DIRECT DEBIT      |
      | hq staff       | CONVENIENCE STORE |