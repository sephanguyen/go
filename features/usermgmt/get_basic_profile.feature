@blocker
Feature: User get basic profile

  Scenario Outline: user get basic profile
    Given a signed in "<role>"
    When user get basic profile
    Then user receive basic profile

    Examples:
      | role                            |
      | staff granted role school admin |
      | staff granted role teacher      |
      | student                         |
      | parent                          |
    
  Scenario Outline: user get basic profile with invalid request
    Given a signed in "<role>"
    When user get basic profile with invalid "<type>" request
    Then user can not get basic profile

    Examples:
      | role                            | type            |
      | staff granted role school admin | invalid user id |
      | staff granted role school admin | missing token   |
      | staff granted role teacher      | invalid user id |
      | staff granted role teacher      | missing token   |
      | student                         | invalid user id |
      | student                         | missing token   |
      | parent                          | invalid user id |
      | parent                          | missing token   |