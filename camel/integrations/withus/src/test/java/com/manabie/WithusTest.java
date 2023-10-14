package com.manabie;

import org.apache.camel.RoutesBuilder;
import org.apache.camel.builder.AdviceWith;
import org.apache.camel.builder.AdviceWithRouteBuilder;
import org.apache.camel.component.mock.MockEndpoint;
import org.apache.camel.test.junit5.CamelTestSupport;
import org.junit.jupiter.api.Test;

import java.io.File;
import java.io.FileInputStream;
import java.io.InputStream;
import java.util.Properties;

public class WithusTest extends CamelTestSupport {

    @Override
    protected Properties useOverridePropertiesWithPropertiesComponent() {
        Properties properties = new Properties();
        properties.put("withusBucket", "local-etl");
        properties.put("withusBucketKey", "service_credential.json");
        properties.put("importManagraStudentEnd", "http://usermgmt/import");
        return properties;
    }

    @Test
    public void testMock() throws Exception {
        AdviceWith.adviceWith("TimerRoute", context, new AdviceWithRouteBuilder() {
            @Override
            public void configure() throws Exception {
                replaceFromWith("direct:start");
            }
        });

        // AdviceWith.adviceWith("ReadManagaraBase", context, new
        // AdviceWithRouteBuilder() {
        // @Override
        // public void configure() throws Exception {
        // weaveById("ImportStudent")
        // .replace().to("mock:http");
        // }
        // });

        AdviceWith.adviceWith("RunManagaraBase", context, new AdviceWithRouteBuilder() {
            @Override
            public void configure() throws Exception {
                File file = new File("src/test/resources/withus-integration-data.tsv");
                InputStream targetStream = new FileInputStream(file);

                weaveById("googleStorageEndpoint")
                        .replace()
                        .process(exchange -> {
                            exchange.getIn().setBody(targetStream);
                        }).to("mock:google-storage");
            }
        });

        context.start();

        MockEndpoint mockEndpointGGStorage = getMockEndpoint("mock:google-storage");
        // MockEndpoint mockEndPointInsertStudent = getMockEndpoint("mock:http");

        mockEndpointGGStorage.expectedMessageCount(1);
        // mockEndPointInsertStudent.expectedMessageCount(6);

        template.sendBody("direct:start", "start running");

        MockEndpoint.assertIsSatisfied(context);
        context.stop();
    }

    @Override
    protected RoutesBuilder createRouteBuilder() {
        return new Withus();
    }
}
