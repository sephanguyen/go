@quarantined
Feature: reproduce case when chat missing message

  Background: default manabie resource path
    Given resource path of school "Manabie" is applied

  Scenario: all user in conversation subscribe to new stream with valid case
    When create a valid student conversation in db with a teacher and a student
    Then all member subscribe this conversation to chat and do not miss any message

  Scenario Outline: consistent hash user connection
    Given grpc metadata with key "x-chat-userhash" in context
    And a user subscribes stream using "<num-grpc>" grpc connections and "<num-grpc-web>" grpc web connections
    And all connections are routed to one node
    When spamming "<num msg>" into conversation
    Then all connections receive "<num msg>" msg in order
    Examples:
      | num msg | num-grpc | num-grpc-web |
      | 10      | 4        | 4            |
      | 20      | 5        | 5            |
      | 20      | 1        | 9            |
      | 20      | 9        | 1            |

  Scenario Outline: requests without hashkey are routed random
    Given a user subscribes stream using "<num-grpc>" grpc connections and "<num-grpc-web>" grpc web connections
    Then all connections are routed to multiple nodes
    Examples:
      | num-grpc | num-grpc-web |
      | 5        | 5            |
      | 10       | 5            |





