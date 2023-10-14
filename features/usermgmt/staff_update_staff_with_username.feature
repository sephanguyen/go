Feature: Update staffs with username
  Background:
    Given a signed in "staff granted role school admin"
    And existed staff with "user group was granted teacher role"

  Scenario Outline: update staffs with "<generate staff profile>" successfully
    Given a profile of staff with name: "update-staff-%s"; user group type: "user group was granted teacher role" and generate a "<generate staff profile>" UpdateStaffProfile and "add more valid locations" locations
    When staff update profile
    Then profile of staff must be updated

    Examples:
      | generate staff profile               |
      | available username                   |
      | available username with email format |

  Scenario Outline: update staffs with "<generate staff profile>" failed
    Given a profile of staff with name: "update-staff-%s"; user group type: "user group was granted teacher role" and generate a "<generate staff profile>" UpdateStaffProfile and "add more valid locations" locations
    When staff update profile
    Then returns "<msg>" status code

    Examples:
      | generate staff profile           | msg             |
      | empty username                   | InvalidArgument |
      | username has spaces              | InvalidArgument |
      | username has special characters  | InvalidArgument |
      | existing username                | AlreadyExists   |
      | existing username and upper case | AlreadyExists   |
