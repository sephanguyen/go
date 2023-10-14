Feature: Check quiz correctness with config partial
  Background: admin create topic and learning objectives
    Given a signed in "school admin"
    When user upsert a topic
    Then returns "OK" status code
    When user create a learning objective
    Then returns "OK" status code

  Scenario Outline: student do multiple choice quiz test with partial config on
    Given a quiz test with partial config on and test case's data with "1,2" correct, "3,4,5" not correct
    When student choose option "<options>"
    Then returns "OK" status code
    And returns isCorrectAll: "<expect>"

    Examples:
      | options | expect |
      | 1,2     | true   |
      | 1,3     | true   |
      | 3,4     | false  |
      | 1       | true   |
      | 5       | false  |

  Scenario Outline: Student do multiple choice quiz test with partial config off
    Given a quiz test with partial config off and test case's data with "1,2" correct, "3,4,5" not correct
    When student choose option "<options>"
    Then returns "OK" status code
    And returns isCorrectAll: "<expect>"

    Examples:
      | options | expect |
      | 1,2     | true   |
      | 1,3     | false  |
      | 3,4     | false  |
      | 1       | false  |
      | 5       | false  |

  Scenario Outline: Student do fill in blank quiz test with partial config on
    Given a FIB quiz test with partial config on and correct answers "A,B,C"
    When student fill in text "<text>"
    Then returns "OK" status code
    And this is absolutely an "<status>" answer with isCorrectAll: "<expect>"

    Examples:
      | text  | expect |
      | a,b,c | true   |
      | a,b,C | true   |
      | a,B,C | true   |
      | A,B,C | true   |

  Scenario Outline: Student do fill in blank quiz test with no config
    Given a FIB quiz test with no config and correct answers "A,B,C"
    When student fill in text "<text>"
    Then returns "OK" status code
    And this is absolutely an "<status>" answer with isCorrectAll: "<expect>"

    Examples:
      | text  | expect |
      | a,b,c | true   |
      | a,b,C | true   |
      | a,B,C | true   |
      | A,B,C | true   |

  Scenario: Student do fill in blank quiz test with case sensitive config
    Given a FIB quiz test with case sensitive config and correct answers "A,B,C"
    When student fill in text "<text>"
    Then returns "OK" status code
    And this is absolutely an "<status>" answer with isCorrectAll: "<expect>"

    Examples:
      | text  | expect |
      | a,b,c | false  |
      | a,b,C | false  |
      | a,B,C | false  |
      | A,B,C | true   |

  Scenario: Student do fill in blank quiz test with case sensitive and partial config
    Given a FIB quiz test with case sensitive and partial config and correct answers "A,B,C"
    When student fill in text "<text>"
    Then returns "OK" status code
    And this is absolutely an "<status>" answer with isCorrectAll: "<expect>"

    Examples:
      | text  | expect |
      | a,b,c | false  |
      | a,b,C | true   |
      | a,B,C | true   |
      | A,B,C | true   |

  Scenario Outline: Student do fill in blank quiz test with japanese sentence and spaces and no config
    Given a FIB quiz test with no config and correct answers "す　みま　　　せ  ん、ゴク サン"
    When student fill in text "<text>"
    Then returns "OK" status code
    And this is absolutely an "<status>" answer with isCorrectAll: "<expect>"

    Examples:
      | text                             | expect |
      | す　みま　　　せ  ん、ゴク サン  | true   |
      | す みま せ ん、ゴク サン         | true   |
      | す			みま せ　　　ん、ｺﾞｸ ｻﾝ     | true   |
      | す　みま　 　　せ  ん、ｺﾞｸ ｻﾝ    | true   |
      | す 　みま　　　せ  ん、ゴク サン | true   |
      | す　みま	 		せ  ん、ゴク サン    | true   |

  Scenario Outline: Student do fill in blank quiz test with japanese sentence and spaces and no config
    Given a FIB quiz test with no config and correct answers "明智 光秀"
    When student fill in text "<text>"
    Then returns "OK" status code
    And this is absolutely an "<status>" answer with isCorrectAll: "<expect>"

    Examples:
      | text       | expect |
      | 明智　光秀 | true   |

  Scenario Outline: Student do fill in blank quiz test with japanese sentence and spaces and no config
    Given a FIB quiz test with no config and correct answers "In the morning"
    When student fill in text "<text>"
    Then returns "OK" status code
    And this is absolutely an "<status>" answer with isCorrectAll: "<expect>"

    Examples:
      | text             | expect |
      | In　the　morning | true   |