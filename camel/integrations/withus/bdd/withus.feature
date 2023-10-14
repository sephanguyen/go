Feature: Veirfy Withus integration

  # Background:
  #   Given URL: http://0.0.0.0:8088
  #   Given HTTP server listening on port 8088
  #   Given HTTP request fork mode is enabled
  #   Given load HTTP response body service_credential.json
  #   And HTTP response body: ["Hello", "Hola", "Hi"]
  #   Given create HTTP server "sampleHttpServer"
  #   Given Kubernetes namespace camel-k
  #   Given Kubernetes service "hello-service"
  #   Given Kubernetes service port 8080
  #   Given create Kubernetes service


  Scenario: Verify Withus
    Given MANABIE can be extended!
    And variable body is "abcd"
    And Camel K integration property file application.properties
    And load Camel K integration Withus.java
    When Camel K integration withus is running
    Then Camel K integration withus should print json line: abcd

