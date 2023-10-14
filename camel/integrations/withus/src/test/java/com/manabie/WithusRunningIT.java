package com.manabie;

import static com.consol.citrus.actions.EchoAction.Builder.echo;

import com.consol.citrus.annotations.CitrusTest;
import com.consol.citrus.junit.jupiter.CitrusExtension;

import io.fabric8.kubernetes.api.model.Pod;
import io.fabric8.kubernetes.api.model.PodList;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.Properties;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

/**
 *
 *
 * @author Unknown
 * @since 2023-06-28
 */
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
