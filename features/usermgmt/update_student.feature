@blocker
Feature: Update student
    As a school admin / school staff
    I need to be able to update a existing student

    Scenario Outline: Update a student account successfully
        Given student account data to update
        When "<signed-in user>" update student account
        Then student account is updated success
        And receives "OK" status code

        Examples:
            | signed-in user                  | has data              |
            | staff granted role school admin | has only student data |

    Scenario Outline: Update a student account with enrollment_status_str successfully
        Given student account data to update with enrollment status string "<enrollment-status string>"
        When "<signed-in user>" update student account
        Then student account is updated success
        And receives "OK" status code

        Examples:
            | signed-in user                  | enrollment-status string               |
            | staff granted role school admin | STUDENT_ENROLLMENT_STATUS_ENROLLED     |
            | staff granted role school admin | STUDENT_ENROLLMENT_STATUS_STRING_EMPTY |

    Scenario Outline: student, teacher, parent don't have permission to update student
        Given student account data to update
        When "<signed-in user>" update student account
        Then receives "<msg>" status code

        Examples:
            | signed-in user             | msg              |
            | staff granted role teacher | PermissionDenied |
            | student                    | PermissionDenied |
            | parent                     | PermissionDenied |

    Scenario Outline: Cannot update student does not exist
        Given student account data to update
        When "<signed-in user>" update student account does not exist
        Then "<signed-in user>" cannot update student account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | msg             |
            | staff granted role school admin | InvalidArgument |

    Scenario Outline: Cannot update student account without student data field: <requiredField>
        Given student account data to update
        When "<signed-in user>" updates student account with new student data missing "<requiredField>" field
        Then "<signed-in user>" cannot update student account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | requiredField    | msg             |
            | staff granted role school admin | name             | InvalidArgument |
            | staff granted role school admin | enrollmentStatus | InvalidArgument |
            | staff granted role school admin | email            | InvalidArgument |
            | staff granted role school admin | location_ids     | InvalidArgument |

    Scenario Outline: Update student account successfully without student data field: <field>
        Given student account data to update
        When "<signed-in user>" updates student account with new student data missing "<field>" field
        Then student account is updated success
        And receives "OK" status code

        Examples:
            | signed-in user                  | field             |
            | staff granted role school admin | studentExternalId |
            | staff granted role school admin | studentNote       |
            | staff granted role school admin | birthday          |
            | staff granted role school admin | gender            |

    Scenario Outline: Cannot update student that has invalid data
        Given student account data to update
        When "<signed-in user>" update student account that has "<invalid data>"
        Then "<signed-in user>" cannot update student account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | invalid data                      | msg             |
            | staff granted role school admin | unknown student enrollment status | InvalidArgument |

    Scenario Outline: Cannot update student that has exist email in our system
        Given student account data to update
        When "<signed-in user>" update student email that exist in our system
        Then "<signed-in user>" cannot update student account
        And receives "AlreadyExists" status code

        Examples:
            | signed-in user                  |
            | staff granted role school admin |

    Scenario Outline: Cannot update student that has invalid locations
        Given student account data to update has invalid locations "<invalid-location>"
        When "<signed-in user>" update student account
        Then "<signed-in user>" cannot update student account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | invalid-location      | msg             |
            | staff granted role school admin | empty                 | InvalidArgument |
            | staff granted role school admin | not found             | InvalidArgument |
            | staff granted role school admin | invalid resource_path | InvalidArgument |

    Scenario Outline: Update student that location under student_course
        Given student account data to update
        And assign course package with "<status>" to exist student
        When "<signed-in user>" update student account
        Then "<signed-in user>" user "<ability>" update student account
        And receives "<msg>" status code

        Examples:
            | signed-in user                  | status                     | ability | msg             |
            | staff granted role school admin | active                     | cannot  | InvalidArgument |
            | staff granted role school admin | inactive                   | can     | OK              |
            | staff granted role school admin | active with valid location | can     | OK              |

    Scenario Outline: Update a student account successfully with first name last name and phonetic name
        Given student account data to update with first name lastname and phonetic name
        When "<signed-in user>" update student account
        Then student account is updated success with first name last name and phonetic name
        And receives "OK" status code

        Examples:
            | signed-in user                  |
            | staff granted role school admin |

    Scenario Outline: Update a student with user tags
        Given existed student data with some "<tag-type>" tags
        And update student info with "<update-tag>" tags
        When "<signed-in user>" update student account
        Then student account is updated success with tags
        And receives "OK" status code

        Examples:
            | signed-in user                  | update-tag            | tag-type              |
            | staff granted role school admin | add more              | USER_TAG_TYPE_STUDENT |
            | staff granted role school admin | remove one            | USER_TAG_TYPE_STUDENT |
            | staff granted role school admin | remove one & add more | USER_TAG_TYPE_STUDENT |

    Scenario Outline: Update a student with invalid user tags
        Given existed student data with some "<tag-type>" tags
        And update student info with "<update-tag>" tags
        When "<signed-in user>" update student account
        Then receives "<msg>" status code

        Examples:
            | signed-in user                  | update-tag          | tag-type              | msg             |
            | staff granted role school admin | not found           | USER_TAG_TYPE_STUDENT | InvalidArgument |
            | staff granted role school admin | tag for only parent | USER_TAG_TYPE_STUDENT | InvalidArgument |

    Scenario Outline: Update a student account successfully
        Given student account data to update with parent info
        When "<signed-in user>" update student account
        Then student account is updated success
        And receives "OK" status code

        Examples:
            | signed-in user                  |
            | staff granted role school admin |