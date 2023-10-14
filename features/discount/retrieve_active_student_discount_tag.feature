@blocker
Feature: retrieve active user discount tag

  Background:
    Given there is an existing discount master data with discount tag

  Scenario Outline: Retrieve Active User Discount Tag successfully
    Given there is a student that has "<record-count>" user discount tag "<discount-types>" records
    And this user discount tag has "<start-date>" start date and "<end-date>" end date
    And a valid request payload for retrieve user discount tag with date today
    When "<signed-in user>" retrieves user discount tag for this student
    Then receives "OK" status code
    And user discount tag "<discount-types>" records are retrieved successfully

    # TODAY minus and plus is equivalent to past and future dates to compare on current day when the test is run
    # End date is for combo and sibling discount type only
    Examples:
      | signed-in user | record-count | discount-types                                                           | start-date | end-date |
      | school admin   | 1            | combo                                                                    | TODAY      | TODAY+1  |
      | hq staff       | 2            | sibling/family                                                           | TODAY-1    | TODAY+2  |
      | centre staff   | 3            | combo/sibling/employee full time                                         | TODAY-1    | TODAY+3  |
      | centre manager | 4            | employee full time/employee part time/single parent/family               | TODAY      | TODAY+1  |
      | centre lead    | 5            | combo/family/employee full time/employee part time/single parent         | TODAY-2    | TODAY+2  |
      | school admin   | 6            | combo/sibling/employee full time/employee part time/single parent/family | TODAY-3    | TODAY+3  |

  Scenario Outline: Retrieve Active User Discount Tag with no active start and end date
    Given there is a student that has "<record-count>" user discount tag "<discount-types>" records
    And this user discount tag has "<start-date>" start date and "<end-date>" end date
    And a valid request payload for retrieve user discount tag with date today
    When "<signed-in user>" retrieves user discount tag for this student
    Then receives "OK" status code
    And there is no user discount tag records retrieved
    # for org level discount the start date will be created date
    Examples:
      | signed-in user | record-count | discount-types                                                           | start-date | end-date |
      | school admin   | 2            | combo/sibling                                                            | TODAY-9    | TODAY-1  |
      | hq staff       | 2            | family/employee full time                                                | TODAY+1    | NONE     |
      | hq staff       | 1            | single parent                                                            | TODAY+1    | NONE     |
      | school admin   | 1            | sibling                                                                  | TODAY+2    | TODAY+9  |
      | centre manager | 1            | sibling                                                                  | TODAY+2    | TODAY+9  |
      | centre lead    | 1            | employee full time/employee part time/single parent/family               | TODAY+1    | NONE     |

  Scenario Outline: Retrieve Active User Discount Tag with invalid student
    Given there is an invalid "<condition>" payload request for retrieve user discount tag
    When "school admin" retrieves user discount tag for this student
    Then receives "FailedPrecondition" status code

    Examples:
      | condition                           |
      | empty student                       |
      | empty date request                  |
      | empty both student and date request |




