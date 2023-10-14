@quarantined
Feature: Handle JprefMasterRegistration
    Scenario: Handle JprefMasterRegistration with ActionKind_ACTION_KIND_UPSERTED
      Given an valid JprefMasterRegistration with ActionKind_ACTION_KIND_UPSERTED
      When send "JprefMasterRegistration" topic "SyncMasterRegistration.Synced" to nats js
      Then our system must upsert CourseClass data correctly

     Scenario: Handle JprefMasterRegistration with ActionKind_ACTION_KIND_DELETED
      Given an valid JprefMasterRegistration with ActionKind_ACTION_KIND_DELETED
      When send "JprefMasterRegistration" topic "SyncMasterRegistration.Synced" to nats js
      Then our system must update CourseClass data correctly
