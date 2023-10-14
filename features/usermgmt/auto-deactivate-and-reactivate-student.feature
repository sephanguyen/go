@blocker
Feature: Auto deactivate and reactivate students
  As a school staff
  I need to be able to upsert a new student

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Auto deactivate and reactivate students when upsertStudent
    When school admin creates a student with "<activeStatusAmount>" "<activeStatus>" status is active and "<inActiveStatusAmount>" "<inActiveStatus>" status is inactive
    Then school admin sees student is "<isDeactivated>"
    Examples:
      | activeStatusAmount | activeStatus                | inActiveStatusAmount | inActiveStatus | isDeactivated |
      | 0                  | non-withdrawn               | 1                    | non-withdrawn  | activated     |
      | 0                  | withdrawn                   | 1                    | withdrawn      | activated     |
      | 1                  | non-withdrawn               | 1                    | withdrawn      | activated     |
      | 1                  | non-withdrawn and withdrawn | 0                    | non-withdrawn  | activated     |
      | 1                  | withdrawn                   | 1                    | withdrawn      | deactivated   |
      | 1                  | withdrawn                   | 1                    | non-withdrawn  | deactivated   |
      | 2                  | withdrawn                   | 0                    | non-withdrawn  | deactivated   |

  Scenario Outline: Auto deactivate and reactivate students when upsertStudent with "<activeStatus>" status is active and "<statusInFuture>" status is inactive
    When school admin creates a student with "0" "<activeStatus>" status is active and "0" "<statusInFuture>" status is inactive
    Then school admin sees student is "<isDeactivated>"
    Examples:
      # statusToDeactivate = [WITHDRAWN, GRADUATED, NON-POTENTIAL]
      # statusInFuture = enrollment statuses have start_date > today [ENROLLED, WITHDRAWN, GRADUATED, LOA]
      # activeStatus = enrollment statuses have start_date <= today
      | activeStatus                | statusInFuture | isDeactivated |
      |                             | ENROLLED       | activated     |
      |                             | WITHDRAWN      | activated     |
      |                             | GRADUATED      | activated     |
      |                             | LOA            | activated     |
      | ENROLLED                    | WITHDRAWN      | activated     |
      | POTENTIAL                   | WITHDRAWN      | activated     |
      | LOA                         | WITHDRAWN      | activated     |
      | TEMPORARY                   | WITHDRAWN      | activated     |
      | POTENTIAL,WITHDRAWN         |                | activated     |
      | POTENTIAL,GRADUATED         |                | activated     |
      | POTENTIAL,NON-POTENTIAL     |                | activated     |
      | WITHDRAWN                   |                | deactivated   |
      | GRADUATED                   |                | deactivated   |
      | NON-POTENTIAL               |                | deactivated   |
      | WITHDRAWN                   | ENROLLED       | deactivated   |
      | GRADUATED                   | ENROLLED       | deactivated   |
      | NON-POTENTIAL               | ENROLLED       | deactivated   |
      | WITHDRAWN,GRADUATED         |                | deactivated   |
      | GRADUATED,NON-POTENTIAL     |                | deactivated   |
      | WITHDRAWN,NON-POTENTIAL     |                | deactivated   |
      | GRADUATED,GRADUATED         |                | deactivated   |
      | NON-POTENTIAL,NON-POTENTIAL |                | deactivated   |

  Scenario Outline: Auto deactivate and reactivate students when sync Order
    Given school admin has created a student being "<statusBeforeSyncOrder>"
    When school admin create "<orderRequest>" by Orders function
    Then school admin sees student is "<statusAfterSyncOrder>"
    Examples:
      | statusBeforeSyncOrder | orderRequest       | statusAfterSyncOrder |
      | deactivated           | Enrollment Request | activated            |
      | deactivated           | Order              | activated            |
      | activated             | Withdrawal Request | deactivated          |
      | activated             | Graduate Order     | deactivated          |

  Scenario Outline: Auto deactivate and reactivate students when sync Order "<statusBeforeSyncOrder>" "<orderRequest>"  "<statusAfterSyncOrder>"
    Given school admin has created a student being "<statusBeforeSyncOrder>"
    When school admin create "<orderRequest>" by Orders function
    Then school admin sees student is "<statusAfterSyncOrder>"
    Examples:
      | statusBeforeSyncOrder | orderRequest       | statusAfterSyncOrder |
      | WITHDRAWN             | Enrollment Request | activated            |
      | WITHDRAWN             | Order              | activated            |
      | WITHDRAWN             | Graduate Order     | deactivated          |
      | WITHDRAWN             | Withdrawal Request | deactivated          |
      | GRADUATED             | Enrollment Request | activated            |
      | GRADUATED             | Order              | activated            |
      | GRADUATED             | Graduate Order     | deactivated          |
      | GRADUATED             | Withdrawal Request | deactivated          |
      | NON-POTENTIAL         | Enrollment Request | activated            |
      | NON-POTENTIAL         | Order              | activated            |
      | NON-POTENTIAL         | Graduate Order     | deactivated          |
      | NON-POTENTIAL         | Withdrawal Request | deactivated          |
      | ENROLLED              | Enrollment Request | activated            |
      | ENROLLED              | Order              | activated            |
      | ENROLLED              | Graduate Order     | deactivated          |
      | ENROLLED              | Withdrawal Request | deactivated          |
      | LOA                   | Enrollment Request | activated            |
      | LOA                   | Order              | activated            |
      | LOA                   | Graduate Order     | deactivated          |
      | LOA                   | Withdrawal Request | deactivated          |
      | POTENTIAL             | Enrollment Request | activated            |
      | POTENTIAL             | Order              | activated            |
      | POTENTIAL             | Graduate Order     | deactivated          |
      | POTENTIAL             | Withdrawal Request | deactivated          |

  Scenario Outline: Auto deactivate and reactivate students when trigger daily job
    Given school admin has created a student who will be "<statusBeforeRunJob>"
    When system run daily job to deactivate and reactivate students
    Then school admin sees student is "<statusAfterRunJob>"
    Examples:
      | statusBeforeRunJob | statusAfterRunJob |
      | deactivated        | deactivated       |
      | activated          | activated         |

  Scenario Outline: Auto deactivate and reactivate students when trigger daily job
    Given school admin has created a student who has "<status>"
    When system run daily job to deactivate and reactivate students
    Then school admin sees student is "<isDeactivated>"
    Examples:
      | status        | isDeactivated |
      | ENROLLED      | activated     |
      | POTENTIAL     | activated     |
      | LOA           | activated     |
      | NON-POTENTIAL | deactivated   |
      | WITHDRAWN     | deactivated   |
      | GRADUATED     | deactivated   |
