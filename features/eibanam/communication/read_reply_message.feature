@cms @teacher @learner @parent
@communication

Feature: Read and reply message

    Background:
        Given "school admin" logins CMS
        And "school admin" has created student with parent info
        And "teacher" logins Teacher App
        And "teacher" has joined student chat group
        And "teacher" has joined parent chat group
        And "student" logins Learner App
        And "parent" logins Learner App
        And "student" is at the conversation screen
        And "parent" is at the conversation screen

    Scenario Outline: Teacher sees status that <userAccount> does not read the message in Teacher App
        Given teacher sends "<messageType>" message to "<userAccount>"
        When "<userAccount>" does not read message
        Then teacher does not see "Read" status next to message in conversation on Teacher App
        And teacher sees "Replied" icon next to chat group in Messages list on Teacher App
        And "<userAccount>" sees chat group with unread message is showed on top on Learner App
        And "<userAccount>" sees "Unread" icon next to chat group in Messages list on Learner App
        Examples:
            | messageType             | userAccount |
            | 1 of [text, image, pdf] | student     |
            | 1 of [text, image, pdf] | parent      |

    Scenario Outline: Teacher sees status that <userAccount> read the message in Teacher App
        Given teacher sends "<messageType>" message to "<userAccount>"
        When "<userAccount>" reads the message
        Then teacher sees "Read" status next to message on Teacher App
        And teacher sees "Replied" icon next to chat group in Messages list on Teacher App
        And "<userAccount>" sees "Unread" icon next to chat group disappeared in Messages list on Learner App
        Examples:
            | messageType             | userAccount |
            | 1 of [text, image, pdf] | student     |
            | 1 of [text, image, pdf] | parent      |

    Scenario Outline: Teacher does not reads reply from <userAccount>
        Given teacher sends "<messageType>" message to "<userAccount>"
        And "<userAccount>" replies to teacher
        When teacher does not read replies
        Then teacher does not see "Replied" icon next to chat group in Messages list on Teacher App
        And teacher sees "Unread" icon next to chat group in Messages list on Teacher App
        And teacher sees chat group with unread message shown on top in Messages list on Teacher App
        And "<userAccount>" does not see "Read" status next to message on Learner App
        Examples:
            | messageType             | userAccount |
            | 1 of [text, image, pdf] | student     |
            | 1 of [text, image, pdf] | parent      |

    Scenario Outline: Teacher reads reply from <userAccount>
        Given teacher sends "<messageType>" message to "<userAccount>"
        And "<userAccount>" replies to teacher
        When teacher read replies
        Then teacher does not see "Replied" icon next to chat group in Messages list on Teacher App
        And teacher sees "Unread" icon next to chat group disappeared in Messages list on Teacher App
        And "<userAccount>" does not see "Read" status next to message on Learner App
        Examples:
            | messageType             | userAccount |
            | 1 of [text, image, pdf] | student     |
            | 1 of [text, image, pdf] | parent      |