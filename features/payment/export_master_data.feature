Feature: Export master data

  Scenario Outline: Export master data success
    Given data of "<master data type>" is existing
    When "<signed-in user>" export "<master data type>" data successfully
    Then receives "OK" status code
    And the "<master data type>" CSV has a correct content
    Examples:
      | signed-in user | master data type              |
      | school admin   | accounting category           |
      | school admin   | billing ratio                 |
      | school admin   | billing schedule              |
      | school admin   | billing schedule period       |
      | school admin   | discount                      |
      | school admin   | fee                           |
      | school admin   | leaving reason                |
      | school admin   | material                      |
      | school admin   | tax                           |
      | school admin   | product setting               |
      | school admin   | product price                 |
      | school admin   | product discount              |
      | school admin   | product location              |
      | school admin   | product grade                 |
      | school admin   | product accounting category   |
      | school admin   | package quantity type mapping |
      | school admin   | package course                |
      | school admin   | package                       |
      | school admin   | notification date             |
