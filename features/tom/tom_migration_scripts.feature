Feature: Sync scripts for chat

    @wip @disabled
    Scenario: Delete all conversations data in Tom and resync
        Given a new school is created with location default "default"
        Given accounts and chats of "<numstudents>" students are created with location "default", each has 2 parents
        When force remove all support conversations of this school
        And run "YasuoSyncUserConversation Create" script
        Then number of student chats created equal "<numstudents>"
        And number of parent membership created equal "<numparents>"
        And total conversations in this school is "<num total>"
        Examples:
            | numstudents | numparents | num total |
            | 3           | 6          | 6         |

    @wip @disabled
    Scenario: YasuoSyncUserConversation Updating locations
        Given a new school is created with location default "default"
        And locations "loc 1" children of location "default" are created
        Given accounts and chats of "<numstudents>" students are created with location "loc 1", each has 1 parents
        And remove all location of conversations in current school
        And run "YasuoSyncUserConversation Update" script
        Then there are "<num total>" support chats with exact locations "loc 1"
        Examples:
            | numstudents | numparents | num total |
            | 3           | 6          | 6         |

    @wip @disabled
    Scenario Outline: Delete all conversations data in Tom and resync
        Given a new school is created
        Given accounts of "<numstudents>" students are created
        And "<num live lesson>" live lesson chats are created including all of the students
        And force remove all lesson conversations
        When run "BobSyncLessonConversation" script
        Then number of lesson chats created equal "<num live lesson>"
        And number of lesson chats updated equal "0"
        And students are added to lesson chats after sync
        Examples:
            | numstudents | num live lesson |
            | 3           | 6               |

    @wip @disabled
    Scenario Outline: Tom has associated conversations for all user, using sync command does not do anything
        Given a new school is created
        Given accounts of "<numstudents>" students are created
        And "<num live lesson>" live lesson chats are created including all of the students
        When run "BobSyncLessonConversation" script
        Then number of lesson chats created equal "0"
        And number of lesson chats updated equal "0"
        And students are added to lesson chats after sync
        Examples:
            | numstudents | num live lesson |
            | 3           | 6               |

    @wip @disabled
    Scenario Outline: Partial sync non exisiting lesson chat, update out of synced lesson chat
        Given a new school is created
        Given accounts of "<numstudents>" students are created
        And "<num live lesson>" live lesson chats are created including all of the students
        And force remove "<removed chats>" lesson chats
        And force remove all students from "<to be updated lesson chat>" remaining lesson chats
        When run "BobSyncLessonConversation" script
        Then number of lesson chats created equal "<removed chats>"
        And number of lesson chats updated equal "<to be updated lesson chat>"
        And students are added to lesson chats after sync
        Examples:
            | numstudents | num live lesson | removed chats | to be updated lesson chat |
            | 5           | 7               | 3             | 3                         |

    Scenario Outline: Delete all conversations data in Tom and resync
        Given a new school is created with location default "default"
        And accounts and chats of "<numstudents>" students are created, each has 1 parents
        And another school "school 2" is created with "2" conversations on elasticsearch
        And wait for "<total chats>" chats of those schools created on elasticsearch
        # need this to ensure after deleting data on elasticsearch, nats handler do not recreate them
        # And wait for elasticsearch consumers to ack all msg from stream chat
        And delete chat data on elasticsearch for current school and school "school 2"
        When run "SynConversationDocument" script
        Then listing chat on elasticsearch with "current school" returns "<current school chats>" items
        # ensure other school data is not affected
        And listing chat on elasticsearch with "school 2" returns "0" items
        Examples:
            | numstudents | current school chats | total chats |
            | 3           | 6                    | 8           |

    Scenario Outline: Delete all conversations data in Tom and resync
        Given a new school is created with location default "default"
        And accounts and chats of "<numstudents>" students are created, each has 1 parents
        And another school "school 2" is created with "2" conversations on elasticsearch
        And wait for "<total chats>" chats of those schools created on elasticsearch
        # need this to ensure after deleting data on elasticsearch, nats handler do not recreate them
        # And wait for elasticsearch consumers to ack all msg from stream chat
        And delete chat data on elasticsearch for current school and school "school 2"
        When run "SynConversationDocument" script
        Then listing chat on elasticsearch with "current school" returns "<current school chats>" items
        # ensure other school data is not affected
        And listing chat on elasticsearch with "school 2" returns "0" items
        Examples:
            | numstudents | numparents | current school chats | total chats |
            | 3           | 6          | 6                    | 8           |

    # Because we enabled RLS for Tom, and this sync script only work when rls is disabled
    @wip @disabled
    Scenario: sync user device token resource path
        Given a new school is created
        Given accounts and chats of "<num students>" students are created, each has 2 parents
        # current bob does not store resource path, fake this behaviour
        And those users have resource path in Bob DB
        And those users have device token in Tom DB without resource path
        When run "BobSyncUserTokenResourcePath" script
        Then db has "<num users>" user device token with resource path of this school
        Examples:
            | num students | num users |
            | 3            | 9         |

    # Because we enabled RLS for Tom, and this sync script only work when rls is disabled
    @wip @disabled
    Scenario Outline: Sync lesson resource path
        Given a new school is created
        Given a ctx with resource_path of current school
        Given accounts of "<numstudents>" students are created
        And "<num live lesson>" live lesson chats are created including all of the students
        And those lesson chats have empty resource path
        When run "BobSyncLessonTokenResourcePath" script
        Then all lesson conversation have correct resource path
        Examples:
            | numstudents | num live lesson |
            | 3           | 6               |
            | 4           | 8               |
