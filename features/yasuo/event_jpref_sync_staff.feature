Feature: Sync staff from JPREP

Scenario Outline: JPREP sync staff to our system
   When JPREP sync "<new_staff>" staffs with action "<action_new_staff>" and "<existed_staff>" staffs with action "<action_existed_staff>"
   Then these staffs must be store in our system
   And data log split store correct "<log_status>"

  Examples:
    | new_staff   | action_new_staff     | existed_staff   | action_existed_staff     | log_status |
    | 3           | ACTION_KIND_UPSERTED | 2               | ACTION_KIND_UPSERTED     | SUCCESS    |
    | 3           | ACTION_KIND_UPSERTED | 2               | ACTION_KIND_DELETED      | SUCCESS    |


Scenario Outline: JPREP re-sync deleted staff and get self profile
   When after the deleted staff were "<sync_action>"
   Then they login our system and "<status_action>" get self-profile info

  Examples:
    | sync_action          | status_action |
    | ACTION_KIND_UPSERTED | can           |
    | ACTION_KIND_DELETED  | cannot        |
