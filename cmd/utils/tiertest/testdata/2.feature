Feature: Test 2

    Scenario:0 
        Given "a"

    @minor
    Scenario: 1
        Given "a"

    @major
    Scenario: 2
        Given "<sample>"
        Examples:
            | sample |
            | a      |
            | b      |

    @critical
    Scenario: 3
        Given "b"

    @blocker
    Scenario: 4
        Given "<sample>"
        Examples:
            | sample |
            | a      |
            | b      |

    @weirdtag @blocker @critical
    Scenario: 5
        Given "<sample>"
        Examples:
            | sample |
            | a      |
            | b      |