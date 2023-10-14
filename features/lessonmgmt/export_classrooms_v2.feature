Feature: Export Classrooms
    Background:
        Given user signed in as school admin
        And have some centers
        And have some classrooms
        And "enable" Unleash feature with feature name "Lesson_LessonManagement_BackOffice_UpdateClassroomFlow"

    Scenario Outline: Export all classrooms
        Given user signed in as school admin
        When user export classrooms
        Then returns "OK" status code
        And returns classrooms in csv with "location_id,location_name,classroom_id,classroom_name,remarks,room_area,seat_capacity" columns
