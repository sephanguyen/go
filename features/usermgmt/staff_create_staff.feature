@blocker
Feature: Create staffs

  Scenario Outline: user try to create account: "<generate staff profile>"
    Given a signed in "<role>"
    And generate a "<generate staff profile>" CreateStaffProfile and choose "<locations kind>" locations
    When "<role>" create staff account
    Then new staff account was created successfully

    Examples:
      | role                            | generate staff profile                          | locations kind  |
      | staff granted role school admin | user group was granted school admin role        | valid locations |
      | staff granted role school admin | user group was granted teacher role             | valid locations |
      | staff granted role school admin | empty user_group_ids                            | valid locations |
      | staff granted role school admin | full field valid                                | valid locations |
      | staff granted role school admin | empty optional field                            | valid locations |
      | staff granted role school admin | full field valid with only primary phone number | valid locations |
      | staff granted role school admin | empty name and valid first and last name        | valid locations |
      | staff granted role school admin | valid name and empty first and last name        | valid locations |
      | staff granted role school admin | non existed external user id                    | valid locations |
      | staff granted role school admin | non existed external user id with space         | valid locations |

  Scenario Outline: user try to create staff with wrong cases: "<generate staff profile>"
    Given a signed in "<role>"
    And generate a "<generate staff profile>" CreateStaffProfile and choose "<locations kind>" locations
    When "<role>" create staff account
    Then returns "<msg>" status code

    Examples:
      | role                            | generate staff profile                   | locations kind    | msg              |
      | unauthenticated                 | user group was granted school admin role | valid locations   | Unauthenticated  |
      | teacher                         | user group was granted teacher role      | valid locations   | PermissionDenied |
      | student                         | user group was granted school admin role | valid locations   | PermissionDenied |
      | parent                          | user group was granted teacher role      | valid locations   | PermissionDenied |
      # | organization manager              | user group was granted school admin role | valid locations   | PermissionDenied |
      | staff granted role school admin | user group was granted school admin role | invalid locations | InvalidArgument  |
      | staff granted role school admin | empty name and empty first and last name | invalid locations | InvalidArgument  |
      | staff granted role school admin | empty email                              | invalid locations | InvalidArgument  |
      | staff granted role school admin | duplicated email                         | valid locations   | AlreadyExists    |
      | staff granted role school admin | empty country                            | valid locations   | InvalidArgument  |
      | staff granted role school admin | wrong type phone number                  | valid locations   | InvalidArgument  |
      | staff granted role school admin | add two phone number have a same         | valid locations   | InvalidArgument  |
      | staff granted role school admin | invalid user_group_ids                   | valid locations   | InvalidArgument  |
      | staff granted role hq staff     | user group was granted school admin role | valid locations   | InvalidArgument  |
      | staff granted role hq staff     | non existing tag                         | valid locations   | InvalidArgument  |
      | staff granted role hq staff     | wrong tag type                           | valid locations   | InvalidArgument  |
      | staff granted role school admin | existed external user id                 | valid locations   | AlreadyExists    |