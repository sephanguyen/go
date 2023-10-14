package com.manabie.concurrency;

import org.apache.camel.dataformat.bindy.annotation.CsvRecord;
import org.apache.camel.dataformat.bindy.annotation.DataField;

@CsvRecord(separator = "\t", crlf = "UNIX", generateHeaderColumns = true, skipFirstLine = true)
public class CustomerCSV {
    @DataField(pos = 1)
    public String id;
    @DataField(pos = 2)
    public String data;
    @DataField(pos = 3)
    public String info1;
    @DataField(pos = 4)
    public String info2;

    public String getId() {
        return this.id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getData() {
        return this.data;
    }

    public void setData(String data) {
        this.data = data;
    }

    public String getInfo1() {
        return this.info1;
    }

    public void setInfo1(String info1) {
        this.info1 = info1;
    }

    public String getInfo2() {
        return this.info2;
    }

    public void setInfo2(String info2) {
        this.info2 = info2;
    }
}