@runsequence
Feature: Sync class and members from JPREP

Scenario Outline: JPREP sync class to our system
   When JPREP sync "<new_class>" class with action "<action_new_class>" and "<existed_class>" class with action "<action_existed_class>"
   Then these classes must be store in our system
   And these new classes must be store in our system
   And data log split store correct "<log_status>"

  Examples:
    | new_class | action_new_class     | existed_class | action_existed_class | log_status |
    | 3         | ACTION_KIND_UPSERTED | 2             | ACTION_KIND_UPSERTED | SUCCESS    |
    | 3         | ACTION_KIND_UPSERTED | 2             | ACTION_KIND_DELETED  | SUCCESS    |

Scenario Outline: JPREP sync student of class to our system
   When some courses existed in db
   Then JPREP sync "<new_class_member>" class members with action "<action_new_class_member>" and "<existed_class_member>" class members with action "<action_existed_class_member>"
   And these class members must be store in out system
   And these new class members must be stored in out system
   And data log split store correct "<log_status>"

  Examples:
    | new_class_member | action_new_class_member | existed_class_member | action_existed_class_member | log_status |
    | 3                | ACTION_KIND_UPSERTED    | 2                    | ACTION_KIND_UPSERTED        | SUCCESS    |
    | 3                | ACTION_KIND_UPSERTED    | 2                    | ACTION_KIND_DELETED         | SUCCESS    |

