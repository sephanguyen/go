Feature: Delete recorded video when lesson deleted

    Background:
        Given user signed in as school admin 
        When enter a school
        And have some centers
        And have some teacher accounts
        And have some student accounts
        And have some courses
        And have some student subscriptions
        And an existing a virtual classroom session
        And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

    Scenario: Recorded videos will be deleted when admin delete lesson
        Given "teacher" signin system
        When user join a virtual classroom session
        Then returns "OK" status code
        When user start to recording
        Then returns "OK" status code
        And start recording state is updated
        When user stop recording
        Then returns "OK" status code
        And recorded videos are saved
        When user deletes a lesson
        And returns "OK" status code
        Then Lessonmgmt must push msg "DeleteLesson" subject "Lesson.Deleted" to nats
        And media and recorded video will be deleted in db and cloud storage
