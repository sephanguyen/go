
Feature: Create staffs with username
  Background:
    Given a signed in "staff granted role school admin"

  Scenario Outline: create staffs with "<generate staff profile>" successfully
    Given generate a "<generate staff profile>" CreateStaffProfile and choose "valid locations" locations
    When "staff granted role school admin" create staff account
    Then new staff account was created successfully

    Examples:
      | generate staff profile               |
      | available username                   |
      | available username with email format |

  Scenario Outline: create staffs with "<generate staff profile>" failed
    Given generate a "<generate staff profile>" CreateStaffProfile and choose "valid locations" locations
    When "staff granted role school admin" create staff account
    Then returns "<msg>" status code

    Examples:
      | generate staff profile           | msg             |
      | empty username                   | InvalidArgument |
      | username has spaces              | InvalidArgument |
      | username has special characters  | InvalidArgument |
      | existing username                | AlreadyExists   |
      | existing username and upper case | AlreadyExists   |