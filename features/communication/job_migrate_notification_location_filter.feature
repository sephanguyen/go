Feature: migrate notification location filter
    Scenario: "staff granted role school admin" create 3 notification
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" courses
        And current staff upsert notification to "student,parent" and "random" course and "random" grade and "random" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And current staff upsert notification to "student,parent" and "random" course and "random" grade and "all" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And current staff upsert notification to "student,parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        When run migration script
        Then data of target group is correctly migrated