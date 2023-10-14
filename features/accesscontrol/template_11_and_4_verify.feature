@runsequence
Feature: Verify policies gen by template 1.1 and 4

  Background:
    Given create permission with name "accesscontrol.ac_test_template_11_4.read"
    And create permission with name "accesscontrol.ac_test_template_11_4.write"
    And create role with name "role-read-write"
    And add permission "accesscontrol.ac_test_template_11_4.read" and "accesscontrol.ac_test_template_11_4.write" to permission role "role-read-write"
    And create user group name "user-group-read-write"
    And add role "read-write" and user group "user-group-read-write" to granted role "granted-role-read-write"
    And add user "user-read-write-permission" to user group "user-group-read-write"
    And create location "A" with access path "A" with parent ""
    And create location "A1" with access path "A,A1" with parent "A"
    And create location "A1.1" with access path "A,A1,A1.1" with parent "A1"
    And create location "A1.2" with access path "A,A1,A1.2" with parent "A1"
    And create location "A2" with access path "A,A2" with parent "A"
    And create location "B" with access path "B" with parent ""
    And create location "C" with access path "C" with parent ""
    And create location "D" with access path "D" with parent ""
    And create location "D1" with access path "D,D1" with parent "D"
    And create location "D2" with access path "D,D2" with parent "D"
    And Assign location "A1" to granted role "granted-role-read-write"
    And add user "user-has-permission" to user group "user-group-read-write"

  Scenario Outline: user can get their own records even don't have permission and location
    Given login with user "user-has-owner"
    And add data "record has owner" with owner is "user-has-owner" to table ac_test_template_11_4
    When user get data "record has owner" from table ac_test_template_11_4
    Then command should return the records user is owners "record has owner"

  Scenario Outline: user can get records if they have permission and assigned location but don't own's records
    Given login with user "user-has-permission"
    And add data "record has permission location" with owner is "no-owner" with location "A1.1" to table ac_test_template_11_4
    When user get data "record has permission location" from table ac_test_template_11_4
    Then command should return the records user is owners "record has permission location"

  Scenario Outline: user can get records if they have permission and assigned location and also own's records
    Given login with user "user-owner-permission"
    And add data "owner and permission" with owner is "user-owner-permission" with location "A1.1" to table ac_test_template_11_4
    When user get data "owner and permission" from table ac_test_template_11_4
    Then command should return the records user is owners "owner and permission"

  Scenario Outline: user can't get any records when don't have permission location or owners
    Given login with user "user-no-permission-owner"
    And add data "no owner and permission" with owner is "no-owner-user" to table ac_test_template_11_4
    When user get data "no owner and permission" from table ac_test_template_11_4
    Then command should return the records user is owners ""

  Scenario Outline: user can insert any data into table even don't have permission location and owners
    Given login with user "<user>"
    When add data "<user>" with owner is "no-owner-insert" to table ac_test_template_11_4
    Then command return successfully
    Examples:
      | user                       | data |
      | user-no-permission-onwer-1 | 11   |
      | user-no-permission-onwer-2 | 22   |
      | user-no-permission-onwer-3 | 33   |
      | user-no-permission-onwer-4 | 44   |
      | user-no-permission-onwer-5 | 55   |
      | user-no-permission-onwer-6 | 66   |


  Scenario Outline: user can update their own records even don't have permission and location
    Given login with user "user-has-owner"
    And add data "record has owner update" with owner is "user-has-owner" to table ac_test_template_11_4
    When user update data "record has owner update" with name "updated" to table ac_test_template_11_4
    Then command return successfully

  Scenario Outline: user can update their records they granted permission without owners
    Given login with user "user-has-permission"
    And add data "record has permission update" with owner is "no-owner" with location "A1.1" to table ac_test_template_11_4
    When user update data "record has permission update" with name "updated" to table ac_test_template_11_4
    Then command return successfully

  Scenario Outline: user can update their records they granted permission without owners
    Given login with user "user-owner-permission"
    And add data "record has owner and permission update" with owner is "user-owner-permission" with location "A1.1" to table ac_test_template_11_4
    When user update data "record has owner and permission update" with name "updated" to table ac_test_template_11_4
    Then command return successfully

  Scenario Outline: user can update their records they granted permission without owners
    Given login with user "user-no-permission-owner"
    And add data "no owner and permission update" with owner is "user-owner-permission" with location "A1.1" to table ac_test_template_11_4
    When user update data "no owner and permission update" with name "updated" to table ac_test_template_11_4
    Then command return fail

  Scenario Outline: user can delete their own records even don't have permission and location
    Given login with user "user-has-owner"
    And add data "record has owner delete" with owner is "user-has-owner" to table ac_test_template_11_4
    When user delete data "record has owner delete" from table ac_test_template_11_4
    Then command return successfully

  Scenario Outline: user can delete their records they granted permission without owners
    Given login with user "user-has-permission"
    And add data "record has permission delete" with owner is "no-owner" with location "A1.1" to table ac_test_template_11_4
    When user delete data "record has permission delete" from table ac_test_template_11_4
    Then command return successfully

  Scenario Outline: user can delete their records they granted permission without owners
    Given login with user "user-owner-permission"
    And add data "record has owner and permission delete" with owner is "user-owner-permission" with location "A1.1" to table ac_test_template_11_4
    When user delete data "record has owner and permission delete" from table ac_test_template_11_4
    Then command return successfully

  Scenario Outline: user can delete their records they granted permission without owners
    Given login with user "user-no-permission-owner"
    And add data "no owner and permission delete" with owner is "user-owner-permission" with location "A1.1" to table ac_test_template_11_4
    When user delete data "no owner and permission delete" from table ac_test_template_11_4
    Then command return fail