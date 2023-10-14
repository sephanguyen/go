@quarantined
Feature: Handle MasterMgmtClassEvent
    Scenario: Handle JoinMasterMgmtClass event
      Given a valid JoinMasterMgmtClass event
      When send "MasterMgmtClassEvent" topic "MasterMgmt.Class.Upserted" to nats js
      Then our system must update MasterMgmtClass data correctly

    Scenario: Handle LeaveMasterMgmtClass event
      Given a valid LeaveMasterMgmtClass event
      When send "MasterMgmtClassEvent" topic "MasterMgmt.Class.Upserted" to nats js
      Then our system must update MasterMgmtClass data correctly

    Scenario: Handle CreateCourseMasterMgmtClass event
      Given a valid CreateCourseMasterMgmtClass event
      When send "MasterMgmtClassEvent" topic "MasterMgmt.Class.Upserted" to nats js
      Then our system must update MasterMgmtClass data correctly
