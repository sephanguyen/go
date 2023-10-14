Feature: Export master data

    Scenario Outline: Export master data success
        Given data of "<master data type>" is existing
        When "<signed-in user>" export "<master data type>" data successfully
        Then receives "OK" status code
        And the "<master data type>" CSV has a correct content
        Examples:
            | signed-in user | master data type              |
            | school admin   | discount tag                  |
