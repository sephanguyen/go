package com.manabie;

import java.text.SimpleDateFormat;
import java.util.Date;

import org.apache.camel.PropertyInject;

// camel-k: dependency=camel:csv
// camel-k: dependency=camel:google-storage
// camel-k: property=withusBucket=staging-etl
// camel-k: trait=telemetry.enabled=true
// camel-k: trait=telemetry.sampler=on
// camel-k: trait=telemetry.service-name=WithusStaging.java
// camel-k: trait=telemetry.endpoint=http://opentelemetry-collector.monitoring.svc:4317/v1/traces
// camel-k: service-account=camel-k-demo-runner

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.dataformat.csv.CsvDataFormat;

public class WithusStaging extends RouteBuilder {
    @Override
    public void configure() throws Exception {

        from("timer://trigger-get-data-withus?fixedRate=true&period=5000")
                .routeId("TimerRoute")
                .to("direct:downloadFile");

        from("direct:downloadFile")
                .routeId("RunManagaraBase")
                .to(getGGStorageEndpoint())
                .id("googleStorageEndpoint")
                .to("direct:readFile");

        from("direct:readFile")
                .routeId("ReadManagaraBase")
                .log("raw data: ${body}")
                .convertBodyTo(String.class, "Shift_JIS")
                .convertBodyTo(String.class, "UTF-8")
                .unmarshal(getCSVDataFormat())
                .log("UTF-8 data: ${body}");
    }

    private CsvDataFormat getCSVDataFormat() {
        CsvDataFormat csvDataFormat = new CsvDataFormat();
        char c = '\t';
        csvDataFormat.setDelimiter(c);
        csvDataFormat.setUseMaps(true);
        return csvDataFormat;
    }

    @PropertyInject("withusBucket")
    private String withusBucket;

    private String ManagaraBase = "ManagaraBase";

    private String getFileName(String org) {
        Date today = new Date();
        String strToday = new SimpleDateFormat("yyyyMMdd").format(today);
        if (org == ManagaraBase) {
            return String.format("/withus/W2-D6L_users%s.tsv", strToday);
        }
        return String.format("/itee/N1-M1_users%s.tsv", strToday);
    }

    private String getGGStorageEndpoint() {
        String tpl = "google-storage://%s?objectName=%s&autoCreateBucket=false&operation=getObject";
        return String.format(tpl, this.withusBucket, getFileName(this.ManagaraBase));
    }
}
