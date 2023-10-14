@runsequence
Feature: sync course from jpref

  Scenario Outline: Jpref sync courses to our system
	When jpref sync "<new_course>" courses with action "<action_new_course>" and "<existed_course>" course with action "<action_existed_course>" to our system
	Then these courses must be store in our system
    And data log split store correct "<log_status>"

    Examples:
            | new_course	| action_new_course     | existed_course	| action_existed_course | log_status |
            | 3				| ACTION_KIND_UPSERTED	| 2					| ACTION_KIND_UPSERTED	| SUCCESS    |
            | 2				| ACTION_KIND_UPSERTED	| 3					| ACTION_KIND_DELETED	| SUCCESS    |

  Scenario: Jpref sync courses to our system with some existed courses in db
  Given some courses existed in db
    And some courses must have icon
    And some courses must have book
	When jpref sync arbitrary number new courses and existed courses with action "ACTION_KIND_UPSERTED" to our system
	Then these courses have to save correctly
