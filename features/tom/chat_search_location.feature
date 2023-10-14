Feature: List conversation by locations

  Background: Students chat and parents chat are created
    Given a new school is created with location "default"
    And a valid "school admin" token
    And locations "loc 1,loc 2" children of location "default" are created
    And a teacher account in db

  @throttle
  Scenario Outline: Location filter
    Given "english" language is used
    And student chat "s1" and parent chat "p1" are created with locations "loc 1"
    And student chat "s2" and parent chat "p2" are created with locations "<s2,p2 locations>"
    And filter chats with location "<filter location>"
    Then chats with ids "<returned chat>" are returned

    Examples:
      | s2,p2 locations | filter location | returned chat |
      | loc 2           | default         | s1,p1,s2,p2   |
      | loc 1,loc 2     | loc 1           | s1,p1,s2,p2   |
      | loc 1,loc 2     | loc 2           | s2,p2         |
      | loc 2           | loc 2           | s2,p2         |

  @throttle
  Scenario Outline: Location filter after user profile update
    Given "english" language is used
    And student chat "s1" and parent chat "p1" are created with locations "loc 1"
    And student chat "s2" and parent chat "p2" are created with locations "default"
    And usermgmt send event upsert user profile for student of chat "s2" with locations "loc 2"
    And filter chats with location "<filter location>"
    Then chats with ids "<returned chat>" are returned

    Examples:
      | filter location | returned chat |
      | loc 2           | s2,p2         |
      | loc 1           | s1,p1         |
      | loc 1,loc 2     | s1,p1,s2,p2   |

  Scenario Outline: Location filter for parent of multiple children
    Given student chat "s1" and parent chat "p1" are created with locations "loc 1"
    And student chat "s2" parent chat "p2" same parent with chat "p1" are created with locations "loc 2"
    When filter chats with location "<filter location>"
    Then chats with ids "<returned chat>" are returned

    Examples:
      | filter location | returned chat |
      | loc 2           | s2,p2         |
      | default         | s1,p1,s2,p2   |

  Scenario Outline: Location filter with course
    Given "english" language is used
    And student chat "s1" and parent chat "p1" are created with locations "loc 1"
    And student chat "s2" and parent chat "p2" are created with locations "<s2,p2 locations>"
    And mappings between student and course "<student course mappings>"
    When filter chats with location "<filter location>"
    And filter chats with filter combination "<chat filters>"
    Then chats with ids "<returned chat>" are returned

    Examples:
      | s2,p2 locations | filter location | returned chat | student course mappings   | chat filters               |
      | loc 2           | default         | p1,p2         | s1-c1-loc 1,s2-c1-loc 2   | contact Parent,courses c1  |
      | loc 1,loc 2     | loc 1           | s1            | s1-c1-loc 1,s2-c2-loc 2   | contact Student,courses c1 |
      | loc 1,loc 2     | loc 2           | s2,p2         | s1-c1-loc 1,s2-c2-loc 2   | contact All,courses c1-c2  |
      | loc 2           | loc 1,loc 2     | s1,p1,s2,p2   | s1-c1-loc 1,s2-c2-loc 2   | contact All,courses c1-c2  |
      | default         | loc 1           | s1,p1         | s1-c1-loc 1,s2-c2-default | courses c1-c2              |

  Scenario Outline: Chat thread launching by locations & user type
    Given "english" language is used
    And location configurations conversation value "<config value>" existed on DB
    And student chat "s1" and parent chat "p1" are created with locations "loc 1"
    And student chat "s2" and parent chat "p2" are created with locations "loc 2"
    When filter chats with location "<filter location>"
    And filter chats with type contact type "All" only
    Then chats with ids "<returned chat>" are returned

    Examples:
      | filter location | returned chat | config value |
      | loc 1,loc 2     |               | false        |
      | loc 1,loc 2     | s1,s2,p1,p2   | true         |

  Scenario Outline: Update location config for chat thread launching
    Given "english" language is used
    And location configurations conversation value "true" existed on DB
    And student chat "s1" and parent chat "p1" are created with locations "loc 1"
    And student chat "s2" and parent chat "p2" are created with locations "loc 2"
    When "<exclude type>" conversation is disabled in location configurations table with location "default"
    And filter chats with location "<filter location>"
    Then chats with ids "<returned chat>" are returned

    Examples:
      | filter location | exclude type | returned chat |
      | loc 1,loc 2     | student      | p1,p2         |
      | loc 1,loc 2     | parent       | s1,s2         |
