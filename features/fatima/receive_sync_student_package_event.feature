@quarantined
Feature: Handle SyncStudentPackageEvent
  Scenario: Handle SyncStudentPackageEvent with ActionKind_ACTION_KIND_UPSERTED
    Given an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_UPSERTED
    When bob send SyncStudentPackageEvent to nats
    Then our system must create StudentPackage data correctly
    And our system must createStudentPackage access path

  Scenario: Handle SyncStudentPackageEvent with ActionKind_ACTION_KIND_DELETED
    Given an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_DELETED
    When bob send SyncStudentPackageEvent to nats
    Then our system must update StudentPackage data correctly