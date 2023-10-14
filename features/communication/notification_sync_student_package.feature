Feature: Sync student course packages
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations

    @blocker
    Scenario Outline: Sync student course packages for notification successfully
        Given school admin creates "<number-of-students>" students
        And school admin creates "<number-of-courses>" courses with "1" classes for each course
        When assigning course packages to existing students
        And waiting for all student course packages to be synced
        Then sync create student courses and class members successfully with "<number-of-students>" students and "<number-of-courses>" courses
        Examples:
            | number-of-students | number-of-courses |
            | 10                 | 05                |
            | 02                 | 06                |
            | 20                 | 20                |

    @blocker
    Scenario Outline: Sync update student course packages for notification successfully
        Given school admin creates "<number-of-students>" students
        And school admin creates "<number-of-courses>" courses with "1" classes for each course
        When assigning course packages to existing students
        And waiting for all student course packages to be synced
        And sync create student courses and class members successfully with "<number-of-students>" students and "<number-of-courses>" courses
        And admin edit assigned course packages with start at "<start-at>" and end at "<end-at>"
        And waiting for all student course packages to be synced
        Then sync update student course packages successfully with start at "<start-at>" and end at "<end-at>"
        Examples:
            | number-of-students | number-of-courses | start-at                 | end-at                   |
            | 2                  | 5                 | 2022-03-16T05:06:20.000Z | 2022-06-06T15:03:20.000Z |
            | 3                  | 10                | 2022-03-17T05:06:20.000Z | 2022-06-07T15:03:20.000Z |
