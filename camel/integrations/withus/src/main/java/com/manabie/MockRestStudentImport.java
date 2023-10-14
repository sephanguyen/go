// camel-k: dependency=camel:rest
// camel-k: dependency=camel:bean
// camel-k: property=file:../../../resources/application.properties

package com.manabie;

import java.util.UUID;

import org.apache.camel.BindToRegistry;
import org.apache.camel.Exchange;
import org.apache.camel.builder.RouteBuilder;

import com.manabie.libs.UnleashContentBaseRoute;

public class MockRestStudentImport extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        // getContext().getRegistry().bind("unleashBean",
        // UnleashContentBaseRoute.class);
        getContext().getRegistry().bind("UUID", UUID.class);

        from("rest:post:import-student")
                .log("info: ${body}")
                // .choice().when()
                // .method("unleashBean",
                // "isEnabled('Architecture_BACKEND_MasterData_Course_TeachingMethod')")
                .to("direct:success");
        // .otherwise()
        // .to("direct:fail");

        from("direct:success").transform().simple("oke");
        from("direct:fail").transform().simple("fail: ${body}")
                .setHeader(Exchange.HTTP_RESPONSE_CODE, constant(500));
    }
}
