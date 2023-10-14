Feature: Handle SyncStudentPackageEvent
  Background: courses is assigned study plan
    Given a valid "teacher" token
      And a valid course background
      And valid assignment in db

    Scenario: Handle SyncStudentPackageEvent with ActionKind_ACTION_KIND_UPSERTED
      Given an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_UPSERTED
      When send "SyncStudentPackageEvent" topic "SyncStudentPackage.Synced" to nats js
      Then our system must upsert CourseStudent data correctly
        And our system must create new study plan for each course student

     Scenario: Handle SyncStudentPackageEvent with ActionKind_ACTION_KIND_DELETED
      Given an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_DELETED
      When send "SyncStudentPackageEvent" topic "SyncStudentPackage.Synced" to nats js
      Then our system must update CourseStudent data correctly
        And our system must remove all course student study plan
