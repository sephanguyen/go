// camel-k: property=file:../../../resources/application.properties

package com.manabie;

import org.apache.camel.PropertyInject;
import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.component.kafka.KafkaComponent;
import org.apache.camel.component.kafka.KafkaConstants;
import org.apache.camel.component.kafka.consumer.KafkaManualCommit;
import org.apache.camel.component.kafka.consumer.errorhandler.KafkaConsumerListener;

import com.manabie.libs.KafkaConsumerPredicate;
import com.manabie.libs.ManabieUnleash;

import io.fabric8.kubernetes.api.model.ExecAction;

import org.apache.camel.Exchange;
import org.apache.camel.Processor;

public class KafkaWithUnleash extends RouteBuilder {

    @PropertyInject("unleash.env")
    String env;

    @PropertyInject("unleash.org")
    String org;

    @PropertyInject("unleash.url")
    String url;

    @PropertyInject("unleash.token")
    String token;

    @PropertyInject("unleash.service")
    String service;

    @Override
    public void configure() throws Exception {
        getContext().getRegistry().bind("unleash", new KafkaConsumerPredicate());
        getContext().getRegistry().bind("KafkaConsumerListener", new KafkaConsumerListener());

        ManabieUnleash unleash = new ManabieUnleash(service, url, token, env, org);

        from("kafka:{{kafka.topic}}?brokers={{kafka.bootstrap-server}}"
                + "&maxPollRecords=10"
                + "&isolationLevel=read_committed"
                + "&consumersCount=1"
                // + "&seekTo=beginning"
                // + "&groupInstanceId=test"
                + "&allowManualCommit=true"
                + "&groupId=test")
                .pausable("KafkaConsumerListener",
                        o -> unleash.isEnabled("Architecture_BACKEND_MasterData_Course_TeachingMethod"))
                .process(new Processor() {
                    public void process(Exchange exchange) {
                        System.out.println("message committing" + exchange.getIn().getHeaders());
                        System.out.println("message committing" + exchange.getIn().getBody());
                        KafkaManualCommit manual = exchange.getIn().getHeader(KafkaConstants.MANUAL_COMMIT,
                                KafkaManualCommit.class);
                        manual.commit();
                    }
                })
                .log("info:${headers}-----${body}");
    }

}
