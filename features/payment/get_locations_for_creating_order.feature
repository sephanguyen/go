Feature: Get Locations For Creating Order

  Scenario Outline: Get locations for creating order Success
    Given new locations
    And new "<signed-in user>" is granted to locations
    When getting locations for creating order
    Then receives "OK" status code
    And  check response data with "<type of location>" locations

    Examples:
      | signed-in user        |  type of location |
      | centre manager        |       list        |
      | school admin          |       list        |
    # | centre lead           |       empty       |
    # Only role Centre Lead do not payment.order.write permission, but is restricted when call API (auth.go) => Commented this case
    # Can uncomment it and add role to Payment rbacDecider to test it in local


