Feature: Update student with grade master
  As a school staff
  I need to be able to update a new student with grade master

  Scenario Outline: Update a student with grade master
    Given generate grade master
    And student info with grade master request
    And "staff granted role school admin" create new student account
    And generate grade master
    And student info with grade master update request
    When "staff granted role school admin" update student account
    Then new student account updated success with grade master
    And receives "OK" status code

  Scenario Outline: Validate grade master
  Given generate grade master
    And student info with grade master request
    And "staff granted role school admin" create new student account
    And student info with invalid grade master update request
    When "staff granted role school admin" update student account
    Then "staff granted role school admin" cannot update student account
    And receives "InvalidArgument" status code
