@blocker
Feature: Update parents
  As a school staff
  I'm able to update parents and family relationship

  Background:
    Given a signed in "staff granted role school admin"
    And parents data to update

  Scenario Outline: update parents successfully by <role>
    When "staff granted role <role>" update new parents
    Then parents were updated successfully

    Examples:
      | role           |
      | school admin   |
      | hq staff       |
      | centre lead    |
      | centre manager |
      | centre staff   |

  Scenario Outline: Update parent with full fields
    Given update parent data "full fields"
    When "staff granted role school admin" update new parents
    Then parents were updated successfully

  Scenario Outline: Update parent to blank in non-required fields <field>
    Given edit update parent data to blank in "<field>"
    When "staff granted role school admin" update new parents
    Then parents were updated successfully

    Examples:
      | field                         |
      | last name phonetic            |
      | first name phonetic           |
      | parent primary phone number   |
      | parent secondary phone number |
  # | remark                        | TODO: update remark case when fixing the bug : https://manabie.atlassian.net/browse/LT-34956

  Scenario Outline: Cannot update if parent data to update has empty or invalid <field>
    Given parent data to update has empty or invalid "<field>"
    When "staff granted role school admin" update new parents
    Then "staff granted role school admin" cannot update parents
    And receives "<msg>" status code

    Examples:
      | field                          | msg             |
      | id                             | InvalidArgument |
      | email                          | InvalidArgument |
      | relationship                   | InvalidArgument |
      | parentPhoneNumber invalid      | InvalidArgument |
      | student_id not exist           | InvalidArgument |
      | parent_id not exist            | InvalidArgument |
      | tag not exist                  | InvalidArgument |
      | tag for only student           | InvalidArgument |
      | last name                      | InvalidArgument |
      | first name                     | InvalidArgument |
      | email already exist            | AlreadyExists   |
      | external_user_id already exist | AlreadyExists   |
      | external_user_id re-update     | InvalidArgument |

  Scenario Outline: <role> don't have permission to update parent
    When "<role>" update new parents
    Then "<role>" cannot update parents
    And receives "PermissionDenied" status code

    Examples:
      | role                       |
      | student                    |
      | staff granted role teacher |
      | parent                     |
