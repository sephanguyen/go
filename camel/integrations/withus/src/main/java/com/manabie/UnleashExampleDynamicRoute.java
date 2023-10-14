package com.manabie;

import org.apache.camel.builder.RouteBuilder;

import com.manabie.libs.UnleashDynamicRoute;
import com.sun.org.apache.xalan.internal.xsltc.compiler.Template;

public class UnleashExampleDynamicRoute extends RouteBuilder {

        @Override
        public void configure() throws Exception {
                getContext().getRegistry().bind("unleash", UnleashDynamicRoute.class);
                from("timer://trigger-get-data-withus?fixedRate=true&period=6000")
                                .dynamicRouter(method("unleash",
                                                "isEnabled('Architecture_BACKEND_MasterData_Course_TeachingMethod')"));

                from("direct:unleashEnabledSte1")
                                .to("log:running state 1");

                from("direct:unleashEnabledSte2")
                                .to("log:running state 2");

        }

}
