package io.manabie.quarkus.route2;

import static org.apache.camel.builder.endpoint.StaticEndpointBuilders.timer;

import org.apache.camel.builder.RouteBuilder;

public class TimerRoute extends RouteBuilder {
    @Override
    public void configure() throws Exception {
        from(timer("foo").period(1500))
                .log("Hello World 2");
    }
}