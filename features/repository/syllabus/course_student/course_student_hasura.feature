Feature: Test CourseStudent hasura

    Scenario Outline: List students in course
        Given some students assigned to <courses>
        When user get students by call <func>
        Then our system must return <data> correctly

        Examples:
            | func                          | data                                               | courses     |
            | CourseStudentsListByCourseIds | course students by course ids                      | many course |
            | CourseSudentsList             | course students by course id                       | course      |
            | CourseSudentsListV2           | course students by course id with limit and offset | course      |

