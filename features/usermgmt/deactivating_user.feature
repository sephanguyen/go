@blocker
Feature: Deactivate and Re-activate Users
  As a school admin staff
  I need to be able to deactivate and re-activate user
  And user should not be able to login to the system when deactivated
  And user should be able to login to the system when re-activated

  Scenario Outline: Deactivated <role> cannot login
    Given a signed in "staff granted role school admin"
    And a "<role>" has been created successfully
    Then this "<role>" "<status-before-deactivating>" login to the system
    When staff "deactivates" this user
    Then this "<role>" "<status-after-deactivating>" login to the system

    Examples:
      | role    | status-before-deactivating | status-after-deactivating |
      | Student | can                        | cannot                    |
      | Parent  | can                        | cannot                    |
      | Staff   | can                        | cannot                    |


  Scenario Outline: Re-activated <role> can login
    Given a signed in "staff granted role school admin"
    And a "<role>" has been created successfully
    Then staff "deactivates" this user
    And this "<role>" "<status-before-reactivating>" login to the system
    When staff "re-activates" this user
    Then this "<role>" "<status-after-reactivating>" login to the system

    Examples:
      | role    | status-before-reactivating | status-after-reactivating |
      | Student | cannot                     | can                       |
      | Parent  | cannot                     | can                       |
      | Staff   | cannot                     | can                       |


  Scenario Outline: Deactivated <role> cannot login
    Given a signed in "staff granted role school admin"
    And a "<role>" has been created successfully
    Then this "<role>" "<status-before-deactivating>" login to the system
    When staff "deactivates" this user
    Then this "<role>" uses the old credential and "<ability-before-reactivating>" get self profile
    When staff "re-activates" this user
    Then this "<role>" uses the old credential and "<ability-after-reactivating>" get self profile

    Examples:
      | role    | status-before-deactivating | ability-before-reactivating | ability-after-reactivating |
      | Student | can                        | cannot                      | can                        |
      | Parent  | can                        | cannot                      | can                        |
      | Staff   | can                        | cannot                      | can                        |


  Scenario Outline: Deactivated <role> cannot login
    Given a signed in "staff granted role school admin"
    When staff "<action>" "<type>" users
    Then receives "<code>" status code

    Examples:
      | action       | type         | code             |
      | deactivates  | none-existed | InvalidArgument  |
      | re-activates | deleted      | InvalidArgument  |
