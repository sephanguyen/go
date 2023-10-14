// camel-k: property=file:../../../resources/application.properties

package com.manabie;

import java.util.UUID;

import org.apache.camel.builder.RouteBuilder;

public class KafkaProducer extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=6000")
                .setHeader("kafka.KEY", method(UUID.class, "randomUUID"))
                .setBody().simple("sent data")
                .to("kafka:{{kafka.topic}}?brokers={{kafka.bootstrap-server}}&additional-properties[transactional.id]=#bean:genUUID&additional-properties[enable.idempotence]=true&additional-properties[retries]=5")
                .routeId("FromKafka")
                .log("${body}");

    }

}
