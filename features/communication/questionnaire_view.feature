Feature: View questionnaire detail using TargetGroup and IndividualTarget

    Scenario: Send notification attached with questionnaire using TargetGroup
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "2" students with "1" parents info for each student
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "<individual receivers>"
        Then notificationmgmt services must send notification to user
        And "<receivers>" see "<count>" unanswered questionnaire in notification bell with correct detail
        And "<individual receivers>" see "<individual count>" unanswered questionnaire in notification bell with correct detail
        Examples:
            | resubmit_allowed | questions                                                     | receivers | count | individual receivers | individual count |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | student   | 1     | all individual       | 1                |
            | true             | 1.multiple_choice.required, 2.check_box.required              | student   | 1     | student individual   | 1                |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | parent    | 1     | parent individual    | 1                |
            | true             | 1.multiple_choice.required, 2.check_box.required              | parent    | 1     | all individual       | 1                |

    Scenario: Send notification attached with questionnaire and parent has many children
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "2" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "<individual receivers>"
        Then notificationmgmt services must send notification to user
        And "<receivers>" see "<count>" unanswered questionnaire in notification bell with correct detail
        And "<individual receivers>" see "<individual count>" unanswered questionnaire in notification bell with correct detail
        Examples:
            | resubmit_allowed | questions                                                     | receivers | count | individual receivers | individual count |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | student   | 1     | all individual       | 1                |
            | true             | 1.multiple_choice.required, 2.check_box.required              | student   | 1     | student individual   | 1                |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | parent    | 2     | parent individual    | 1                |
            | true             | 1.multiple_choice.required, 2.check_box.required              | parent    | 2     | student individual   | 1                |

    Scenario: Staff and parent see empty answered questionaire with fully question is answered
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "2" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice, 2.free_text, 3.check_box" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "parent individual"
        Then notificationmgmt services must send notification to user
        Then parent answer questionnaire for "<answered for students>" with empty answer
        And "parent individual" answer questionnaire for themself
        Then "parent" see 2 notifications in notification bell with correct detail and answer status
        And "parent individual" see 1 notifications in notification bell with correct detail and answer status
        And current staff see "<number of answers in filter>" answers in questionnaire answers list with search_text is a full name of target "all" and total answers is "<total count>" and fully answer for question
        Examples:
            | receivers | answered for students | number of answers in filter | total count |
            | parent    | 1,2                   | 3                           | 3           |
            | parent    | 1                     | 2                           | 2           |

    @blocker
    Scenario: Staff and parent see answered questionaire
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "2" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice, 2.free_text.required, 3.check_box.required" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "parent individual"
        Then parent answer questionnaire for "<answered for students>"
        And "parent individual" answer questionnaire for themself
        Then "parent" see 2 notifications in notification bell with correct detail and answer status
        And current staff see "<number of answers in filter>" answers in questionnaire answers list with search_text is a full name of target "all" and total answers is "<total count>" and fully answer for question
        Examples:
            | receivers | answered for students | number of answers in filter | total count |
            | parent    | 1,2                   | 3                           | 3           |
            | parent    | 1                     | 2                           | 2           |

    Scenario: Staff and parent see answered questionaire (case parent submit for themself)
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students with "1" parents info for each student
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice, 2.free_text.required, 3.check_box.required" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "parent individual"
        Then notificationmgmt services must send notification to user
        Then "parent" answer questionnaire for themself
        And "parent individual" answer questionnaire for themself
        Then "parent" see 1 notifications in notification bell with correct detail and answer status
        And "parent individual" see 1 notifications in notification bell with correct detail and answer status
        And current staff see "<number of answers in filter>" answers in questionnaire answers list with search_text is a full name of target "all" and total answers is "<total count>" and fully answer for question
        Examples:
            | receivers | number of answers in filter | total count |
            | parent    | 2                           | 2           |

    Scenario: Staff, parent and student see answered questionaire
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "2" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice.required, 2.free_text, 3.check_box.required" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "all individual"
        Then notificationmgmt services must send notification to user
        Then parent answer questionnaire for "<answered for students>"
        And "student" answer questionnaire for themself
        And "student individual" answer questionnaire for themself
        Then "parent" see 2 notifications in notification bell with correct detail and answer status
        And "student" see 1 notifications in notification bell with correct detail and answer status
        And "all individual" see 1 notifications in notification bell with correct detail and answer status
        And current staff see "<number of answers in filter>" answers in questionnaire answers list with search_text is a full name of target "<search target>" and total answers is "<total count>" and fully answer for question
        Examples:
            | receivers | answered for students | search target | number of answers in filter | total count |
            | all       | 1,2                   | all           | 5                           | 5           |
            | all       | 1,2                   | student       | 1                           | 1           |
            | all       | 1                     | parent        | 1                           | 1           |
            | all       | 1                     | student       | 1                           | 1           |

    @blocker
    Scenario: Pagination get answers by filters
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "5" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice, 2.free_text.required, 3.check_box" respectively
        When current staff send a notification with attached questionnaire to "all" and individual "all individual"
        Then parent answer questionnaire for "1,2,3"
        And "student" answer questionnaire for themself
        And "student individual" answer questionnaire for themself
        Then current staff get answers list with limit is "<limit number>" and offset is "<offset number>"
        And returns "OK" status code
        And current staff see "<number of answers in filter>" answers in questionnaire answers list and total items is "9" and previous offset is "<previous offset number>" and next offset is "<next offset number>"
        Examples:
            | offset number | limit number | number of answers in filter | previous offset number | next offset number |
            | 0             | 1            | 1                           | 0                      | 1                  |
            | 0             | 5            | 5                           | 0                      | 5                  |
            | 6             | 3            | 3                           | 3                      | 9                  |
            | 3             | 7            | 6                           | 0                      | 9                  |

    Scenario: Pagination get answers by filters by staff with granted role
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "5" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice, 2.free_text.required, 3.check_box" respectively
        When current staff send a notification with attached questionnaire to "all" and individual "all individual"
        Then parent answer questionnaire for "1,2,3"
        And "student" answer questionnaire for themself
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        Then current staff get answers list with limit is "1" and offset is "0"
        And returns "OK" status code
        And current staff see "1" answers in questionnaire answers list and total items is "8" and previous offset is "0" and next offset is "1"
        Examples:
            | staff with granted role           |
            | staff granted role school admin   |
            # | staff granted role school admin and teacher |
            | staff granted role teacher        |
            | staff granted role hq staff       |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario: Pagination get answers by filters by staff with granted denied role
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "5" students with the same parent
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice, 2.free_text.required, 3.check_box" respectively
        When current staff send a notification with attached questionnaire to "all" and individual "none"
        Then parent answer questionnaire for "1,2,3"
        And "student" answer questionnaire for themself
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        Then current staff get answers list with limit is "1" and offset is "0"
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    Scenario: individual receivers see unanswered questionnaire using UserNotificationID
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        When current staff send a notification with attached questionnaire to "<receivers>" and individual "<individual receivers>"
        Then notificationmgmt services must send notification to user
        And "<individual receivers>" see "<individual count>" unanswered questionnaire in notification bell with correct detail using new api RetrieveNotificationDetail
        Examples:
            | resubmit_allowed | questions                                                     | receivers | individual receivers | individual count |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | none      | all individual       | 1                |
            | true             | 1.multiple_choice.required, 2.check_box.required              | none      | student individual   | 1                |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required | none      | parent individual    | 1                |
            | true             | 1.multiple_choice.required, 2.check_box.required              | none      | all individual       | 1                |
