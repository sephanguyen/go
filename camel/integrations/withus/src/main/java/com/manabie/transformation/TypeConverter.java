package com.manabie.transformation;

import org.apache.camel.builder.RouteBuilder;
import com.manabie.libs.Customer;

public class TypeConverter extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .setBody().constant("1")
                .to("log:info")
                .convertBodyTo(Customer.class)
                .log("info ${body.id} and ${body.data}")
                .to("log:info")
                .log("info ${body.id} and ${body.data}");
    }
}