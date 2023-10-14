package com.manabie.aggregate;

import org.apache.camel.builder.RouteBuilder;

import com.manabie.concurrency.FileAggStrategy;

public class AggCorrelation extends RouteBuilder {
    @Override
    public void configure() throws Exception {

        from("direct:start")
                .to("log:info")
                .aggregate(header("correlationID"), new FileAggStrategy())
                .completionTimeout(500L)
                .log("info:${body}");

    }
}
