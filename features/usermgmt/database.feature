Feature: Database

  Scenario: Check shard id value in session variable
    Given a database with shard id
    When client acquires a connection
    Then the connection should have corresponding shard id in session variable

  Scenario: Generate valid sharded id
    When client generate sharded id via database func
    Then the client receive valid sharded id