package io.manabie.quarkus.route1;

import static org.apache.camel.builder.endpoint.StaticEndpointBuilders.timer;

import javax.enterprise.context.ApplicationScoped;

import org.apache.camel.builder.RouteBuilder;

@ApplicationScoped
public class TimerRoute extends RouteBuilder {
    private String message;

    public TimerRoute(Route1Config c) {
        this.message = c.message();
    }

    public TimerRoute() {
        this.message = "This is the default message";
    }


    @Override
    public void configure() throws Exception {
        from(timer("foo").period("1000")).log(this.message);
    }
}
