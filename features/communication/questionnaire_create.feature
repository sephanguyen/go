Feature: Create questionnaire
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students with "1" parents
        And school admin creates "1" course
        And school admin add packages data of those courses for each student

    @blocker
    Scenario: Happy case
        Given current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then questionnaire and qn_question are correctly stored in db
        Examples:
            | resubmit_allowed | questions                                                                           |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required                       |
            | true             | 1.free_text, 2.free_text.required, 3.check_box.required, 4.multiple_choice.required |

    Scenario: Upsert removing questionnaire
        Given current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then questionnaire and qn_question are correctly stored in db
        When admin remove questionnaire and upsert notification again
        Then notification has no questionnaire in DB
        And questionnaire and qn_question are soft deleted in db
        Examples:
            | resubmit_allowed | questions                                                     |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required |

    Scenario: Upsert update questions in questionnaire
        Given current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<1st questions>" respectively
        And current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And update questionnaire with resubmit allowed "<2nd resubmit_allowed>", questions "<2nd questions>" respectively
        When admin upsert notification with updated questionnaire
        Then questionnaire and qn_question are correctly stored in db
        And questions with order_index "<deleted question ids>" are soft deleted
        Examples:
            | resubmit_allowed | 1st questions                                                 | 2nd resubmit_allowed | 2nd questions                           | deleted question ids |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | true                 | 1.free_text.required, 2.multiple_choice | 1,2,3                |
            | false            | 1.multiple_choice                                             | true                 | 1.free_text, 2.free_text, 3.free_text   | 1                    |

    Scenario: Upsert update questions in questionnaire with full questionnaire_id, questionnaire_question_id in payload request (test bulk force upsert)
        Given current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<1st questions>" respectively
        And current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And fill all questionnaire_id and questionnaire_question_id from db into payload
        And update questionnaire with resubmit allowed "<2nd resubmit_allowed>", questions "<2nd questions>" respectively
        When admin upsert notification with updated questionnaire
        Then questionnaire and qn_question are correctly stored in db
        And questions with order_index "<deleted question ids>" are soft deleted
        Examples:
            | resubmit_allowed | 1st questions                                                 | 2nd resubmit_allowed | 2nd questions                           | deleted question ids |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | true                 | 1.free_text.required, 2.multiple_choice | 3                    |
            | false            | 1.multiple_choice                                             | true                 | 1.free_text, 2.free_text, 3.free_text   |                      |

    Scenario: Upsert notification with questionnaire and discard notification
        Given current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then questionnaire and qn_question are correctly stored in db
        When current staff discards notification
        Then notification is discarded
        Examples:
            | resubmit_allowed | questions                                                     |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required |

    Scenario: Upsert notification with questionnaire template
        Given current staff create a questionnaire with resubmit allowed "<1st resubmit_allowed>", questions "<1st questions>" respectively
        And current staff create questionnaire template from questionnaire
        And current staff update with updated questionnaire template with resubmit allowed "<2nd resubmit_allowed>", questions "<2nd questions>" respectively
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then questionnaire and qn_question are correctly stored in db
        Examples:
            | 1st resubmit_allowed | 1st questions                                                 | 2nd resubmit_allowed | 2nd questions                           |
            | false                | 1.multiple_choice, 2.free_text.required, 3.check_box.required | true                 | 1.free_text.required, 2.multiple_choice |
