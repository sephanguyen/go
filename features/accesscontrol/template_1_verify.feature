@runsequence
Feature: Verify policies gen by template 1
  Background:
    Given table ac_test_template_1 with location "A1" and permission "accesscontrol.b.read"
    And login as userId "user1" and group "1"
    And user assigned "A1" and "accesscontrol.b.write"
    And user "insert" data "new data" belong to location "A1" into table ac_test_template_1

  Scenario Outline: user get only records they assigned from table ac_test_template_1
    Given login as userId "<user>" and group "<user_group>"
    And user assigned "<location>" and "<permission>"
    When user get data from table ac_test_template_1
    Then user should only get "<records>" their assigned
    Examples:
      | user  | user_group | location | permission           | records |
      | user1 | 1          | A1       | accesscontrol.b.read | 1       |
      | user2 | 2          | A1       | accesscontrol.c.read | 0       |
      | user3 | 3          | A1       | accesscontrol.e.read | 0       |
      | user4 | 4          | A4       | accesscontrol.b.read | 0       |
      | user5 | 5          | A5       | accesscontrol.b.read | 0       |
      | user6 | 6          | A6       | accesscontrol.b.read | 0       |

  Scenario Outline: assigned location user can update record by their location
    Given login as userId "user1" and group "1"
    And user assigned "A1" and "accesscontrol.b.write"
    And user "insert" data "new data update" belong to location "A1" into table ac_test_template_1
    When user "update" data "new data update" belong to location "A1" into table ac_test_template_1
    Then return successfully

  Scenario Outline: user can not update record if not yet assigned location
    Given login as userId "user2" and group "2"
    When user "update" data "new data" into table ac_test_template_1
    Then return fail

  Scenario Outline: assigned location user can insert record by their location
    Given login as userId "user1" and group "1"
    And user assigned "A1" and "accesscontrol.b.write"
    When user "insert" data "new data insert" belong to location "A1" into table ac_test_template_1
    Then return successfully

  Scenario Outline: user can not insert record if not yet assigned location
    Given login as userId "user2" and group "2"
    When user "insert" data "new data insert" into table ac_test_template_1
    Then return fail

  Scenario Outline: assigned location user can delete record by their location
    Given login as userId "user1" and group "1"
    And user assigned "A1" and "accesscontrol.b.write"
    And user assigned "A1" and "accesscontrol.b.read"
    And user "insert" data "new data delete" belong to location "A1" into table ac_test_template_1
    When user "delete" data "new data delete" belong to location "A1" into table ac_test_template_1
    Then return successfully

  Scenario Outline: user can not delete record if not yet assigned location
    Given login as userId "user2" and group "2"
    When user "delete" data "new data" into table ac_test_template_1
    Then return fail

