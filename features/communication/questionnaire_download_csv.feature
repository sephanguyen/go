Feature: View questionnaire CSV using TargetGroup and IndividualTarget

    Scenario: Staff download questionnaire answers csv file
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "2" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "<questions>" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "all individual"
        Then notificationmgmt services must send notification to user
        Then parent answer questionnaire for "<answered for students>"
        And "student" answer questionnaire for themself
        And "student individual" answer questionnaire for themself
        And current staff download questionnaire answers csv file successfully with "<num_rows>" rows and "<num_answers>" submissions
        Examples:
            | receivers | answered for students | num_rows | num_answers | questions                                                              |
            | all       | 1,2                   | 6        | 5           | 1.multiple_choice.required, 2.free_text, 3.check_box.required          |
            | all       | 1,2                   | 6        | 5           | 1.multiple_choice.required, 2.free_text.required, 3.check_box.required |
            | all       | 1                     | 6        | 4           | 1.multiple_choice.required, 2.check_box.required, 3.free_text          |
            | all       | 1                     | 6        | 4           | 1.check_box.required, 2.multiple_choice.required, 3.free_text          |
