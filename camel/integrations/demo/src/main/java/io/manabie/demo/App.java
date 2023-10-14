package io.manabie.demo;

import org.apache.camel.LoggingLevel;
import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.main.Main;

import io.manabie.demo.helloworld.HelloWorld;

public class App extends RouteBuilder {
    public static void main(String[] args) throws Exception {
        Main main = new Main();
        // main.configure().addRoutesBuilder(App.class);
        main.configure().addRoutesBuilder(Withus.class);

        main.run(args);
    }

    @Override
    public void configure() throws Exception {
        HelloWorld.Say();

        from("timer://trigger?fixedRate=true&period=1000")
                .routeId("Timer")
                .log(LoggingLevel.INFO, HelloWorld.Get());
    }
}
