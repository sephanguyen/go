Feature: Interact with the course when applying RLS in our database

    @wip
    Scenario: School admin create a new course
        Given school admin logins CMS App with resource path is "<schooladminResourcePath>"
        And teacher logins Teacher App with resource path is "<teacherResourcePath>"
        And enable RLS on "courses" table
        When school admin create a new course
        Then teacher "<result>" the new course on Teacher App
        And disable RLS on "courses" table
        Examples:
            | schooladminResourcePath    | teacherResourcePath      | result
            | resourcePath 1             | resourcePath 1           | can see
            | resourcePath 2             | resourcePath 3           | cannnot see

