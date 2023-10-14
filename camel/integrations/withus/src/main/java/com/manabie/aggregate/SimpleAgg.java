package com.manabie.aggregate;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.model.dataformat.BindyType;

import com.manabie.concurrency.CustomerCSV;
import com.manabie.concurrency.FileAggStrategy;

public class SimpleAgg extends RouteBuilder {
    @Override
    public void configure() throws Exception {

        from("file:target?fileName=bigfile.csv&noop=true")
                .to("log:debug")
                .unmarshal().bindy(BindyType.Csv, CustomerCSV.class)
                .split(body(), new FileAggStrategy())
                .to("updateCustomer")
                .to("log:info")
                .end()
                .log("Done update customer ${body}");

    }
}
