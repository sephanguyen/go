Feature: Calculate highest score

    Scenario Outline: Calculate highest score
        Given user create a study plan of "<type>" to database
        When user create a shuffle quiz set with some of quizzes
        And user submitted with some answers
        Then user calculate highest submission score correctly
        Examples:
            | type      |
            | LO        |
            | ExamLO    |
            | FlashCard |

    Scenario: Calculate highest score with reterned status
    Given user create study plan of exam lo to database
    When user create a shuffle quiz set with some of quizzes
    And user submitted with some answers
    Then user calculate highest exam lo submission score correctly

    Scenario: Get highest score of exam lo
    Given user create study plan of exam lo to database
    When user create a shuffle quiz set with some of quizzes
    And user submitted with some answers
    And valid completeness exam lo in database
    Then user get highest exam lo score correctly

