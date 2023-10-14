Feature: Retrieve students associated to parent account

  Scenario: Father and Mother have same children in manabie
    Given a "newly created" signed in user with "<role>"
    When create handsome father and pretty mother as parent and the relationship with his children who're students at manabie
    And retrieve students profiles associated to each account
    Then returns "OK" status code
    And fetched students exactly associated to parent
    And returns the same students profiles
  
  Examples:
    |role          |
    | school admin |

  Scenario: Father and Mother have has no children in manabie
    Given a "newly created" signed in user with "<role>"
    When create handsome father and pretty mother as parent and the relationship with his children who're students at manabie
    And remove relationship of student
    And retrieve students profiles associated to each account
    Then returns "OK" status code
    And no students profiles are fetched

  Examples:
    | role         |
    | school admin |