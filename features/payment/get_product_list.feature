Feature: Get Product List

  Scenario Outline: Get product list successful
    Given create products to get product list
    And "school admin" get product list after creating product with "<filter>"
    Then get product list successfully

    Examples:
      | filter                                |
      | without filter                        |
      | with empty filter                     |
      | with product type filter              |
