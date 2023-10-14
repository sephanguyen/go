Feature: Update parents with username
  As a school staff
  I'm able to update parents and family relationship and their username

  Background:
    Given a signed in "staff granted role school admin"
    And parents data to update

  Scenario: Update parent with available username
    Given update parent data "available username"
    When "staff granted role school admin" update new parents
    Then parents were updated successfully

  Scenario Outline: Cannot update if parent data to update has <field>
    Given parent data to update has empty or invalid "<field>"
    When "staff granted role school admin" update new parents
    Then "staff granted role school admin" cannot update parents
    And receives "InvalidArgument" status code

    Examples:
      | field                            |
      | empty username                   |
      | username has spaces              |
      | username has special characters  |
      | existing username                |
      | existing username and upper case |
