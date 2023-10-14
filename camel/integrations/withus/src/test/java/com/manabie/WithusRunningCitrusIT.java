package com.manabie;

import static com.consol.citrus.actions.EchoAction.Builder.echo;

import com.consol.citrus.TestCaseRunner;
import com.consol.citrus.annotations.CitrusResource;
import com.consol.citrus.annotations.CitrusTest;
import com.consol.citrus.dsl.endpoint.CitrusEndpoints;
import com.consol.citrus.junit.jupiter.CitrusExtension;
import com.consol.citrus.kubernetes.client.KubernetesClient;

import static org.junit.Assert.assertTrue;

import java.util.Properties;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import static com.consol.citrus.kubernetes.actions.KubernetesExecuteAction.Builder.kubernetes;

/**
 *
 *
 * @author Unknown
 * @since 2023-06-28
 */
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
