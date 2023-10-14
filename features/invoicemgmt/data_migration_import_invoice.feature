@quarantined
Feature: Migrate Invoice from CSV data

  Scenario: HQ manager successfully migrate invoices from CSV data
    Given there are "<student_count>" students that have "<bill_item_count>" bill items migrated with "<total_prices>" total price
    And there is invoice CSV file for these students
    And "school admin" logins to backoffice app
    When admin imports invoice migration data
    Then receives "OK" status code
    And there are no error lines in import invoice response
    And there are "<student_count>" invoices of students migrated successfully
    And migrated invoices have correct amount based on its status
    And migrated invoice have saved reference number and migrated_at

    Examples:
      | student_count | bill_item_count  | total_prices                     |
      | 5             | 3                | 3000.00&216.30&-99.66&21.33&6.30 |
      | 3             | 1                | 1200&300&200                     |

  Scenario: HQ manager failed to upload with student not existing
    Given there is invoice CSV file for non existing students
    And "school admin" logins to backoffice app
    When admin imports invoice migration data
    Then receives "OK" status code
    And there are error lines in import invoice response

  Scenario: HQ manager failed to upload with invoice have invalid amount
    Given there are "2" students that have "1" bill items migrated with "1000&500" total price
    And there is invoice CSV file for these students with invalid amount
    And "school admin" logins to backoffice app
    When admin imports invoice migration data
    Then receives "OK" status code
    And there are error lines in import invoice response
