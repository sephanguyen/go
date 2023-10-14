@blocker
Feature: Sync student from jprep
  Scenario Outline: Jpref sync student to our system
    When jprep sync "<new_student>" students with action "<action_new_student>" and "<existed_student>" students with action "<action_existed_student>"
    Then these students must be store in our system
    And data log split store correct "<log_status>"

    Examples:
      | new_student | action_new_student   | existed_student | action_existed_student | log_status |
      | 3           | ACTION_KIND_UPSERTED | 2               | ACTION_KIND_UPSERTED   | SUCCESS    |
      | 3           | ACTION_KIND_UPSERTED | 2               | ACTION_KIND_DELETED    | SUCCESS    |
