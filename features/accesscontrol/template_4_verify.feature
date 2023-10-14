@runsequence
Feature: Verify policies gen by template 4
  Scenario Outline: user only get record which is own
    Given table ac_test_template_4 with record "A" and owner "user1"
    And login with user "<user>"
    When user get data from table ac_test_template_4
    Then user should only get "<records>" their with signed
    Examples:
      | user  | records |
      | user1 | 1       |
      | user2 | 0       |
      | user7 | 0       |
      | user4 | 0       |
      | user5 | 0       |
      | user6 | 0       |

  Scenario Outline: user can insert any data into table ac_test_template_4
    Given login with user "<user>"
    When user insert data "<data>" into table ac_test_template_4
    Then command return successfully
    Examples:
      | user   | data |
      | user11 | 11   |
      | user22 | 22   |
      | user33 | 33   |
      | user44 | 44   |
      | user55 | 55   |
      | user66 | 66   |

  Scenario Outline: user can not insert into table ac_test_template_4 if owners column is not set their user_id
    Given login with user "user-insert-fail"
    When user insert data "A1" with owners "user-1"
    Then command return fail

  Scenario Outline: user can update only record their insert
    Given login with user "user3"
    And user insert data "A1" into table ac_test_template_4
    When user "update" data "A1" with name "B"
    Then command return successfully

  Scenario Outline: user can not update record their own
    Given login with user "user3"
    And user insert data "A2" into table ac_test_template_4
    And login with user "user2"
    When user "update" data "A2" with name "B"
    Then command return fail

  Scenario Outline: user can delete only record their insert
    Given login with user "user3"
    And user insert data "A3" into table ac_test_template_4
    When user "delete" data "A3" with name "B"
    Then command return successfully

  Scenario Outline: user can not delete record their own
    Given login with user "user3"
    And user insert data "A4" into table ac_test_template_4
    And login with user "user2"
    When user "delete" data "A4" with name "B"
    Then command return fail

