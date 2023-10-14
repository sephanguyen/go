package com.manabie.transformation;

import org.apache.camel.builder.RouteBuilder;

import com.manabie.transformation.utils.ArrayListAggregationStrategy;
import com.manabie.transformation.utils.CSVCustomer;

import java.text.SimpleDateFormat;
import org.apache.camel.model.dataformat.BindyType;

import java.util.Date;

public class CSVMapper extends RouteBuilder {
    private String ManagaraBase = "ManagaraBase";

    private String getGGStorageEndpoint(String bucketName) {
        String ggEndpointTemplate = "google-storage://{{withusBucket}}?serviceAccountKey={{withusBucketKey}}&objectName=%s&autoCreateBucket=false&operation=getObject";
        return String.format(ggEndpointTemplate, bucketName);
    }

    private String getFileName(String org) {
        Date today = new Date();
        String strToday = new SimpleDateFormat("yyyyMMdd").format(today);
        if (org == ManagaraBase) {
            return String.format("/withus/W2-D6L_users%s.tsv", strToday);
        }
        return String.format("/itee/N1-M1_users%s.tsv", strToday);
    }

    @Override
    public void configure() throws Exception {
        String ggStorageEndpointManagaraBase = getGGStorageEndpoint(getFileName(this.ManagaraBase));

        from("timer://trigger-get-data-withus?fixedRate=true&period=6000000")
                .routeId("TimerRoute")
                .to("log:info")
                .to("direct:downloadFile");

        from("direct:downloadFile")
                .routeId("RunManagaraBase")
                .to(ggStorageEndpointManagaraBase)
                .id("googleStorageEndpoint")
                .to("direct:reaFile");

        from("direct:reaFile")
                .log("file raw: ${body}")
                .convertBodyTo(String.class, "Shift_JIS") // convert the incoming message to Shift_JIS
                .routeId("ReadManagaraBase")
                .convertBodyTo(String.class, "UTF-8")
                .to("log:info")
                .unmarshal().bindy(BindyType.Csv, CSVCustomer.class)
                .to("log:info")
                .split(body())
                .log("log:info test ${body.customerNumber}")
                .aggregate(new ArrayListAggregationStrategy()).constant(true)
                .completionTimeout(500L)
                .marshal().bindy(BindyType.Csv, CSVCustomer.class)
                .to("log:info")
                .to("file://withus.csv");
    }
}
