Feature: Create student with grade master
  As a school staff
  I need to be able to create a new student with grade master

  Scenario Outline: Create a student with grade master
    Given generate grade master
    And student info with grade master request
    When "staff granted role school admin" create new student account
    Then new student account created success with grade master
    And receives "OK" status code

  Scenario Outline: Validate grade master
    Given student info with invalid grade master request
    When "staff granted role school admin" create new student account
    Then "staff granted role school admin" cannot create that account
    And receives "InvalidArgument" status code
