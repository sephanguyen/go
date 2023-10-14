package com.manabie.transformation;

import org.apache.camel.builder.RouteBuilder;

public class Template extends RouteBuilder {
    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .setBody().constant(1)
                .setHeader("firstName", constant("Christian"))
                .setHeader("lastName", constant("Christian"))
                .setHeader("item", constant("Christian"))
                .setBody(constant("Nice to see"))
                .to("velocity:classpath:letter.vm")
                .to("log:info");
    }
}