Feature: Delete courses

    Scenario Outline: user with invalid role try to delete courses
        Given a random number
            And a signed in "<signed as>"
            And a list of "content" courses are existed in DB
            And a DeleteCourseRequest with id "course-valid-2"
        When user delete courses
        Then returns "<msg>" status code

        Examples:
            | signed as       | msg              |
            | unauthenticated | Unauthenticated  |
            | student         | PermissionDenied |
            | parent          | PermissionDenied |

    Scenario: Admin delete valid courses
        Given a random number
            And a signed in "school admin"
            And a list of "content" courses are existed in DB
            And a DeleteCourseRequest with id "course-valid-2"
            And a DeleteCourseRequest with id "course-1-JP"
        When user delete courses
        Then returns "OK" status code

    Scenario: Admin delete invalid courses
        Given a random number
            And a signed in "school admin"
            And a list of "content" courses are existed in DB
            And a DeleteCourseRequest with id "course-valid-2"
            And a DeleteCourseRequest with id "course-deleted"
        When user delete courses
        Then returns "NotFound" status code
            And yasuo must store activity logs "/manabie.yasuo.CourseService/DeleteCourses_NotFound"

    Scenario Outline: user with invalid role try to delete live courses
        Given a random number
            And a signed in "<signed as>"
            And a list of "live" courses are existed in DB
            And a DeleteCourseRequest with id "live-course-valid-2"
        When user delete courses
        Then returns "<msg>" status code

        Examples:
            | signed as       | msg              |
            | unauthenticated | Unauthenticated  |
            | student         | PermissionDenied |
            | parent          | PermissionDenied |
            | teacher         | PermissionDenied |