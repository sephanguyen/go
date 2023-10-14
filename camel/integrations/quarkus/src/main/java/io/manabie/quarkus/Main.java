package io.manabie.quarkus;

import javax.enterprise.context.ApplicationScoped;
import javax.inject.Inject;

import org.apache.camel.builder.RouteBuilder;

@ApplicationScoped
public class Main extends RouteBuilder {
    @Inject
    MainConfig c;

    @Override
    public void configure() throws Exception {
        log.info("Running with route name: " + this.c.Global().RouteName());

        getContext().addRoutes(lookup(this.c.Global().RouteName()));
    }

    private RouteBuilder lookup(String routeName) throws Exception {
        switch (routeName) {
            case "TimerRoute":
                return new io.manabie.quarkus.route1.TimerRoute(this.c.Route1Config());
            case "TimerRoute2":
                return new io.manabie.quarkus.route2.TimerRoute();
            case "Withus":
                return new io.manabie.quarkus.withus.Route(c.Global(), c.Usermgmt());
            default:
                throw new Exception("invalid route name: " + routeName);
        }
    }
}
