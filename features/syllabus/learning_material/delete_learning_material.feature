Feature: Delete learning material

    Background:
        Given <learning_material>a signed in "school admin"
        And <learning_material>a valid book content
        # make sure exisited at least 1 record for each type (LM)
        And some existing learning materials in an arbitrary topic of the book

    Scenario Outline: authenticate <role> delete learning material
        Given <learning_material>a signed in "<role>"
        When user deletes an arbitrary learning material
        Then <learning_material>returns "<msg>" status code
        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | school admin   | OK               |
            | hq staff       | OK               |
            | teacher        | PermissionDenied |
            | centre lead    | PermissionDenied |
            | centre manager | PermissionDenied |
            | teacher lead   | PermissionDenied |

    Scenario Outline: user tries to delete a learning material
        When user deletes the "<learning material>"
        And our system must delete the "<learning_material>" correctly
        And our system must delete the "<learning_material>" correctly
        Examples:
            | learning material  |
            | assignment         |
            | flashcard          |
            | learning_objective |
    #TODO: exam_lo
    #TODO: task_assignment

    Scenario: user tries to delete a learning material with wrong ID
        When user deletes the "assignment" with wrong ID
        Then <learning_material>returns "Internal" status code
#TODO: cover missing field to unit test
