Feature: Health check the unleash and unleash proxy
  @critical
  Scenario: Health check the unleash
    Given the request to check the unleash health
    When send request to check health
    Then unleash must return healthy status

  @critical
  Scenario: Health check the unleash-proxy
    Given the request to check the unleash-proxy health
    When send request to check health
    Then unleash-proxy must return health status
