@zeus
Feature: JetStream authorization

  Scenario Outline: publish a message with subject that he does not have permission (and have permission) to publish message
    Given A user with username is "<User>" and password is "<PassWord>"
    When "<User>" publishes a message with subject "<Subject Name>"
    Then "<User>" publishes this message "<Publish Status>"

    Examples:
      | User   | PassWord | Subject Name        | Publish Status |
      | Bob    | m@n@bi3  | ActivityLog.Created | successfully   |
      | Tom    | m@n@bi3  | ActivityLog.Created | successfully   |
      | Eureka | m@n@bi3  | ActivityLog.Created | successfully   |
      | Fatima | m@n@bi3  | ActivityLog.Created | successfully   |
      | Yasuo  | m@n@bi3  | ActivityLog.Created | successfully   |
      | Shamir | m@n@bi3  | ActivityLog.Created | successfully   |
      | Bob    | m@n@bi3  | SomeThing.Created   | failed         |
      | Tom    | m@n@bi3  | SomeThing.Created   | failed         |
      | Yasuo  | m@n@bi3  | SomeThing.Created   | failed         |
      | Eureka | m@n@bi3  | SomeThing.Created   | failed         |
      | Fatima | m@n@bi3  | SomeThing.Created   | failed         |
      | Shamir | m@n@bi3  | SomeThing.Created   | failed         |
      | Zeus   | m@n@bi3  | SomeThing.Created   | failed         |
      | Zeus   | m@n@bi3  | ActivityLog.Created | failed         |


  Scenario Outline: subscribe a message with subject that he does not have permission (and have permission) to subscribe
    Given A user with username is "<User>" and password is "<Password>"
    When "<User>" subscribes a message with subject "<Subject Name>"
    Then "<User>" subscribes this message "<Subscribe Status>"

    Examples:
      | User   | Password | Subject Name        | Subscribe Status |
      | Zeus   | m@n@bi3  | ActivityLog.Created | successfully     |
      | Bob    | m@n@bi3  | ActivityLog.Created | failed           |
      | Tom    | m@n@bi3  | ActivityLog.Created | failed           |
      | Yasuo  | m@n@bi3  | ActivityLog.Created | failed           |
      | Fatima | m@n@bi3  | ActivityLog.Created | failed           |
      | Eureka | m@n@bi3  | ActivityLog.Created | failed           |
      | Shamir | m@n@bi3  | ActivityLog.Created | failed           |