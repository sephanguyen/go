Feature: Upload enrollment pdf

  Scenario Outline: Upload enrollment pdf success
    Given prepare enrollment pdf for upload
    When "school admin" upload file
    Then receives "OK" status code

  Scenario Outline: Get url-down enrollment pdf success
    Given prepare enrollment pdf for upload
    And "school admin" upload file
    And receives "OK" status code
    When "school admin" get download url enrollment file
    Then receives "OK" status code
