@cms @learner @parent
@communication
@ignore

Feature: Send notification for selected recipients
    Background:
        Given "school admin" logins CMS
        And "school admin" has created student "student S1" with grade and parent "parent P1", parent "parent P2" info
        And "school admin" has created student "student S2" with grade and parent "parent P3" info
        And "school admin" has created 2 courses "Course C1" and "Course C2"
        And "school admin" has added course "Course C1" for student "student S1"
        And "school admin" has added course "Course C2" for student "student S2"
        And "student S1, student S2" login Learner App
        And "parent P1, parent P2, parent P3" login Learner App
        And "school admin" is at Notification page
    
    Scenario Outline: Send notification to specific user in <type> list
        Given school admin has created notification
        When school admin sends a notification to the "<type>" list in "<course>", "<grade>", "<userEmail>"
        Then "<type>" who relates to "<userEmail>" receive the notification
        Examples:
            | type                                       | course | grade | userEmail                                                                              |
            | 1 of [student and parent, student, parent] | empty  | empty | 1 of [student S1's email, student S2's email, student S1's email & student S2's email] |
    
    Scenario Outline: Send a notification to the <type> list in <course>, <grade>, <userEmail>
    Given school admin has created notification
    When school admin sends a notification to the "<type>" list in "<course>", "<grade>", "<userEmail>"
    Then matching recipients receive the notification
    Examples:
        | type                                       | course                                                         | grade                                                                    | userEmail                                            |
        | 1 of [student and parent, student, parent] | 1 of [All courses, Course C1, Course C2]                       | 1 of [All grade, student S1's grade]                                     | 1 of [empty, student S1's email, student S2's email] |
        | 1 of [student and parent, student, parent] | 1 of [All courses, Course C1, Course C2]                       | 1 of [All grade, student S1's grade]                                     | 1 of [student S1's email & student S2's email]       |
        | 1 of [student and parent, student, parent] | 1 of [All courses, Course C1, Course C2]                       | 1 of [student S1's grade & student S2's grade, All & student S2's grade] | 1 of [student S1's email & student S2's email]       |
        | 1 of [student and parent, student, parent] | 1 of [All courses, Course C1, Course C2]                       | 1 of [student S1's grade & student S2's grade, All & student S2's grade] | 1 of [empty, student S1's email, student S2's email] |
        | 1 of [student and parent, student, parent] | 1 of [Course C1 & Course C2, All & Course C1, All & Course C2] | 1 of [student S1's grade & student S2's grade, All & student S2's grade] | 1 of [empty, student S1's email, student S2's email] |
        | 1 of [student and parent, student, parent] | 1 of [Course C1 & Course C2, All & Course C1, All & Course C2] | 1 of [student S1's grade & student S2's grade, All & student S2's grade] | 1 of [student S1's email & student S2's email]       |
        | 1 of [student and parent, student, parent] | 1 of [Course C1 & Course C2, All & Course C1, All & Course C2] | 1 of [All grade, student S1's grade]                                     | 1 of [student S1's email & student S2's email]       |
        | 1 of [student and parent, student, parent] | 1 of [Course C1 & Course C2, All & Course C1, All & Course C2] | 1 of [All grade, student S1's grade]                                     | 1 of [empty, student S1's email, student S2's email] |