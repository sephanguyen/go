package com.manabie.libs;

import java.io.FileInputStream;
import java.io.IOException;
import java.util.Properties;

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.component.mock.MockEndpoint;
import org.apache.camel.test.junit4.CamelTestSupport;
import org.junit.Test;

public class UnleashComponentProducerIT extends CamelTestSupport {

    @Override
    protected Properties useOverridePropertiesWithPropertiesComponent() {
        Properties properties = new Properties();
        try {
            properties.load(new FileInputStream(
                    "./src/test/resources/test.properties"));

            System.out.println("properties" + properties);
        } catch (IOException e) {
            fail(e.getMessage());
        }

        return properties;
    }

    @Test
    public void testProducerUnleash() throws Exception {
        MockEndpoint mockResultTrue = getMockEndpoint("mock:resultTrue");
        mockResultTrue.expectedMessageCount(1);
        MockEndpoint mockResultFalse = getMockEndpoint("mock:resultFalse");
        mockResultFalse.expectedMessageCount(0);

        context.createProducerTemplate().sendBody("direct:test", "Hello World");

        MockEndpoint.assertIsSatisfied(context);

    }

    @Override
    protected RouteBuilder createRouteBuilder() throws Exception {
        return new RouteBuilder() {

            public void configure() {
                from("direct:test")
                        .to("unleash://Architecture_BACKEND_MasterData_Course_TeachingMethod?env={{unleash.env}}&org={{unleash.org}}&url={{unleash.url}}&apiToken={{unleash.token}}&serviceName={{unleash.service}}")
                        .choice()
                        .when(simple("${header.UNLEASH_ENABLED} == true"))
                        .to("mock:resultTrue")
                        .when(simple("${header.UNLEASH_ENABLED} == false"))
                        .to("mock:resultFalse");
            }
        };
    }

}
