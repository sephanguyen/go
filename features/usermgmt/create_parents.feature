@blocker
Feature: Create parents
    As a school staff
    I'm able to create new parents

    Scenario Outline: Create "1" new parents by staff granted "<role>"
        Given a signed in "staff granted role <role>"
        And new "1" parents data "full fields"
        When "staff granted role <role>" create "1" new parents
        Then new parents were created successfully

        Examples:
            | role           |
            | school admin   |
            | hq staff       |
            | centre lead    |
            | centre manager |
            | centre staff   |

    Scenario Outline: Create "<times>" new parents with full fields
        Given a signed in "staff granted role school admin"
        And new "<times>" parents data "full fields"
        And assign "<tag type>" tag to parents
        When "staff granted role school admin" create "<times>" new parents
        Then new parents were created successfully

        Examples:
            | times | tag type                      |
            | 1     | USER_TAG_TYPE_PARENT          |
            | 2     | USER_TAG_TYPE_PARENT_DISCOUNT |

    Scenario Outline: Create parent without non-required <fields>
        Given a signed in "staff granted role school admin"
        And new "1" parents data "without <fields>"
        When "staff granted role school admin" create "1" new parents
        Then new parents were created successfully

        Examples:
            | fields                        |
            | first name phonetic           |
            | last name phonetic            |
            | parent primary phone number   |
            | parent secondary phone number |
            | parent phone number           |
            | remark                        |
            | external_user_id              |

    Scenario Outline: Cannot create if parent data to create has empty or invalid <field>
        Given a signed in "staff granted role school admin"
        And parent data has empty or invalid "<field>"
        When "staff granted role school admin" create new parents
        Then "staff granted role school admin" cannot create that account
        And receives "InvalidArgument" status code
        Examples:
            | field                |
            | empty email          |
            | password             |
            | last name            |
            | first name           |
            | first name           |
            | last name            |
            | country code         |
            | relationship         |
            | parentPhoneNumber    |
            | studentID empty      |
            | studentID not exist  |
            | tag for only student |

    Scenario: cannot create if resource path is invalid
        Given a signed in "staff granted role school admin"
        And new parents data
        When "staff granted role school admin" create new parents with invalid resource path
        Then "staff granted role school admin" cannot create that account
        And receives "InvalidArgument" status code

    Scenario Outline: teacher, parent, student don't have permission to create new parents
        Given a signed in "staff granted role school admin"
        And new parents data
        When "<signed-in user>" create new parents
        Then returns "PermissionDenied" status code

        Examples:
            | signed-in user             |
            | staff granted role teacher |
            | parent                     |
            | student                    |

    Scenario: Create parents in different organizations but sharing the same email/phone number
        Given a signed in "staff granted role school admin"
        When "staff granted role school admin" in organization 1 create parent 1
        And "staff granted role school admin" in organization 2 create parent 2 with the same email as parent 1
        Then parent 1 will be created successfully and belonged to organization 1
        And parent 2 will be created successfully and belonged to organization 2

