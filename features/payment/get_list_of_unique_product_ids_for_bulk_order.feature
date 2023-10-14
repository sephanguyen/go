@quarantined
Feature: Get Unique Product List For Bulk Order

  Scenario: Get unque product list success
    Given create data for bulk order for unique product
    And "school admin" create bulk orders data for get unique product successfully
    When "school admin" get unique product for bulk order 
    Then receives "OK" status code
    And list of unique products for bulk order were returned correctly