package com.manabie.libs;

import org.apache.camel.Converter;

import org.apache.camel.TypeConverters;

@Converter(generateLoader = true)
public class CustomerConverter implements TypeConverters {

    @Converter(allowNull = true)
    public static Customer toCustomer(String data) throws Exception {
        String value = String.valueOf(data);
        Customer customer = new Customer(value + "1", data);
        return customer;
    }
}