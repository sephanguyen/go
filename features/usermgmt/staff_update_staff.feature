@blocker
Feature: Update User Profile

  Scenario Outline: User update profile of another staff profile
    Given a signed in "<role>"
    And existed staff with "<init staff_profile type>"
    And a profile of staff with name: "<name>"; user group type: "<user_group_ids type>" and generate a "<generate staff profile>" UpdateStaffProfile and "<locations kind>" locations
    When staff update profile
    Then profile of staff must be updated

    Examples:
      | role                            | name            | init staff_profile type                  | user_group_ids type                      | generate staff profile                         | locations kind           |
      | staff granted role school admin | update-staff-%s | user group was granted teacher role      | user group was granted school admin role | email-non-existed                              | add more valid locations |
      | staff granted role school admin | update-staff-%s | user group was granted school admin role | user group was granted teacher role      | email-non-existed                              | add more valid locations |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | user group was granted school admin role | email-non-existed                              | add more valid locations |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | user group was granted teacher role      | empty-optional-field                           | add more valid locations |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | user group was granted school admin role | empty-optional-field-only-primary-phone-number | add more valid locations |
      | staff granted role school admin | update-staff-%s | user group was granted school admin role | user group was granted teacher role      | add more tags                                  | add more valid locations |
      | staff granted role school admin | update-staff-%s | user group was granted school admin role | user group was granted teacher role      | remove all tags                                | add more valid locations |
      | staff granted role school admin | update-staff-%s | empty external_user_id                   | user group was granted teacher role      | external_user_id-non-existed                   | add more valid locations |
      | staff granted role school admin | update-staff-%s | empty external_user_id                   | user group was granted teacher role      | external_user_id-non-existed and space         | add more valid locations |

  Scenario Outline: User cannot update profile of another without proper permission
    Given a signed in "<role>"
    And existed staff with "<init staff_profile type>"
    And a profile of staff with name: "<name>"; user group type: "<user_group_ids type>" and generate a "<generate staff profile>" UpdateStaffProfile and "<locations kind>" locations
    When staff update profile
    Then returns "<error>" status code

    Examples:
      | role            | name            | init staff_profile type                  | user_group_ids type                 | generate staff profile | locations kind           | error            |
      | unauthenticated | update-staff-%s | user group was granted teacher role      | user group was granted teacher role | email-non-existed      | add more valid locations | Unauthenticated  |
      | unauthenticated | update-staff-%s | user group was granted school admin role | user group was granted teacher role | empty-optional-field   | add more valid locations | Unauthenticated  |
      | student         | update-staff-%s | empty user_group_ids                     | user group was granted teacher role | email-non-existed      | add more valid locations | PermissionDenied |
      | teacher         | update-staff-%s | empty user_group_ids                     | user group was granted teacher role | email-non-existed      | add more valid locations | PermissionDenied |
      | parent          | update-staff-%s | empty user_group_ids                     | user group was granted teacher role | email-non-existed      | add more valid locations | PermissionDenied |

  Scenario Outline: User with missing required profile info cannot update profile
    Given a signed in "<role>"
    And existed staff with "<init staff_profile type>"
    And a profile of staff with name: "<name>"; user group type: "<user_group_ids type>" and generate a "<generate staff profile>" UpdateStaffProfile and "<locations kind>" locations
    When staff update profile
    Then returns "<error>" status code

    Examples:
      | role                            | name            | init staff_profile type                  | user_group_ids type                 | generate staff profile                 | locations kind           | error           |
      | staff granted role school admin | update-staff-%s | user group was granted teacher role      | user group was granted teacher role | email-existed                          | add more valid locations | AlreadyExists   |
      | staff granted role school admin | update-staff-%s | user group was granted school admin role | empty user_group_ids                | email-empty                            | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | user group was granted teacher role | duplicated-phone-number                | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | user group was granted teacher role | phone-number-has-wrong-type            | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | user group was granted teacher role | phone-numbers-have-same-type           | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | invalid user_group_ids              | email-non-existed                      | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | empty user_group_ids                     | user group was granted teacher role | start-date-is-less-than-end-date       | add more valid locations | InvalidArgument |
      | staff granted role school admin |                 | user group was granted teacher role      | user group was granted teacher role | empty-name-and-empty-first&last-name   | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | user group was granted teacher role      | user group was granted teacher role | email-non-existed                      | invalid locations        | InvalidArgument |
      | staff granted role school admin | update-staff-%s | user group was granted teacher role      | user group was granted teacher role | non existing tag                       | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | user group was granted teacher role      | user group was granted teacher role | wrong tag type                         | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | empty external_user_id                   | user group was granted teacher role | external_user_id-existed               | add more valid locations | AlreadyExists   |
      | staff granted role school admin | update-staff-%s | external_user_id-non-existed             | user group was granted teacher role | external_user_id-existed               | add more valid locations | AlreadyExists   |
      | staff granted role school admin | update-staff-%s | external_user_id-non-existed             | user group was granted teacher role | external_user_id-existed and space     | add more valid locations | AlreadyExists   |
      | staff granted role school admin | update-staff-%s | external_user_id-non-existed             | user group was granted teacher role | external_user_id-non-existed           | add more valid locations | InvalidArgument |
      | staff granted role school admin | update-staff-%s | external_user_id-non-existed             | user group was granted teacher role | external_user_id-non-existed and space | add more valid locations | InvalidArgument |

  Scenario Outline: staff cannot update another staff without permission to assign user_group
    Given a signed in "<role>"
    And existed staff with "<init staff_profile type>"
    And a profile of staff with name: "<name>"; user group type: "<user_group_ids type>" and generate a "<generate staff profile>" UpdateStaffProfile and "<locations kind>" locations
    When staff update profile
    Then returns "<error>" status code

    Examples:
      | role                        | name       | init staff_profile type | user_group_ids type                      | generate staff profile | locations kind           | error           |
      | staff granted role hq staff | updated-%s | empty user_group_ids    | user group was granted school admin role | email-non-existed      | add more valid locations | InvalidArgument |
