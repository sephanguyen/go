package com.manabie;

import org.apache.camel.builder.RouteBuilder;

import com.manabie.libs.UnleashContentBaseRoute;

public class UnleashExampleContentRoute extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        getContext().getRegistry().bind("unleashBean", UnleashContentBaseRoute.class);
        from("timer://trigger-get-data-withus?fixedRate=true&period=6000")
                .choice().when()
                .method("unleashBean", "isEnabled('Architecture_BACKEND_MasterData_Course_TeachingMetho')")
                .to("log: flag is enabled")
                .otherwise()
                .to("log: flag is not enabled")
                .end();
    }

}
