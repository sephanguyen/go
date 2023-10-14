Feature: Incremental snapshot new table
    debezium will using this to snapshot new added table

  Scenario: Job upsert kafka connect connector successfully and send incremental snapshto for new table 
# Create table "tc" in bob with several records and create table in fatima as well. 
# When run job upsert_kafka_connector in hephaestus, 
# It will sync data in table tc from bob to fatima (we check that in fatima is synced with correct data in bob)

# Create another table "td" in bob and fatima with several records in source bob
# Sync this new table (by updating field table.include.list) in source connector config
# Run job upsert_kafka_connector again
# It will sync data in table td from bob to fatima
    Given table "tc" in database bob and fatima
    # And insert several records to table "tc"

    And source connector file for table "tc" in bob
    And sink connector file for table "tc" in fatima

    When run job upsert kafka connector
    # Then records is synced in source and sink table "tc"

    When insert several records to table "tc"
    Then records is synced in source and sink table "tc"

    Given table "td" in database bob and fatima
    And insert several records to table "td"

    And add table "td" to captured table list in source connector
    And sink connector file for table "td" in fatima

    When run job upsert kafka connector
    Then records is synced in source and sink table "td"
