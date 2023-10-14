package com.manabie;

import org.apache.camel.builder.RouteBuilder;

public class AuthExampleCamelK extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .to("auth://userCredential?googleApiKey=example_key&authServiceAddress=localhost&authServicePort=5050&tenantId=exampleTenantId&username=exampleUsername&password=123456")
                .log("header: ${header.Authorization}"); // Auth component exchange user credential to manabie token then attach it to header
    }

}
