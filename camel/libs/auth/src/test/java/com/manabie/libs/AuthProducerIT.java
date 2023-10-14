package com.manabie.libs;

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.component.mock.MockEndpoint;
import org.apache.camel.test.junit4.CamelTestSupport;
import org.junit.Test;

import java.net.URI;
import java.util.Objects;

public class AuthProducerIT extends CamelTestSupport {

    // remember to expose auth service before running IT
    // kubectl -n local-manabie-backend  port-forward service/bob 5050:5050

    @Test
    public void testAuthProcedure() throws Exception {
        MockEndpoint mockEndpoint = getMockEndpoint("mock:result");
        mockEndpoint.expectedMessageCount(1);
        mockEndpoint.expectedBodiesReceived("Hello World");

        context.createProducerTemplate().sendBody("direct:test", "Hello World");

        String token =  mockEndpoint.getExchanges().get(0).getIn().getHeader("Authorization").toString();
        if (token.isEmpty()) {
            throw new Exception("expect header has authorization token");
        }
        String[] values = token.split(" ", 2);
        if (!Objects.equals(values[0], "Bearer")) {
            throw new Exception("expect token has Bearer prefix");
        }
        if (values[1].isEmpty()) {
            throw new Exception("expect token has value");
        }

        MockEndpoint.assertIsSatisfied(context);
    }

    @Override
    protected RouteBuilder createRouteBuilder() throws Exception {
        String authUri = "auth://userCredential?googleApiKey=AIzaSyAlE26SQ0OMGjmr4IiF9D6CJkK0eRvV6HA&authServiceAddress=localhost&authServicePort=5050&tenantId=withus-managara-base-0wf23&username=schedule_job%2Busermgmt@manabie.com&password=Manabie123";
        URI uri = new URI(authUri);
        String asciiUrl = uri.toASCIIString();
        String plusReplaced = asciiUrl.replace("+", "%2B");
        return new RouteBuilder() {

            public void configure() {
                from("direct:test")
                        .to(plusReplaced)
                        .to("mock:result");
            }
        };
    }
}
