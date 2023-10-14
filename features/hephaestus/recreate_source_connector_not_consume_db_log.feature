Feature: When we delete and recreate source connector, it will not sync data in the off time

  Scenario: disable sync data and enable again will not sync data in the off time
    Given table "tb" in database bob and fatima
    And create debezium source connector for that table "tb" in bob
    And create sink connector for that table "tb" in fatima
    And delete debezium source connector
    And insert several records to table "tb"

    When create debezium source connector for that table "tb" in bob
    Then the data insert before will not be synced
    
    When insert several records to table "tb"
    Then the data will be synced

