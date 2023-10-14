### **Required:**
- Java > 17
- Maven
- Yaks
- Kamel 

### **Run code:**
```
./deployments/sk.bash
CAMEL_K_ENABLED=true ./deployments/sk.bash -f skaffold2.camel-k.yaml
make withus
```
### **Run Unit Test:**
```
make unit-test
```
### **Run Integration Test with Yaks:**
Pls copy file `service_credential.json` to folder `camel/integrations/withus/bdd` first
```
make install-testing-tool
make integration-test-yaks
```
Style:
```
Feature: Veirfy Withus integration
  Scenario: Verify Withus
    Given MANABIE can be extended!
    And variable body is "abcd"
    And Camel K integration property file application.properties
    And load Camel K integration Withus.java
    When Camel K integration withus is running
    Then Camel K integration withus should print json line: abcd
```
### **Run Integration Test with Citrus:**
Before we run integration test pls make sure we run all Integrations
```
integration-test-only
```
Style:

**Normal Junit5**
```
@ExtendWith(CitrusExtension.class)
@ExtendWith(EnvPropertiesResolver.class)
public class WithusRunningIT {

    @CitrusTest
    @Test
    public void testWithusLogSuccess(Properties env) {
        echo("test running");
        String integration = "withus";
        String namespace = "camel-k";

        KubernetesClient client = new DefaultKubernetesClient(env.getProperty("K8S_URL"));

        PodList pods = client
                .pods()
                .inNamespace(namespace)
                .withLabel("camel.apache.org/integration", integration)
                .list();

        assertFalse(pods.getItems().isEmpty());

        Pod pod = pods.getItems().get(0);

        String log = client.pods()
                .inNamespace(namespace)
                .withName(pod.getMetadata().getName()).getLog();
        echo("log: " + log);
        assertTrue(log.contains("json line:"));
    }
}
```
**Junit5 with support Citrus**
```
@ExtendWith(CitrusExtension.class)
@ExtendWith(EnvPropertiesResolver.class)
public class WithusRunningCitrusIT {

    private KubernetesClient GetClientConfig(Properties env) {
        return CitrusEndpoints
                .kubernetes()
                .client()
                .url(env.getProperty("K8S_URL"))
                .build();
    }

    @CitrusTest
    @Test
    public void testWithusRunning(@CitrusResource TestCaseRunner runner, Properties env) {
        runner.run(echo("test running"));
        runner.run(kubernetes()
                .client(GetClientConfig(env))
                .pods()
                .list()
                .name("withus")
                .namespace("camel-k")
                .validate((result, context) -> {
                    echo("result.getResult().getItems()" +
                            result.getResult().getItems().toString());
                    assertTrue(!result.getResult().getItems().isEmpty());
                }));
    }

}

```
### **Documents:**
How to write unit test: https://camel.apache.org/components/3.20.x/others/test-junit5.html
How to mock unit test: https://camel.apache.org/components/3.20.x/mock-component.html
How to write bdd test with yaks: https://citrusframework.org/yaks/reference/html/index.html#conditional-scripts
Camel K with Yaks: https://camel.apache.org/blog/2023/01/camel-k-yaks/
Camel k modeline: https://camel.apache.org/camel-k/1.12.x/cli/modeline.html
Camel K runtime property: https://camel.apache.org/camel-k/1.12.x/configuration/runtime-properties.html
### **Maven useful command*:*
Install dependency:
```
mvn clean install -U      
```
Run test/compile/package:
```
mvn package
```
Build page without test:
```
mvn package -Dmaven.test.skip
```
Compile:
```
mvn compile
```
Run test only:
```
mvn test
```