@quarantined
Feature: Resume recurring products

  Scenario Outline: Resume billing success
    Given prepare data for resume order when student status is LOA
    When "school admin" submit order
    Then paused products are resumed successfully
    And receives "OK" status code