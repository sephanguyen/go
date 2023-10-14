Feature: Upsert individual study plan
    Scenario Outline: authenticate insert allocate marker
        Given <allocate_marker>a signed in "<role>"
        Then admin insert allocate marker with "<number_submission>" submissions and first teacher has "<number_allocated_submission_1>" submissions,second teacher has "<number_allocated_submission_2>" submissions
        Then <allocate_marker>returns "<msg>" status code
        Examples:
            | role           | msg              | number_submission | number_allocated_submission_1 | number_allocated_submission_2 |
            | parent         | PermissionDenied | 10                | 2                             | 8                             |  
            | student        | PermissionDenied | 10                | 2                             | 8                             |  
            | teacher lead   | PermissionDenied | 10                | 2                             | 8                             |  
            | teacher        | PermissionDenied | 10                | 2                             | 8                             |  

    Scenario Outline: admin insert allocate marker
        Given <allocate_marker>a signed in "<role>"
        Then admin insert allocate marker with "<number_submission>" submissions and first teacher has "<number_allocated_submission_1>" submissions,second teacher has "<number_allocated_submission_2>" submissions
        Then <allocate_marker>returns "<msg>" status code
        And our system stores allocate marker correctly
        Examples:
            | role           | msg | number_submission | number_allocated_submission_1 | number_allocated_submission_2 |
            | school admin   | OK  | 10                | 2                             | 8                             |

    Scenario Outline: admin list allocate teacher
        Given "<number_teacher>" teachers access "<number_course>" courses by location
        Then admin insert allocate marker with "<number_submission>" submissions for first teacher
        Then admin lists allocate teacher
        And our system returns allocate teacher correctly

        Examples:
            | number_teacher | number_course | number_submission |
            | 2              | 1             | 5                 |
