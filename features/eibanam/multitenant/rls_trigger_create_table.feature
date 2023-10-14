Feature: Auto create column resource_path when create table, add rls policy, force rls and enable rls

    @wip
    Scenario: Create new table
    When create some table with random name
    Then those tables must have column resource_path
    And those tables must have rls enabled and rls forced