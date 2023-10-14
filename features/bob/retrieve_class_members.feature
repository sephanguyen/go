
@quarantined
Feature: Retrieve class members
    Background:
        Given "disable" Unleash feature with feature name "Architecture_BACKEND_RetrieveClassMembers_Use_Mastermgmt_Repo"
    Scenario Outline: Teacher get class members
        Given some class members
        And a signed in user as a teacher
        When the teacher gets "<type group user>" class members
        Then our system returns class members correctly
    
    Examples:
        | type group user         |
        | teacher                 |
        | student                 | 
        | both teacher and student| 

  
