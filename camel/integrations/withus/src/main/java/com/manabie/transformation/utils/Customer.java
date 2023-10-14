package com.manabie.transformation.utils;

import java.io.UnsupportedEncodingException;

import org.apache.camel.Converter;
import org.apache.camel.Exchange;

public class Customer {
    String data;
    String id;

    public String getData() {
        return data;
    }

    public void setData(String data) {
        this.data = data;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }
}
