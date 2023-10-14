Feature: exam lo submission list

    Scenario Outline: exam lo submission list
        Given user create a set of exam lo submissions to database
        When user call function exam lo submission list
        And system returns correct list exam lo submissions
