@runsequence
Feature: Verify policies gen by template 4 with hasura
  Background:
    Given mastermgmt hasura

  Scenario Outline: user only get record which is own
    Given hasura table ac_test_template_4 with record "HASURA_A_OWNER" and owner "user_hasura_1"
    And login hasura with user "<user>"
    When user get data "HASURA_A_OWNER" from hasura table ac_test_template_4
    Then user should only get "<records>" their with signed
    Examples:
      | user          | records |
      | user_hasura_1 | 1       |
      | user_hasura_2 | 0       |
      | user_hasura_7 | 0       |
      | user_hasura_4 | 0       |
      | user_hasura_5 | 0       |
      | user_hasura_6 | 0       |

  Scenario Outline: user can insert any data into hasura table ac_test_template_4
    Given login hasura with user "<user>"
    When user insert data "<data>" into hasura table ac_test_template_4
    Then command return successfully
    Examples:
      | user           | data      |
      | user_hasura_11 | hasura_11 |
      | user_hasura_22 | hasura_22 |
      | user_hasura_33 | hasura_33 |
      | user_hasura_44 | hasura_44 |
      | user_hasura_55 | hasura_55 |
      | user_hasura_66 | hasura_66 |

  Scenario Outline: user can not insert into hasura table ac_test_template_4 if owners column is not set their user_id
    Given login hasura with user "user_hasura_insert_fail"
    When user insert data "HASURA_INSERT" with owners "user_hasura_1" into hasura ac_test_template_4
    Then command return "0" row affected

  Scenario Outline: user can update only record their insert
    Given login hasura with user "user_hasura_3"
    And user insert data "HASURA_A1" into hasura table ac_test_template_4
    When user "update" data "HASURA_A1" with name "HASURA_B" into hasura ac_test_template_4
    Then command return "1" row affected

  Scenario Outline: user can not update record their own
    Given login hasura with user "user_hasura_3"
    And user insert data "HASURA_A2" into hasura table ac_test_template_4
    And login hasura with user "user_hasura_2"
    When user "update" data "HASURA_A2" with name "HASURA_B" into hasura ac_test_template_4
    Then command return "0" row affected

  Scenario Outline: user can delete only record their insert
    Given login hasura with user "user_hasura_3"
    And user insert data "HASURA_A3" into hasura table ac_test_template_4
    When user "delete" data "HASURA_A3" with name "HASURA_B" into hasura ac_test_template_4
    Then command return "1" row affected

  Scenario Outline: user can not delete record their own
    Given login hasura with user "user_hasura_3"
    And user insert data "HASURA_A4" into hasura table ac_test_template_4
    And login hasura with user "user_hasura_2"
    When user "delete" data "HASURA_A4" with name "HASURA_B" into hasura ac_test_template_4
    Then command return "0" row affected