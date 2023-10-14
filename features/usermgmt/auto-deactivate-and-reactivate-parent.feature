@blocker
Feature: Auto deactivate and reactivate parent

  Background: Sign in with role "staff granted role school admin"
    Given a signed in as "staff granted role school admin" in "manabie" organization

  Scenario Outline: Auto deactivate and reactivate parent when creating parent
    Given school admin creates 1 students with "<condition>" by OpenAPI in folder "enrollment_status_histories"
    When school admin creates a parent with 1 student(s) by OpenAPI
    Then school admin sees parent "<activation>"

    Examples:
      | condition                   | activation  |
      | enrollment status potential | activated   |
      | enrollment status enrolled  | activated   |
      | enrollment status withdraw  | deactivated |

  Scenario Outline: Auto deactivate and reactivate parent when updating parent
    Given school admin creates 1 students with "<first student enrollment status>" by OpenAPI in folder "enrollment_status_histories"
    And school admin creates a parent with 1 student(s) by OpenAPI
    When school admin adds more "<second student enrollment status>" student to the parent by OpenAPI
    Then school admin sees parent "<parent-activation>"

    Examples:
      | first student enrollment status | second student enrollment status | activation  |
      | enrollment status potential     | enrollment status withdraw       | activated   |
      | enrollment status non-potential | enrollment status potential      | activated   |
      | enrollment status withdraw      | enrollment status withdraw       | deactivated |
      | enrollment status withdraw      | enrollment status potential      | activated   |

  Scenario Outline: Auto deactivate and reactivate parent when updating parent
    Given school admin creates 1 students with "<first student enrollment status>" by OpenAPI in folder "enrollment_status_histories"
    And school admin creates 1 students with "<second student enrollment status>" by OpenAPI in folder "enrollment_status_histories"
    And school admin creates a parent with 2 student(s) by OpenAPI
    When school admin removes 1 student from the parent by OpenAPI
    Then school admin sees parent "<parent-activation>"

    Examples:
      | first student enrollment status | second student enrollment status | activation  |
      | enrollment status potential     | enrollment status withdraw       | activated   |
      | enrollment status non-potential | enrollment status potential      | activated   |
      | enrollment status withdraw      | enrollment status withdraw       | deactivated |
      | enrollment status withdraw      | enrollment status potential      | deactivated |
