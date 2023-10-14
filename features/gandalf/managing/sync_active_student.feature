Feature: sync active student

    Check tool migrate active student work correctly

    Scenario: sync start time and end time of student in course
        Given signed as "school admin" account
        And school admin creates an existed course
        When school admin creates student with course package
        Then yasuo returns "OK" status code
        And eureka store student course info

        # suppose for old data: after create student we dont store start date and end date of student course
        Given delete start date and end date of this student course
        And  signed as "school admin" account
        When run migration tool sync active student
        Then after sync active student eureka store correct student course info 
