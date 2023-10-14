
// camel-k: resource=file:service_credential.json

import org.apache.camel.builder.RouteBuilder;

public class CrdSvc extends RouteBuilder {
    @Override
    public void configure() throws Exception {
        log.info("About to start route: http Server -> Log ");
        from("rest:get:crd")
                .transform().simple("resource:classpath:service_credential.json");

    }
}