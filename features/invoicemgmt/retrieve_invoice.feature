@quarantined
# Deprioritized 
# If continued, new permission_role record required to associate payment.invoice.read to Parent role
Feature: Retrieve invoice details
  As a parent
  I am able to view invoice details
  Background:
    Given "parent" logins Learner App

  Scenario Outline: Parent retrieves invoice details successfully
    Given there is an existing invoice
    And invoice has "<invoice-status>" status with "<bill-items-count>" bill items count
    When logged-in user views an invoice
    Then receives "OK" status code
    And receives "<bill-items-count>" bill items count

    Examples:
      | invoice-status  | bill-items-count |
      | ISSUED          | 3                |
      | REFUNDED        | 4                |
      | ERROR           | 5                |
      | PAID            | 6                |
      | ISSUED          | 0                |

  Scenario Outline: Parent failed to retrieve invoice details of invoice in draft status
    Given there is an existing invoice
    And invoice has "DRAFT" status with "2" bill items count
    When logged-in user views an invoice
    Then receives "InvalidArgument" status code

  Scenario Outline: Parent failed to retrieve invoice details for non-existing invoice
    Given invoice ID is non-existing
    When logged-in user views an invoice
    Then receives "Internal" status code

  Scenario Outline: Unauthorized users failed to retrieve invoice details of invoice
    Given "<unauthorized-user>" logins Learner App
    And logged-in user views an invoice
    Then receives "<status-code>" status code

    Examples:
      | invoice-status  | bill-items-count | unauthorized-user | status-code      |
      | ERROR           | 5                | school admin      | PermissionDenied |
      | ISSUED          | 3                | teacher           | PermissionDenied |
      | ISSUED          | 1                | unauthenticated   | Unauthenticated  |