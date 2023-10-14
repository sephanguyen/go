package com.manabie.transformation;

import org.apache.camel.Exchange;
import org.apache.camel.builder.RouteBuilder;

import com.manabie.transformation.utils.EnrichAggregationStrategy;

public class ContentEnricher extends RouteBuilder {
    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .setBody(constant("request_body"))
                .setHeader(Exchange.HTTP_METHOD, constant(org.apache.camel.component.http.HttpMethods.POST))
                .enrich("http://localhost:8080/import-student", new EnrichAggregationStrategy())
                .to("log:info")
                .to("log:info  done");
    }
}