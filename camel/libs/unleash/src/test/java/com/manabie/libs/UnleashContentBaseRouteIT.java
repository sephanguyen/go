package com.manabie.libs;

import java.io.FileInputStream;
import java.io.IOException;
import java.util.Properties;

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.component.mock.MockEndpoint;
import org.apache.camel.language.bean.Bean;
import org.apache.camel.test.junit4.CamelTestSupport;
import org.junit.Test;

public class UnleashContentBaseRouteIT extends CamelTestSupport {

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
                getContext().getRegistry().bind("unleash", UnleashContentBaseRoute.class);
                from("direct:test")
                        .choice()
                        .when()
                        .method("unleash",
                                "isEnabled('Architecture_BACKEND_MasterData_Course_TeachingMethod')")
                        .to("mock:resultTrue")
                        .otherwise()
                        .to("mock:resultFalse");
            }
        };
    }

}
