package com.manabie.transformation;

import org.apache.camel.builder.RouteBuilder;

import com.manabie.transformation.utils.CustomerParser;

public class Bean extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        getContext().getRegistry().bind("mybean", CustomerParser.class);

        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .setBody().constant(1)
                .to("log:info")
                .bean("mybean", "convert")
                .to("log:info")
                .marshal().json()
                .to("log:info")
                .unmarshal().json()
                .to("log:info");

    }

}
