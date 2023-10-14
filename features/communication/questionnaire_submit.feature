Feature: Submit questionnaire
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
 
    @blocker
    Scenario: Happy case
        Given current staff creates "2" students with the same parent
        And parent login to Learner App
        And update user device token to an "valid" device token
        And current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        And current staff upsert notification and send to parent
        When parent submit answers list for questions in questionnaire for student 0 with "full" answers
        Then parent see answers in previous step calling RetrieveNotificationDetail with notification for student 0
        And parent does not see answers in previous step calling RetrieveNotificationDetail with notification for student 1
        Examples:
            | resubmit_allowed | questions                                                     | user_group |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | parent     |

    Scenario: Expired questionnaire
        Given current staff creates "1" students with "1" parent info
        And parent login to Learner App
        And update user device token to an "valid" device token
        And current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        And current staff upsert notification and send to parent
        And change expiration_date of questionnaire in db to yesterday
        When parent submit answers list for questions in questionnaire for student 0 with "full" answers
        Then parent receive "expired questionnaire, you cannot submit questionnaire after the expiration date has passed" error
        Examples:
            | resubmit_allowed | questions                                                     |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required |

    Scenario: Required questions
        Given current staff creates "1" students with "1" parent info
        And parent login to Learner App
        And update user device token to an "valid" device token
        And current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        And current staff upsert notification and send to parent
        When parent submit answers list for questions in questionnaire for student 0 with "missing" answers
        Then parent receive "missing required question, you need to fill all required question" error
        Examples:
            | resubmit_allowed | questions                                                     |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required |

    Scenario: Resubmit not allowed
        Given current staff creates "1" students with "1" parent info
        And parent login to Learner App
        And update user device token to an "valid" device token
        And current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        And current staff upsert notification and send to parent
        And parent submit answers list for questions in questionnaire for student 0 with "full" answers
        When parent re-submit answers list for questions in questionnaire for student 0
        Then parent receive "<error type>" error
        Examples:
            | resubmit_allowed | questions                                                     | error type                                                    |
            | true             | 1.multiple_choice, 2.free_text.required, 3.check_box.required | none                                                          |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | resubmit not allowed, you cannot re-submit this questionnaire |

    Scenario: Validate request answer
        Given current staff creates "1" students with "1" parent info
        And parent login to Learner App
        And update user device token to an "valid" device token
        And current staff create a questionnaire with resubmit allowed "true", questions "1.multiple_choice, 2.free_text.required, 3.check_box.required" respectively
        And current staff upsert notification and send to parent
        When parent submit answers list for questions in questionnaire for student 0 with "<error type>" answers
        Then parent receive "<error type>" error
        Examples:
            | error type                                                                  |
            | you cannot answer the question not in questionnaire                         |
            | you cannot have multiple answer for multiple choices and free text question |
            | your answer doesn't in questionnaire question choices                       |

    Scenario: Parent submit for them self (need to update later)
        Given current staff creates "1" students with the same parent
        And parent login to Learner App
        And update user device token to an "valid" device token
        And current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        And current staff upsert notification and send to parent
        And current staff set target for notification of student 0 to parent
        When parent submit answers list for questions in questionnaire for themself with "full" answers
        Then parent see answers in previous step calling RetrieveNotificationDetail with notification for themself
        Examples:
            | resubmit_allowed | questions                                                     | user_group |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | parent     |
