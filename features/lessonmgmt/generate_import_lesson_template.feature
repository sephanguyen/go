Feature: Generate Import Lesson CSV Template

    Scenario: Generate Import Lesson CSV Template
        Given user signed in as school admin
        And "disable" Unleash feature with feature name "Lesson_LessonManagement_BackOffice_ImportLessonByCSVV2"
        When user download sample csv file to import lesson
        Then returns a lesson template csv with columns: "partner_internal_id,start_date_time,end_date_time,teaching_method"

    @runsequence
    Scenario: Generate Import Lesson CSV Template and Unleash turned on
        Given user signed in as school admin
        And "enable" Unleash feature with feature name "Lesson_LessonManagement_BackOffice_ImportLessonByCSVV2"
        When user download sample csv file to import lesson
        Then returns a lesson template csv with columns: "partner_internal_id,start_date_time,end_date_time,teaching_method,teaching_medium,teacher_ids,student_course_ids"
