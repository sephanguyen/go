@blocker
Feature: Import Parent and assign to student

    Background: Sign in with role "staff granted role school admin"
        Given a signed in "staff granted role school admin"

    Scenario Outline: Import valid csv file "<row condition>"
        When "school admin" import 1 parent(s) and assign to 1 student(s) with valid payload having "<row condition>"
        Then the valid parent lines are imported successfully
        And returns "OK" status code

        Examples:
            | row condition       |
            | valid rows          |
            | only mandatory rows |


    # this case will take long time to run and leads to context time out
    @quarantined
    Scenario: Import parent valid csv file 1000 rows
        When school admin create a student with "general info" and "<condition>" by GRPC
        When "school admin" import 1000 parent(s) and assign to 1 student(s) with valid payload having "valid rows"
        Then the valid parent lines are imported successfully
        And returns "OK" status code

    Scenario: Import parent invalid csv file 1001 rows
        When "school admin" import 1001 parent(s) and assign to 1 student(s) with valid payload having "valid rows"
        Then the invalid parent lines are returned with error code "invalidNumberRow"

    Scenario Outline: Import invalid csv file "<invalid format>"
        When "school admin" import 1 parent(s) and assign to 1 student(s) with valid payload having "<invalid format>"
        Then the invalid parent lines are returned with error code "<error code>"

        Examples:
            | invalid format                                                                    | error code                          |
            | have 1 invalid row with missing mandatory                                         | missingMandatory                    |
            | have 1 invalid row with student_email does not exist in db                        | notFollowParentTemplate             |
            | have 1 invalid row with student_email is existed parent email                     | notFollowParentTemplate             |
            | have 1 invalid row with 1 student_email exists and 1 student_email does not exist | notFollowParentTemplate             |
            | have 1 invalid row with relationship is zero                                      | notFollowParentTemplate             |
            | have 1 invalid row with email duplicated in payload                               | duplicationRow                      |
            | have 1 invalid row with email existed in database                                 | alreadyRegisteredRow                |
            | have 1 invalid row with phone_number duplication rows                             | duplicationRow                      |
            | have 1 invalid row with missing first_name                                        | missingMandatory                    |
            | have 1 invalid row with missing last_name                                         | missingMandatory                    |
            | have 1 invalid row with external_user_id duplicated in payload                    | duplicationRow                      |
            | have 1 invalid row with external_user_id existed in database                      | alreadyRegisteredRow                |
            | have 1 invalid row with not match relationship and email student                  | notMatchRelationshipAndEmailStudent |

# todo: add more case for invalid format user tags
