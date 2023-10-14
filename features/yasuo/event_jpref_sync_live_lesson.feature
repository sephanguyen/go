Feature: Sync lesson and members from jprep

Scenario Outline: Jprep sync lesson to our system
   When jprep sync "<new_lesson>" lesson with action "<action_new_lesson>" and "<existed_lesson>" lesson with action "<action_existed_lesson>"
   Then these lessons must be store in our system correctly
   And data log split store correct "<log_status>"
   And last time received message store correctly with config key "JprepMasterRegistrationLastTime_1"

  Examples:
    | new_lesson | action_new_lesson    | existed_lesson | action_existed_lesson | log_status |
    | 3          | ACTION_KIND_UPSERTED | 2              | ACTION_KIND_UPSERTED  | SUCCESS    |
    | 3          | ACTION_KIND_UPSERTED | 2              | ACTION_KIND_DELETED   | SUCCESS    |

Scenario Outline: Jprep sync student of lesson to our system
   When jprep sync "<new_lesson_member>" lesson members with action "<action_new_lesson_member>" and "<existed_lesson_member>" lesson members with action "<action_existed_lesson_member>" at <hours> ago
   Then these lesson members must be store in our system

  Examples:
    | new_lesson_member | action_new_lesson_member | existed_lesson_member | action_existed_lesson_member | hours  |
    | 3                 | ACTION_KIND_UPSERTED     | 2                     | ACTION_KIND_UPSERTED         | 0      |
    | 3                 | ACTION_KIND_UPSERTED     | 2                     | ACTION_KIND_DELETED          | 0      |

Scenario Outline: Subscribes receive messages in the past
   When jprep sync "<new_lesson_member>" lesson members with action "<action_new_lesson_member>" and "<existed_lesson_member>" lesson members with action "<action_existed_lesson_member>" at <hours> ago
   Then these no lesson members store in our system
  Examples:
    | new_lesson_member | action_new_lesson_member | existed_lesson_member | action_existed_lesson_member | hours  |
    | 1                 | ACTION_KIND_UPSERTED     | 0                     | ACTION_KIND_UPSERTED         | 2      |

Scenario: Jprep sync lesson to our system with existed lesson with updated lesson type
  Given some existed lesson in database
    And these lesson updated type "LESSON_TYPE_HYBRID"
  When jprep sync some new lesson with action "ACTION_KIND_UPSERTED" and some existed lesson with action "ACTION_KIND_UPSERTED"
  Then these lessons must be store in our system correctly

Scenario: Jprep sync deleted lesson to our system
  Given some existed lesson in database
    And these lesson have to deleted
  When jprep sync some new lesson with action "ACTION_KIND_UPSERTED" and some existed lesson with action "ACTION_KIND_UPSERTED"
  Then these lessons must be store in our system correctly

  Scenario: Jprep sync lesson members, upserting will remove lessons not included in request
    Given jprep sync some lessons to student
    When jprep resync lesson members but excluding a lesson
    Then yasuo must push event removing lesson members to "SyncStudentLessonsConversations.Synced" for excluded lesson