Feature: Upsert kafka connect
  kafka connect help us in sync data pipeline
  service hephaestus will update these kafka connector

  Scenario: Job upsert kafka connect connector successfully
    Given table "ta" in database bob and fatima
    And source connector file for table "ta" in bob
    And sink connector file for table "ta" in fatima
    When run job upsert kafka connector
    And insert several records to table "ta"
    Then records is synced in source and sink table "ta"
