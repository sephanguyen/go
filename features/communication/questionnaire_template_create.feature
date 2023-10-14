Feature: Create questionnaire template
    Background: 
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations

    Scenario Outline: user create new questionnaire template and update successfully
        Given current staff create a questionnaire template with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        Then questionnaire template and question are correctly stored in db
        Examples:
            | resubmit_allowed | questions                                                                           |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required                       |
            | true             | 1.free_text, 2.free_text.required, 3.check_box.required, 4.multiple_choice.required |
    
    Scenario Outline: user update questions in questionnaire template
        Given current staff create a questionnaire template with resubmit allowed "<1st resubmit_allowed>", questions "<1st questions>" respectively
        Then questionnaire template and question are correctly stored in db
        When current staff update questionnaire template with resubmit allowed "<2nd resubmit_allowed>", questions "<2nd questions>" respectively
        Then questionnaire template and question are correctly stored in db
        And questions with order_index "<deleted question ids>" are soft deleted
        Examples:
            | 1st resubmit_allowed | 1st questions                                                 | 2nd resubmit_allowed | 2nd questions                           | deleted question ids |
            | false                | 1.multiple_choice, 2.free_text.required, 3.check_box.required | true                 | 1.free_text.required, 2.multiple_choice | 1,2,3                |
            | false                | 1.multiple_choice                                             | true                 | 1.free_text, 2.free_text, 3.free_text   | 1                    |

    Scenario: user upsert questionnaire template but the name is exist
        Given current staff create a questionnaire template with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        And current staff create questionnaire template with name is exist
        Then return error questionnaire template name existed
        Examples:
            | resubmit_allowed | questions                                                                           |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required                       |
