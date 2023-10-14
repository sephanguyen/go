Feature: Global merge block
    Scenario: Pr created is block if repo is blocked globally
        Given a repo "test-ci" of owner "manabie-com"
        When workflow "block" is called to repo
        Then repo has merge status is "block"
        When workflow "unblock" is called to repo
        Then repo has merge status is "unblock"

