Feature: Run migration job to migrate update student enrollment original status to new status in our system

  Scenario: Migrate update student enrollment original status to new status in our system
    Given enrollment status outdate in our system
    When system run job to disable access path location for outdate enrollment status
    Then student no longer access location when access path removed
