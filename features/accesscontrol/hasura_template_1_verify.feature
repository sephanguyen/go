@runsequence
Feature: Verify policies gen by template 1
  Background:
    Given mastermgmt hasura
    And table ac_hasura_test_template_1 with location "HA1" and permission "accesscontrol.b.read"
    And login as userId "user1" and group "1"
    And user assigned "HA1" and "accesscontrol.b.write"
    And user insert data "new data hasura" belong to location "HA1" into table ac_hasura_test_template_1 in hasura

  Scenario Outline: user get only records they assigned from table ac_hasura_test_template_1
    Given login as userId "<user>" and group "<user_group>"
    And user assigned "<location>" and "<permission>"
    When user get data "new data hasura" from table ac_hasura_test_template_1 in hasura
    Then user should only get "<records>" their assigned
    Examples:
      | user  | user_group | location | permission           | records |
      | user1 | 1          | HA1      | accesscontrol.b.read | 1       |
      | user2 | 2          | HA1      | accesscontrol.c.read | 0       |
      | user3 | 3          | HA1      | accesscontrol.e.read | 0       |
      | user4 | 4          | A4       | accesscontrol.b.read | 0       |
      | user5 | 5          | A5       | accesscontrol.b.read | 0       |
      | user6 | 6          | A6       | accesscontrol.b.read | 0       |

  Scenario Outline: assigned location user can update record by their location
    Given login as userId "user1" and group "1"
    And user assigned "HA1" and "accesscontrol.b.write"
    And user "insert" data "new data hasura update" belong to location "HA1" into table ac_hasura_test_template_1 in hasura
    When user "update" data "new data hasura update" belong to location "HA1" into table ac_hasura_test_template_1 in hasura
    Then hasura return "successfully"

  Scenario Outline: user can not update record if not yet assigned location
    Given login as userId "user2" and group "2"
    When user "update" data "new data hasura" belong to location "HA-UP" into table ac_hasura_test_template_1 in hasura
    Then hasura return "fail"

  Scenario Outline: assigned location user can insert record by their location
    Given login as userId "user1" and group "1"
    And user assigned "HA1" and "accesscontrol.b.write"
    When user "insert" data "new data hasura insert" belong to location "HA1" into table ac_hasura_test_template_1 in hasura
    Then hasura return "successfully"

  Scenario Outline: user can not insert record if not yet assigned location
    Given login as userId "user2" and group "2"
    When user "insert" data "new data hasura not insert" belong to location "HA-IN" into table ac_hasura_test_template_1 in hasura
    Then hasura return "fail"

  Scenario Outline: assigned location user can delete record by their location
    Given login as userId "user1" and group "1"
    And user assigned "HA1" and "accesscontrol.b.write"
    And user assigned "HA1" and "accesscontrol.b.read"
    And user "insert" data "new data hasura delete" belong to location "HA1" into table ac_hasura_test_template_1 in hasura
    When user "delete" data "new data hasura delete" belong to location "HA1" into table ac_hasura_test_template_1 in hasura
    Then hasura return "successfully"

  Scenario Outline: user can not delete record if not yet assigned location
    Given login as userId "user2" and group "2"
    When user "delete" data "new data hasura" belong to location "HA1" into table ac_hasura_test_template_1 in hasura
    Then hasura return "fail"

