Feature: Create live lesson

    Background:
        # And a generate school
        And a teacher account
        And a class
        And a live course


    Scenario Outline: user try to create live lesson
        Given signed as "<signed as>" account
        And a CreateLiveLessonRequest with start time is "2022-06-30 06:30:00" and end time is "2022-07-30 06:30:00"
        When user create live lesson
        Then yasuo returns "OK" status code
        And yasuo must store live lesson with start time is "2022-06-30 06:30:00" and end time is "2022-07-30 06:30:00" and push msg "CreateLesson" subject "Lesson.Created" to nats
        And tom must record new conversation and record new conversation_lesson

        Examples:
            | signed as    |
            | admin        |
            | school admin |
