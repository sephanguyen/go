package com.manabie.concurrency;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.builder.ThreadPoolBuilder;
import org.apache.camel.builder.ThreadPoolProfileBuilder;
import org.apache.camel.model.dataformat.BindyType;
import org.apache.camel.spi.ThreadPoolProfile;

public class SimpleConcurrency extends RouteBuilder {
    // Took: 8s752ms
    // Took: 2s159ms
    // Took: 2s83ms
    @Override
    public void configure() throws Exception {
        // use for small task grows and shrinks on demand
        // no bound no maximum size
        // no task queue
        ExecutorService threadPool = Executors.newCachedThreadPool();
        // fixed thread pool
        ExecutorService fixedThreadPool = Executors.newFixedThreadPool(2);
        // thread pool profile
        ThreadPoolProfile custom = new ThreadPoolProfileBuilder("bigPool")
                .maxPoolSize(200).build();
        // custom thread pool
        ThreadPoolBuilder builder = new ThreadPoolBuilder(getContext());
        ExecutorService myPool = builder.poolSize(5).maxPoolSize(25)
                .maxQueueSize(200).build("myPool");

        getContext().getExecutorServiceManager().registerThreadPoolProfile(custom);

        from("file:target?fileName=bigfile.csv&noop=true")
                .to("log:debug")
                .unmarshal().bindy(BindyType.Csv, CustomerCSV.class)
                .to("log:info")

                // .split(body())
                // .to("direct:update")
                // .end();

                // .split(body()).parallelProcessing()
                // .to("direct:update")
                // .end();

                .split(body())
                .to("seda:update")
                .end();
        // // .end().log("Done update customer ${body}");

        // .split(body()).streaming().executorService(threadPool)
        // .bean(CustomerService.class, "updateCustomer")
        // .end();

        // .split(body()).streaming().executorService(fixedThreadPool)
        // .bean(CustomerService.class, "updateCustomer")
        // .end();

        // .split(body()).streaming().executorService("bigPool")
        // .bean(CustomerService.class, "updateCustomer")
        // .end()
        // .log("info: done");

        // .split(body()).streaming().executorService(myPool)
        // .bean(CustomerService.class, "updateCustomer")
        // .end()
        // .log("info: done");

        // from("direct:update")
        // .bean(CustomerService.class, "updateCustomer").log("running");

        from("seda:update?concurrentConsumers=20")
                .bean(CustomerService.class, "updateCustomer")
                .log("Done update customer ${body}");

        // .aggregate(new FileAggStrategy()).constant(true)
        // .completionTimeout(500L)
        // .to("seda:concurrent");

        // from("seda:concurrent")
        // .to("log:info");
    }
}
