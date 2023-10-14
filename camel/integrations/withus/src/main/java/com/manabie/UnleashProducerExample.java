// camel-k: property=file:../../../resources/application.properties

package com.manabie;

import org.apache.camel.builder.RouteBuilder;

public class UnleashProducerExample extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=6000")
                .setBody(simple("test"))
                .to("unleash://Architecture_BACKEND_MasterData_Course_TeachingMethod?env={{unleash.env}}&org={{unleash.org}}&url={{unleash.url}}&apiToken={{unleash.token}}&serviceName={{unleash.service}}")
                .choice().when(simple("${header.UNLEASH_ENABLED} == true"))
                .log("sent: ${body}")
                .when(simple("${header.UNLEASH_ENABLED} == false"))
                .log("send to empty");
    }

}
