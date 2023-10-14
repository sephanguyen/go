Feature: Retrieve class members with unleash enable
    Scenario: Teacher get class members
        Given some class members
        And a signed in teacher
        And "enable" Unleash feature with feature name "Architecture_BACKEND_RetrieveClassMembers_Use_Mastermgmt_Repo"
        When the teacher gets "both teacher and student" class members
        Then our system returns class members correctly
