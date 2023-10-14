package com.manabie;

import org.apache.camel.builder.RouteBuilder;

public class TriggerTimer extends RouteBuilder {
    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .to("vm:projects");
    }
}
