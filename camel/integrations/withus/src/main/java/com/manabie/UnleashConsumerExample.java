// camel-k: property=file:../../../resources/application.properties

package com.manabie;

import org.apache.camel.builder.RouteBuilder;

public class UnleashConsumerExample extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        from("unleash://Architecture_BACKEND_MasterData_Course_TeachingMethod?env={{unleash.env}}&org={{unleash.org}}&url={{unleash.url}}&apiToken={{unleash.token}}&serviceName={{unleash.service}}")
                .choice()
                .when(simple("${header.UNLEASH_ENABLED} == true"))
                .to("log: flag is enabled")
                .when(simple("${header.UNLEASH_ENABLED} == false"))
                .to("log: flag is not enabled");
    }

}
