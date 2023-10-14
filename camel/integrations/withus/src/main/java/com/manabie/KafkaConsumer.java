
// camel-k: dependency=camel:kafka
// camel-k: property=file:../../../resources/application.properties

package com.manabie;

import org.apache.camel.builder.RouteBuilder;

public class KafkaConsumer extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        from("kafka:{{kafka.topic}}?brokers={{kafka.bootstrap-server}}"
                + "&maxPollRecords=10"
                + "&isolationLevel=read_committed"
                + "&consumersCount=1"
                + "&seekTo=beginning"
                + "&groupId=test")
                .log("info:${headers}-----${body}");

    }
}
