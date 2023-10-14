Feature: List learning material

    Background:
        Given <learning_material>a signed in "school admin"
        And <learning_material>a valid book content
        And some existing learning materials in an arbitrary topic of the book

    Scenario Outline: authenticate when list learning material
        Given <learning_material>a signed in "<role>"
        When user send list arbitrary learning material request
        Then <learning_material>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | admin          | OK               |
            | teacher        | OK               |
            | student        | OK               |
            | hq staff       | OK               |
            | center lead    | OK               |
            | center manager | OK               |
            | center staff   | OK               |
            | lead teacher   | OK               |

    Scenario Outline: list learning material 
        When user send list learning material "<type>" request
        Then <learning_material>returns "OK" status code
        And our system must return learning material "<type>" correctly

        Examples: 
            | type                  |
            | assignment            |
            | exam_lo               |
            | flashcard             |
            | learning_objective    |
            | task_assignment       |