@blocker
Feature: Create student
    As a school staff
    I need to be able to create a new student

    Scenario Outline: Create a student with only student info
        Given student data with some "<tag-type>" tags
        When "<signed-in user>" create new student account
        Then new student account created success with student info
        And receives "OK" status code

        Examples:
            | signed-in user                    | tag-type                       |
            | staff granted role school admin   | USER_TAG_TYPE_STUDENT          |
            | staff granted role hq staff       | USER_TAG_TYPE_STUDENT          |
            | staff granted role centre lead    | USER_TAG_TYPE_STUDENT          |
            | staff granted role centre manager | USER_TAG_TYPE_STUDENT          |
            | staff granted role centre staff   | USER_TAG_TYPE_STUDENT          |
            | staff granted role school admin   | USER_TAG_TYPE_STUDENT_DISCOUNT |
            | staff granted role hq staff       | USER_TAG_TYPE_STUDENT_DISCOUNT |
            | staff granted role centre lead    | USER_TAG_TYPE_STUDENT_DISCOUNT |
            | staff granted role centre manager | USER_TAG_TYPE_STUDENT_DISCOUNT |
            | staff granted role centre staff   | USER_TAG_TYPE_STUDENT_DISCOUNT |

    Scenario Outline: Create a student with only student info
        Given only student info
        When "<signed-in user>" create new student account
        Then new student account created success with student info
        And receives "OK" status code

        Examples:
            | signed-in user                    |
            | staff granted role school admin   |
            | staff granted role hq staff       |
            | staff granted role centre lead    |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario Outline: Create a student only student info with enrollment status string
        Given only student info with enrollment status string "<enrollment-status string>"
        When "<signed-in user>" create new student account
        Then new student account created success with student info
        And receives "OK" status code

        Examples:
            | signed-in user                  | enrollment-status string               |
            | staff granted role school admin | STUDENT_ENROLLMENT_STATUS_ENROLLED     |
            | staff granted role school admin | STUDENT_ENROLLMENT_STATUS_STRING_EMPTY |

    Scenario Outline: Cannot create student account with unknown student enrollment status
        Given student data with unknown student enrollment status
        When "<signed-in user>" create new student account
        Then "<signed-in user>" cannot create that account

        Examples:
            | signed-in user                  |
            | staff granted role school admin |

    Scenario Outline: Cannot create student without <requiredField>
        Given student data missing "<requiredField>"
        When "<signed-in user>" create new student account
        Then "<signed-in user>" cannot create that account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | requiredField    | msg             |
            | staff granted role school admin | username         | InvalidArgument |
            | staff granted role school admin | password         | InvalidArgument |
            | staff granted role school admin | name             | InvalidArgument |
            | staff granted role school admin | enrollmentStatus | InvalidArgument |
            | staff granted role school admin | location_ids     | InvalidArgument |

    Scenario Outline: Create student successfully without <field>
        Given student data missing "<field>"
        When "<signed-in user>" create new student account
        Then new student account created success with student info
        And receives "OK" status code

        Examples:
            | signed-in user                  | field             |
            | staff granted role school admin | studentExternalId |
            | staff granted role school admin | studentNote       |
            | staff granted role school admin | birthday          |
            | staff granted role school admin | gender            |

    Scenario Outline: Cannot create student account with invalid resource path
        Given only student info
        When "<signed-in user>" create new student account with invalid resource path
        Then "<signed-in user>" cannot create that account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | msg             |
            | staff granted role school admin | InvalidArgument |

    Scenario Outline: Cannot create student with invalid locations
        Given student info with invalid locations "<invalid-location>"
        When "<signed-in user>" create new student account
        Then "<signed-in user>" cannot create that account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | invalid-location      | msg             |
            | staff granted role school admin | empty                 | InvalidArgument |
            | staff granted role school admin | not found             | InvalidArgument |
            | staff granted role school admin | invalid resource_path | InvalidArgument |

    Scenario Outline: Cannot create student with invalid tags
        Given student info with invalid tags "<invalid-tag>"
        When "<signed-in user>" create new student account
        Then "<signed-in user>" cannot create that account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | invalid-tag           | msg             |
            | staff granted role school admin | not found             | InvalidArgument |
            | staff granted role school admin | invalid resource_path | InvalidArgument |
            | staff granted role school admin | tag for only parent   | InvalidArgument |


    Scenario Outline: teacher, student, parent don't have permission to create new student
        Given only student info
        When "<signed-in user>" create new student account
        Then returns "<msg>" status code

        Examples:
            | signed-in user             | msg              |
            | student                    | PermissionDenied |
            | parent                     | PermissionDenied |
            | staff granted role teacher | PermissionDenied |

    Scenario Outline: Create a student with first name and last name and phonetic name
        Given only student info with first name last name and phonetic name
        When "<signed-in user>" create new student account
        Then new student account created success with student info and first name, last name, phonetic name
        And receives "OK" status code

        Examples:
            | signed-in user                  |
            | staff granted role school admin |

    Scenario Outline: Create students in different organizations but sharing the same email/phone number
        When "<signed-in user>" in organization 1 create user 1
        And "<signed-in user>" in organization 2 create user 2 with the same "<user attribute>" as user 1
        Then user 1 will be created successfully and belonged to organization 1
        And user 2 will be created successfully and belonged to organization 2

        Examples:
            | signed-in user                  | user attribute |
            | staff granted role school admin | email          |
            | staff granted role school admin | phone number   |
